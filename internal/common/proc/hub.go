package proc

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
)

const (
	initState int32 = iota
	runningState
	stoppedState
)

var ErrStopped = errors.New("hub is stopped by signal without errors")

type Hub interface {
	AddProc(proc Proc)
	Wait() error
}

type (
	stopFunc func() error
)

type hub struct {
	state           int32
	wg              sync.WaitGroup
	stopChan        <-chan struct{}
	stoppers        []stopFunc
	startProcChan   chan Proc
	errors          chan error
	reportErrorOnce sync.Once
}

func NewHub(ctx context.Context) Hub {
	startProcChan := make(chan Proc)

	h := &hub{
		startProcChan: startProcChan,
		stopChan:      ctx.Done(),
		state:         initState,
		errors:        make(chan error),
	}

	go func() {
		for proc := range startProcChan {
			h.startProc(proc)
		}
	}()

	return h
}

func (h *hub) AddProc(server Proc) {
	h.startProcChan <- server
}

func (h *hub) Wait() error {
	var err error

	// Wait for error or stopChan message and stop all servers
	h.wg.Add(1)
	go func() {
		select {
		case err = <-h.errors:
			_ = h.stop()
		case <-h.stopChan:
			err = h.stop()
			if err == nil {
				err = ErrStopped
			}
		}
		h.wg.Done()
	}()

	// Wait until all goroutines finished
	h.wg.Wait()
	return err
}

func (h *hub) startProc(proc Proc) {
	started := atomic.CompareAndSwapInt32(&h.state, initState, runningState)

	h.wg.Add(1)
	if !started && atomic.LoadInt32(&h.state) == stoppedState {
		h.wg.Done()
		return
	}

	h.stoppers = append(h.stoppers, proc.Stop)

	go func() {
		err := proc.Start()
		h.reportError(err)
		h.wg.Done()
	}()
}

func (h *hub) stop() error {
	stopped := atomic.CompareAndSwapInt32(&h.state, initState, stoppedState) ||
		atomic.CompareAndSwapInt32(&h.state, runningState, stoppedState)
	if !stopped {
		return nil
	}

	var err error

	for _, stopper := range h.stoppers {
		stopErr := stopper()
		if err == nil && stopErr != nil {
			err = stopErr
		}
	}

	return err
}

func (h *hub) reportError(err error) {
	if atomic.LoadInt32(&h.state) == stoppedState {
		return
	}

	h.reportErrorOnce.Do(func() {
		h.errors <- err
	})
}
