//go:build linux
// +build linux

package tty

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

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
