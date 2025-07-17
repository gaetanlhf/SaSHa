package main

import (
	"github.com/briandowns/spinner"
	"sync"
	"time"
)

type SpinnerManager struct {
	spinner *spinner.Spinner
	active  bool
	mu      sync.Mutex
}

var globalSpinner *SpinnerManager
var once sync.Once

func getSpinnerManager() *SpinnerManager {
	once.Do(func() {
		globalSpinner = &SpinnerManager{}
	})
	return globalSpinner
}

func (sm *SpinnerManager) Start(message string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if !sm.active {
		sm.spinner = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		sm.spinner.Suffix = " " + message
		sm.spinner.Start()
		sm.active = true
	}
}

func (sm *SpinnerManager) UpdateMessage(message string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.active && sm.spinner != nil {
		sm.spinner.Suffix = " " + message
	}
}

func (sm *SpinnerManager) Stop() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.active && sm.spinner != nil {
		sm.spinner.Stop()
		sm.active = false
		sm.spinner = nil
	}
}

func (sm *SpinnerManager) IsActive() bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.active
}

func showSpinner(message string, operation func() error) error {
	sm := getSpinnerManager()

	if sm.IsActive() {
		sm.UpdateMessage(message)
		return operation()
	} else {
		sm.Start(message)
		err := operation()
		sm.Stop()
		return err
	}
}

func startSpinner(message string) {
	sm := getSpinnerManager()
	sm.Start(message)
}

func updateSpinnerMessage(message string) {
	sm := getSpinnerManager()
	sm.UpdateMessage(message)
}

func stopSpinner() {
	sm := getSpinnerManager()
	sm.Stop()
}
