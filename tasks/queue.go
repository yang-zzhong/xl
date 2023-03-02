package tasks

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Do func(ctx context.Context, data any) error

func (d Do) Do(ctx context.Context, data any) error {
	return d(ctx, data)
}

type Doer interface {
	Do(ctx context.Context, data any) error
}

type Queue interface {
	Add(ctx context.Context, r ...any) error
}

type Starter interface {
	Start(ctx context.Context) error
}

type Stopper interface {
	Stop() error
}

type queueWorker struct {
	doer Doer
	busy bool
	lock sync.RWMutex
}

type queue struct {
	bufSize int
	workers []*queueWorker
	buffer  []any
	offset  int
	addlock sync.Mutex
	buflock sync.RWMutex
	ch      chan struct{}
	quit    chan struct{}
}

func QueueWorkers(doer Doer, n int) []Doer {
	ret := make([]Doer, n)
	for i := 0; i < n; i++ {
		ret[i] = doer
	}
	return ret
}

func NewQueue(inserter []Doer, bufSize int) *queue {
	if bufSize == 0 {
		bufSize = 1
	}
	return &queue{
		workers: func() []*queueWorker {
			ret := make([]*queueWorker, len(inserter))
			for i := 0; i < len(inserter); i++ {
				ret[i] = &queueWorker{doer: inserter[i]}
			}
			return ret
		}(),
		buffer:  make([]interface{}, bufSize),
		ch:      make(chan struct{}, bufSize),
		bufSize: bufSize,
		quit:    make(chan struct{}),
	}
}

func (in *queue) Add(ctx context.Context, r ...interface{}) error {
	in.addlock.Lock()
	defer in.addlock.Unlock()
	var err error
	var start, add int
	for start = 0; start < len(r); start += add {
		add, err = in.add(ctx, r[start:])
		if err != nil {
			return err
		}
	}
	return nil
}

func (in *queue) add(ctx context.Context, r []interface{}) (int, error) {
	for in.getOffset() >= in.bufSize {
	}
	in.buflock.Lock()
	offset := in.getOffset()
	var i int
	for i = 0; offset < in.bufSize && i < len(r); i++ {
		in.buffer[offset] = r[i]
		in.setOffset(offset + 1)
		offset++
	}
	in.buflock.Unlock()
	in.ch <- struct{}{}
	return i, nil
}

func (in *queue) Stop() error {
	in.quit <- struct{}{}
	return nil
}

func (in *queue) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	for {
		select {
		case <-in.quit:
			wg.Wait()
			return nil
		case <-in.ch:
			in.buflock.Lock()
			offset := in.getOffset()
			if offset == 0 {
				in.buflock.Unlock()
				continue
			}
			data := make([]interface{}, offset)
			copy(data, in.buffer[:offset])
			in.setOffset(0)
			in.buflock.Unlock()
			worker := in.getWorker()
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer func() {
					worker.lock.Lock()
					worker.busy = false
					worker.lock.Unlock()
				}()
				worker.lock.Lock()
				worker.busy = true
				worker.lock.Unlock()
				for i := 0; i < 3; i++ {
					if err := worker.doer.Do(ctx, data); err != nil {
						fmt.Printf("%s", err.Error())
						continue
					}
					break
				}
			}()
		}
	}
}

func (in *queue) getWorker() *queueWorker {
	for {
		for _, worker := range in.workers {
			worker.lock.RLock()
			if !worker.busy {
				worker.lock.RUnlock()
				return worker
			}
			worker.lock.RUnlock()
		}
		time.Sleep(time.Millisecond)
	}
}

func (in *queue) setOffset(offset int) {
	in.offset = offset
}

func (in *queue) getOffset() int {
	return in.offset
}
