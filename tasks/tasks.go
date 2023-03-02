package tasks

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/yang-zzhong/xl/notify"
)

type Initializer interface {
	Init() error
}

type Closer interface {
	Close() error
}

type handleItem func(ctx context.Context, item ...interface{}) error

type Task interface {
	Total() (int, error)
	Do(ctx context.Context, page int) error
}

type Dispatcher struct {
	MaxRetryTimes int
	Concurrence   int
	Savepoint     Savepoint
	Task          Task
	PageSize      int
	Notifiers     []notify.Notifier
	Logger        Logger
	once          sync.Once
	optionError   error
}

func errWrap(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

func isExists(name string) (bool, error) {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (f *Dispatcher) setDefaultOptions() error {
	f.once.Do(func() {
		if f.PageSize == 0 {
			f.PageSize = 500
		}
		if f.Concurrence == 0 {
			f.Concurrence = 10
		}
		if f.Logger == nil {
			f.Logger = StdLogger(os.Stdout)
		}
		if f.MaxRetryTimes == 0 {
			f.MaxRetryTimes = 10000
		}
		if initializer, ok := f.Logger.(Initializer); ok {
			if err := initializer.Init(); err != nil {
				f.optionError = err
				return
			}
		}
		if initializer, ok := f.Savepoint.(Initializer); ok {
			if err := initializer.Init(); err != nil {
				f.optionError = err
				return
			}
		}
		if f.Task == nil {
			f.optionError = errors.New("请提供资源参数。Resources")
			return
		}
	})
	return f.optionError
}

func (f *Dispatcher) Dispatch(ctx context.Context) error {
	if err := f.setDefaultOptions(); err != nil {
		return err
	}
	total, err := f.Task.Total()
	if err != nil {
		return err
	}
	f.Logger.Infof("将会处理[%d]条数据。", total)
	return f.batch(ctx, total, f.Concurrence, f.PageSize, func(page int) error {
		start := time.Now()
		var err error
		for i := 0; i < f.MaxRetryTimes; i++ {
			if err = f.Task.Do(ctx, page); err != nil {
				end := time.Now()
				f.Logger.Errorf("处理第%d页资源出错。耗时 %dS: %s", page, end.Sub(start)/time.Second, err.Error())
				sec := (i + 1) * 2
				f.Logger.Infof("处理第%d页资源出错。[%d] 将在[%d]秒后重试...", page, i, sec)
				time.Sleep(time.Second * time.Duration(sec))
				if i != 0 && i%10 == 0 {
					for _, notifier := range f.Notifiers {
						notifier.Notify(ctx, "有任务阻塞，请即时处理", fmt.Sprintf("处理第%d页资源出错。[%d] 将在[%d]秒后重试...", page, i, sec))
					}
				}
				start = time.Now()
				continue
			}
			end := time.Now()
			f.Logger.Infof("第%d页资源处理完成。耗时: %dS", page, end.Sub(start)/time.Second)
			return nil
		}
		f.Logger.Errorf("重试第%d页资源%d次均未成功。", page, f.MaxRetryTimes)
		return err
	})
}

func (f *Dispatcher) batch(ctx context.Context, total, maxc, pageSize int, handle func(page int) error) error {
	reqs := (total / pageSize) + 1
	var start int
	if f.Savepoint != nil {
		var s int
		if err := f.Savepoint.Offset(ctx, &s); err != nil {
			f.Logger.Infof("没有获取到savepoint。从0页开始处理")
		} else if s == -1 {
			f.Logger.Infof("该数据之前已经处理完成，无需重复处理")
			return nil
		}
		start = s
	}
	f.Logger.Infof("从第%d页开始，总共%d页, %d条数据", start, reqs, total)
	var offset int
	for i := start; i < reqs; i += maxc {
		offset = i
		end := i + maxc
		if end > reqs {
			end = reqs
		}
		var err error
		func() {
			if f.Savepoint != nil {
				defer func() {
					if err := f.Savepoint.SetOffset(ctx, offset); err != nil {
						f.Logger.Errorf("设置保存点失败: %s", err.Error())
					}
				}()
			}
			var wg sync.WaitGroup
			var lock sync.Mutex
			for j := i; j < end; j++ {
				wg.Add(1)
				go func(page int) {
					defer wg.Done()
					if e := handle(page); e != nil {
						lock.Lock()
						err = e
						lock.Unlock()
					}
				}(j)
			}
			wg.Wait()
		}()
		if err != nil {
			return err
		}
	}
	if f.Savepoint != nil {
		if err := f.Savepoint.SetOffset(ctx, -1); err != nil {
			f.Logger.Errorf("设置保存点失败: %s", err.Error())
		}
	}
	f.Logger.Infof("处理完成。共%d页数据被处理", total)
	return nil
}
