package notify

import "context"

type Notifier interface {
	Notify(ctx context.Context, title, msg string) error
}

type Notify func(ctx context.Context, title, msg string) error

func (n Notify) Notify(ctx context.Context, title, msg string) error {
	return n(ctx, title, msg)
}
