package tty

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"
)

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

func NewSignalManager() *SignalManager {
	ctx, cancel := context.WithCancel(context.Background())

	sm := &SignalManager{
		ctx:       ctx,
		cancel:    cancel,
		winchChan: make(chan os.Signal, 1),
		termChan:  make(chan os.Signal, 1),
	}

	signal.Notify(sm.winchChan, syscall.SIGWINCH)
	signal.Notify(sm.termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	return sm
}

func (sm *SignalManager) SetChildPid(pid int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.childPid = pid
}

func (sm *SignalManager) SetTerminalState(state *term.State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.oldState = state
}

func (sm *SignalManager) SetWindowResizeHandler(handler func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onWindowResize = handler
}

func (sm *SignalManager) SetTerminateHandler(handler func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onTerminate = handler
}

func (sm *SignalManager) AddCleanup(fn func()) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.cleanup = append(sm.cleanup, fn)
}

func (sm *SignalManager) Start() {
	sm.wg.Add(2)

	go sm.handleWindowResize()
	go sm.handleTermination()
}

func (sm *SignalManager) Stop() {
	sm.cancel()
	sm.wg.Wait()
	sm.performCleanup()
}

func (sm *SignalManager) Context() context.Context {
	return sm.ctx
}

func (sm *SignalManager) handleWindowResize() {
	defer sm.wg.Done()

	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-sm.winchChan:
			sm.mu.RLock()
			handler := sm.onWindowResize
			sm.mu.RUnlock()

			if handler != nil {
				handler()
			}
		}
	}
}

func (sm *SignalManager) handleTermination() {
	defer sm.wg.Done()

	select {
	case <-sm.ctx.Done():
		return
	case <-sm.termChan:
		sm.mu.RLock()
		handler := sm.onTerminate
		sm.mu.RUnlock()

		if handler != nil {
			handler()
		}

		sm.performCleanup()
		os.Exit(1)
	}
}

func (sm *SignalManager) performCleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.oldState != nil {
		term.Restore(int(os.Stdin.Fd()), sm.oldState)
		sm.oldState = nil
	}

	if sm.childPid != 0 {
		syscall.Kill(sm.childPid, syscall.SIGTERM)
	}

	for _, fn := range sm.cleanup {
		fn()
	}

	signal.Stop(sm.winchChan)
	signal.Stop(sm.termChan)
	close(sm.winchChan)
	close(sm.termChan)
}
