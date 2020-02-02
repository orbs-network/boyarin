package supervized

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestGoForever(t *testing.T) {
	t.Skip("manual test")

	count := 0
	maxIterations := 5
	numOfIterationsBeforeCrash := 2

	GoForever(func(_ bool) {
		if count > maxIterations {
			return
		} else if count > numOfIterationsBeforeCrash {
			panic("foo")
		} else {
			count++
			fmt.Println(count)
		}
	})

	time.Sleep(1 * time.Millisecond)
	require.EqualValues(t, 3, count)
}
