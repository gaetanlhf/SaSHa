package tty

import (
	"context"
	"golang.org/x/term"
	"os"
	"regexp"
	"sync"
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

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

type PromptAction int

type PromptMatcher struct {
	patterns []PromptPattern
}

type PromptPattern struct {
	name     string
	patterns []string
	regexes  []*regexp.Regexp
	action   PromptAction
}

type SignalManager struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex

	childPid int
	oldState *term.State

	winchChan chan os.Signal
	termChan  chan os.Signal

	onWindowResize func()
	onTerminate    func()

	cleanup []func()
}
