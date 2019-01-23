package strelets

import (
	"context"
	"fmt"
	"time"
)

func Try(parentContext context.Context, tries int, timeoutPerTry time.Duration, retryInterval time.Duration, f func(ctxWithTimeout context.Context) error) (err error) {
	ctx, cancel := context.WithTimeout(parentContext, timeoutPerTry)
	defer cancel()

	retryAfter := retryInterval

	for i := 0; i < tries; i++ {
		err = f(ctx)
		if err == nil {
			if i > 0 {
				fmt.Println(fmt.Sprintf("attempt #%d: success", i))
			}
			return
		}

		fmt.Println(fmt.Sprintf("attempt #%d, retry in %s: %s", i, retryAfter, err))
		time.Sleep(retryAfter)
		retryAfter = retryAfter * 2
	}

	return
}
