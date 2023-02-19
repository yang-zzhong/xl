package utils

import (
	"errors"
	"math"
	"sync"
)

var (
	ErrAcquireTimeout         = errors.New("acquire goroutine timeout")
	DefaultConcurrenceTotal   = 30000
	DefaultConcurrenceManager ConcurrenceManager
)

func init() {
	DefaultConcurrenceManager = ConcurrenceManager{
		Total: 30000,
		ch:    make(chan struct{}),
	}
}

func MaxC(maxC uint, tasks uint, worker func(start, end int)) {
	DefaultConcurrenceManager.MaxC(maxC, tasks, worker)
}

func Go(handle func()) error {
	return DefaultConcurrenceManager.Go(handle)
}

func SetTotal(total int) {
	DefaultConcurrenceManager.SetTotal(total)
}

// ConcurrenceManager manager the concurrence acquire, ensure the reasonable number in whole process
type ConcurrenceManager struct {
	Total    int
	Acquired int
	lock     sync.RWMutex
	ch       chan struct{}
}

// MaxC call worker with max concurrence
func (m *ConcurrenceManager) MaxC(maxC uint, tasks uint, worker func(start, end int)) {
	block := int(math.Ceil(float64(tasks) / float64(maxC)))
	var wg sync.WaitGroup
	for i := 0; i < int(tasks); i += block {
		wg.Add(1)
		func(i int) {
			m.Go(func() {
				end := i + block
				if end > int(tasks) {
					end = int(tasks)
				}
				worker(i, end)
				wg.Done()
			})
		}(i)
	}
	wg.Wait()
}

func (m *ConcurrenceManager) Go(handle func()) error {
	if err := m.acquire(1); err != nil {
		return err
	}
	go func() {
		handle()
		m.lock.Lock()
		m.Acquired -= 1
		m.lock.Unlock()
		m.ch <- struct{}{}
	}()
	return nil
}

func (m *ConcurrenceManager) SetTotal(total int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Total = total
}

func (m *ConcurrenceManager) acquire(acquired int) error {
	timeout := make(chan struct{})
	m.lock.RLock()
	rest := m.Total - m.Acquired
	m.lock.RUnlock()
	if rest <= 0 {
		select {
		case <-m.ch:
			return m.acquire(acquired)
		case <-timeout:
			return ErrAcquireTimeout
		}
	}
	m.lock.Lock()
	m.Acquired += acquired
	m.lock.Unlock()
	return nil
}
