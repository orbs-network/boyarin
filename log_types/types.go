package log_types

import "github.com/orbs-network/scribe/log"

func VirtualChainId(value int64) *log.Field {
	return log.Int64("vcid", value)
}
