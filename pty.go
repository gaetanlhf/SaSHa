package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/creack/pty"
)

const (
	PASSWORD_PROMPT  = "assword:"
	HOST_AUTH_PROMPT = "The authenticity of host "
)

type PTYManager struct {
	cmd            *exec.Cmd
	ptmx           *os.File
	password       string
	prevMatch      bool
	state1         int
	state2         int
	passwordPrompt string
}

func NewPTYManager(binary string, args []string, password string) *PTYManager {
	cmd := exec.Command(binary, args...)
	return &PTYManager{
		cmd:            cmd,
		password:       password,
		passwordPrompt: PASSWORD_PROMPT,
	}
}

func (p *PTYManager) Start() error {
	var err error
	p.ptmx, err = pty.Start(p.cmd)
	if err != nil {
		return fmt.Errorf("failed to start pty: %w", err)
	}

	return nil
}

func (p *PTYManager) Run() error {
	if err := p.Start(); err != nil {
		return err
	}
	defer p.ptmx.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, p.ptmx); err != nil {
			}
		}
	}()
	ch <- syscall.SIGWINCH
	defer func() { signal.Stop(ch); close(ch) }()

	done := make(chan error, 1)

	go func() {
		done <- p.handleOutput()
	}()

	go func() {
		_, err := io.Copy(p.ptmx, os.Stdin)
		if err != nil {
		}
	}()

	select {
	case err := <-done:
		if err != nil {
			return err
		}
	}

	return p.cmd.Wait()
}

func (p *PTYManager) handleOutput() error {
	buffer := make([]byte, 256)

	for {
		numRead, err := p.ptmx.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if numRead > 0 {
			os.Stdout.Write(buffer[:numRead])

			if p.password != "" {
				ret := p.processBuffer(buffer[:numRead], numRead)
				if ret != 0 {
					switch ret {
					case 1:
						return fmt.Errorf("authentication failed: incorrect password")
					case 2:
						return fmt.Errorf("host key verification failed")
					default:
						return fmt.Errorf("authentication error")
					}
				}
			}
		}
	}

	return nil
}

func (p *PTYManager) processBuffer(buffer []byte, numRead int) int {
	p.state1 = p.match(p.passwordPrompt, buffer, numRead, p.state1)

	if p.state1 >= len(p.passwordPrompt) {
		if !p.prevMatch {
			p.writePassword()
			p.state1 = 0
			p.prevMatch = true
		} else {
			return 1
		}
	}

	p.state2 = p.match(HOST_AUTH_PROMPT, buffer, numRead, p.state2)

	if p.state2 >= len(HOST_AUTH_PROMPT) {
		return 2
	}

	return 0
}

func (p *PTYManager) match(reference string, buffer []byte, bufSize int, state int) int {
	for i := 0; state < len(reference) && i < bufSize; i++ {
		if reference[state] == buffer[i] {
			state++
		} else {
			state = 0
			if state < len(reference) && reference[state] == buffer[i] {
				state++
			}
		}
	}
	return state
}

func (p *PTYManager) writePassword() error {
	time.Sleep(100 * time.Millisecond)

	_, err := p.ptmx.Write([]byte(p.password + "\n"))
	if err != nil {
		return fmt.Errorf("failed to send password: %w", err)
	}

	return nil
}

func (p *PTYManager) Close() error {
	if p.ptmx != nil {
		return p.ptmx.Close()
	}
	return nil
}

func executeSSHWithPassword(binary string, args []string, password string) error {
	ptyManager := NewPTYManager(binary, args, password)

	return ptyManager.Run()
}
