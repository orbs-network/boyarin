package utils

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"time"
)

func Try(parentContext context.Context, tries int, timeoutPerTry time.Duration, retryInterval time.Duration, f func(ctxWithTimeout context.Context) error) (err error) {
	ctx, cancel := context.WithTimeout(parentContext, timeoutPerTry)
	defer cancel()

	retryAfter := retryInterval

	for i := 0; i < tries; i++ {
		select {
		case <-parentContext.Done():
			return fmt.Errorf("%s after %d attempts", parentContext.Err(), i)
		default:
			err = f(ctx)
			if err == nil {
				if i > 0 {
					fmt.Println(fmt.Sprintf("attempt #%d: success", i))
				}
				return
			}
			err = errors.WithStack(err)
			fmt.Println(fmt.Sprintf("attempt #%d, retry in %s: %+v", i, retryAfter, err))
			time.Sleep(retryAfter)
			retryAfter = retryAfter * 2
		}
	}

	return
}
