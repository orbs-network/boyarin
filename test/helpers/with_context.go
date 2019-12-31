package helpers

import (
	"context"
	"github.com/orbs-network/govnr"
)

func WithContext(f func(ctx context.Context)) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	f(ctx)
}

func WithContextAndShutdown(f func(ctx context.Context) govnr.ShutdownWaiter) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	waiter := f(ctx)
	ch := make(chan struct{})
	go func() {
		waiter.WaitUntilShutdown(context.Background())
		close(ch)
	}()
	cancel()
	select {
	case <-ch:
	}
}
