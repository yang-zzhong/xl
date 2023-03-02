package tasks

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
)

func baseDispatcher(t *testing.T, ctrl *gomock.Controller) *Dispatcher {
	savepoint := NewMockSavepoint(ctrl)
	savepoint.EXPECT().SetOffset(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	savepoint.EXPECT().Offset(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	logger := NewMockLogger(ctrl)
	logger.EXPECT().Infof(gomock.Any(), gomock.Any()).DoAndReturn(func(format string, args ...interface{}) {
		t.Logf(format, args...)
	}).AnyTimes()
	logger.EXPECT().Errorf(gomock.Any(), gomock.Any()).DoAndReturn(func(format string, args ...interface{}) {
		t.Logf(format, args...)
	}).AnyTimes()
	return &Dispatcher{
		Savepoint: savepoint,
		Logger:    logger,
	}
}

func TestDispatcher_batch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	f := baseDispatcher(t, ctrl)
	var totalPage = 10000000 / 1000
	var lock sync.Mutex
	f.batch(context.Background(), 10000000, 20, 101, func(page int) error {
		lock.Lock()
		totalPage -= 1
		lock.Unlock()
		t.Logf("page: %d\n", page)
		return nil
	})
	if totalPage > 0 {
		t.Fatal("batch error")
	}
}

func TestDispatcher_Fetch(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	res := NewMockTask(ctrl)
	res.EXPECT().Total().Return(10000, nil).AnyTimes()
	res.EXPECT().Do(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, page int) error {
		time.Sleep(time.Millisecond * 20)
		if rand.Intn(10)%4 == 0 {
			return errors.New("random error")
		}
		return nil
	}).AnyTimes()
	f := baseDispatcher(t, ctrl)
	f.Task = res
	if err := f.Dispatch(context.Background()); err != nil {
		t.Fatal(err)
	}
}
