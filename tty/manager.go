package tty

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/term"
)

const (
	BufferSize       = 4096
	OutputBufferSize = 256
	PTYPollInterval  = 10 * time.Millisecond
	PasswordDelay    = 100 * time.Millisecond
)

type Manager struct {
	masterFd  int
	slaveFd   int
	slaveName string
	password  string

	signalManager *SignalManager
	promptMatcher *PromptMatcher

	childPid     int
	passwordSent bool

	mu sync.RWMutex
}

func NewManager(password string) *Manager {
	return &Manager{
		password:      password,
		signalManager: NewSignalManager(),
		promptMatcher: NewPromptMatcher(),
	}
}

func (m *Manager) openPTY() error {
	master, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		return fmt.Errorf("open master: %w", err)
	}
	m.masterFd = master

	zero := 0
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(m.masterFd), syscall.TIOCSPTLCK, uintptr(unsafe.Pointer(&zero))); errno != 0 {
		syscall.Close(m.masterFd)
		return fmt.Errorf("unlock PTY: %w", errno)
	}

	var ptsNum int
	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(m.masterFd), syscall.TIOCGPTN, uintptr(unsafe.Pointer(&ptsNum))); errno != 0 {
		syscall.Close(m.masterFd)
		return fmt.Errorf("get PTS number: %w", errno)
	}

	m.slaveName = fmt.Sprintf("/dev/pts/%d", ptsNum)

	slave, err := syscall.Open(m.slaveName, syscall.O_RDWR|syscall.O_NOCTTY, 0)
	if err != nil {
		syscall.Close(m.masterFd)
		return fmt.Errorf("open slave: %w", err)
	}
	m.slaveFd = slave

	if err := syscall.SetNonblock(m.masterFd, true); err != nil {
		syscall.Close(m.masterFd)
		syscall.Close(m.slaveFd)
		return fmt.Errorf("set non-blocking: %w", err)
	}

	return nil
}

func (m *Manager) setupTerminal() (*term.State, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, nil
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, fmt.Errorf("set raw mode: %w", err)
	}

	m.signalManager.SetTerminalState(oldState)
	return oldState, nil
}

func (m *Manager) Run(binary string, args []string) error {
	if err := m.openPTY(); err != nil {
		return err
	}

	m.signalManager.AddCleanup(func() {
		syscall.Close(m.masterFd)
		if m.slaveFd != 0 {
			syscall.Close(m.slaveFd)
		}
	})

	oldState, err := m.setupTerminal()
	if err != nil {
		return err
	}

	if err := m.setTerminalSize(); err != nil {
	}

	m.signalManager.SetWindowResizeHandler(func() {
		m.setTerminalSize()
	})

	m.signalManager.SetTerminateHandler(func() {
		if oldState != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
		if m.childPid != 0 {
			syscall.Kill(m.childPid, syscall.SIGTERM)
		}
	})

	m.signalManager.Start()
	defer m.signalManager.Stop()

	childPid, err := syscall.ForkExec(binary, append([]string{binary}, args...), &syscall.ProcAttr{
		Env:   os.Environ(),
		Files: []uintptr{uintptr(m.slaveFd), uintptr(m.slaveFd), uintptr(m.slaveFd)},
		Dir:   "/",
		Sys: &syscall.SysProcAttr{
			Setsid:  true,
			Setctty: true,
			Ctty:    0,
		},
	})

	if err != nil {
		return fmt.Errorf("fork/exec: %w", err)
	}

	m.childPid = childPid
	m.signalManager.SetChildPid(childPid)

	ctx := m.signalManager.Context()
	var wg sync.WaitGroup

	errChan := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := m.handleOutput(ctx); err != nil && err != io.EOF {
			errChan <- err
		}
	}()

	go func() {
		defer wg.Done()
		if err := m.handleInput(ctx); err != nil && err != io.EOF {
			errChan <- err
		}
	}()

	childDone := make(chan int)
	go func() {
		exitCode := m.waitForChild()
		childDone <- exitCode
	}()

	select {
	case err := <-errChan:
		syscall.Kill(m.childPid, syscall.SIGTERM)
		if oldState != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
		return err
	case exitCode := <-childDone:
		if oldState != nil {
			term.Restore(int(os.Stdin.Fd()), oldState)
		}
		os.Exit(exitCode)
	case <-ctx.Done():
		return nil
	}

	wg.Wait()
	return nil
}

func (m *Manager) waitForChild() int {
	var wstatus syscall.WaitStatus
	for {
		wpid, err := syscall.Wait4(m.childPid, &wstatus, syscall.WNOHANG, nil)
		if err != nil || wpid == m.childPid {
			exitCode := 0
			if wstatus.Exited() {
				exitCode = wstatus.ExitStatus()
			} else if wstatus.Signaled() {
				exitCode = 128 + int(wstatus.Signal())
			}
			return exitCode
		}
		time.Sleep(100 * time.Millisecond)
	}
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func (m *Manager) setTerminalSize() error {
	var size winsize

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size))); errno != 0 {
		return fmt.Errorf("get window size: %w", errno)
	}

	if _, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(m.masterFd), syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&size))); errno != 0 {
		return fmt.Errorf("set window size: %w", errno)
	}

	return nil
}

func (m *Manager) handleInput(ctx context.Context) error {
	buffer := make([]byte, BufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		n, err := os.Stdin.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}
		if n > 0 {
			if _, err := syscall.Write(m.masterFd, buffer[:n]); err != nil {
				return fmt.Errorf("write to PTY: %w", err)
			}
		}
	}
}

func (m *Manager) handleOutput(ctx context.Context) error {
	buffer := make([]byte, OutputBufferSize)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		numRead, err := syscall.Read(m.masterFd, buffer)
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				time.Sleep(PTYPollInterval)
				continue
			}
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read output: %w", err)
		}

		if numRead > 0 {
			if _, err := os.Stdout.Write(buffer[:numRead]); err != nil {
				return fmt.Errorf("write output: %w", err)
			}

			if m.password != "" {
				if err := m.processPrompts(buffer[:numRead]); err != nil {
					return err
				}
			}
		}
	}
}

func (m *Manager) processPrompts(buffer []byte) error {
	action, matched := m.promptMatcher.MatchPrompt(buffer)
	if !matched {
		return nil
	}

	if action == ActionSendPassword {
		m.mu.Lock()
		if m.passwordSent {
			m.mu.Unlock()
			return fmt.Errorf("authentication failed: incorrect password")
		}
		m.passwordSent = true
		m.mu.Unlock()

		return m.writePassword()
	}

	return nil
}

func (m *Manager) writePassword() error {
	time.Sleep(PasswordDelay)

	if _, err := syscall.Write(m.masterFd, []byte(m.password)); err != nil {
		return fmt.Errorf("send password: %w", err)
	}

	if _, err := syscall.Write(m.masterFd, []byte("\n")); err != nil {
		return fmt.Errorf("send newline: %w", err)
	}

	return nil
}
