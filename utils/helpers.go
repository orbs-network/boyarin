package utils

import (
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"strings"
)

type logErrors struct {
	logger log.Logger
}

func (le *logErrors) Error(err error) {
	le.logger.Info("error in service status reporter", log.Error(err))
}

func NewLogErrors(logger log.Logger) govnr.Errorer {
	return &logErrors{logger: logger}
}

func AggregateErrors(errors []error) error {
	if errors == nil {
		return nil
	}

	var lines []string

	for _, err := range errors {
		lines = append(lines, err.Error())
	}

	return fmt.Errorf(strings.Join(lines, ", "))
}
