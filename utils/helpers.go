package utils

import (
	"fmt"
	"github.com/orbs-network/govnr"
	"github.com/orbs-network/scribe/log"
	"strings"
)

type logErrors struct {
	name   string
	logger log.Logger
}

func (le *logErrors) Error(err error) {
	le.logger.Info(fmt.Sprintf("error in %s", le.name), log.Error(err))
}

func NewLogErrors(name string, logger log.Logger) govnr.Errorer {
	return &logErrors{logger: logger, name: name}
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
