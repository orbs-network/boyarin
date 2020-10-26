package helpers

import (
	"github.com/orbs-network/scribe/log"
	"os"
)

func DefaultTestLogger() log.Logger {
	return log.GetLogger().WithOutput(log.NewFormattingOutput(os.Stdout, log.NewHumanReadableFormatter()))
}
