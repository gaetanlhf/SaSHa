//go:build darwin
// +build darwin

package tty

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func (m *Manager) openPTY() error {
	master, err := syscall.Open("/dev/ptmx", syscall.O_RDWR|syscall.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("open master: %w", err)
	}
	m.masterFd = master

	var slaveBuf [256]byte
	if err := ptsname(m.masterFd, slaveBuf[:]); err != nil {
		syscall.Close(m.masterFd)
		return fmt.Errorf("get slave name: %w", err)
	}

	var slaveNameBytes []byte
	for i, b := range slaveBuf {
		if b == 0 {
			slaveNameBytes = slaveBuf[:i]
			break
		}
	}
	m.slaveName = string(slaveNameBytes)

	if err := grantpt(m.masterFd); err != nil {
		syscall.Close(m.masterFd)
		return fmt.Errorf("grantpt: %w", err)
	}

	if err := unlockpt(m.masterFd); err != nil {
		syscall.Close(m.masterFd)
		return fmt.Errorf("unlockpt: %w", err)
	}

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

func ptsname(fd int, buf []byte) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x40807453, uintptr(unsafe.Pointer(&buf[0])))
	if errno != 0 {
		return errno
	}
	return nil
}

func grantpt(fd int) error {
	return nil
}

func unlockpt(fd int) error {
	return nil
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
