package tasks

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/spf13/cast"
)

func TestQueue(t *testing.T) {
	total := 10000
	should := func() int {
		t := 0
		for i := 0; i < total; i++ {
			t += i
		}
		return t
	}()
	result := 0
	queue := NewQueue(QueueWorkers(Do(func(ctx context.Context, r interface{}) error {
		t.Logf("handle resource: %v", r)
		val := reflect.ValueOf(r)
		switch val.Type().Kind() {
		case reflect.Slice:
			for i := 0; i < val.Len(); i++ {
				result += cast.ToInt(val.Index(i).Interface())
			}
		}
		return nil
	}), 10), 100)
	ctx := context.Background()
	go func() {
		var wg sync.WaitGroup
		for i := 0; i < total; i++ {
			wg.Add(1)
			go func(i int) {
				_ = queue.Add(ctx, i)
				wg.Done()
			}(i)
		}
		wg.Wait()
		_ = queue.Stop()
	}()
	if err := queue.Start(ctx); err != nil {
		t.Fatal(err)
	}
	if should != result {
		t.Fatal("failed")
	}
}
