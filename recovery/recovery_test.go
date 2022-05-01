package recovery

import (
	"context"
	"errors"
	"testing"

	"github.com/orbs-network/scribe/log"
)

func Test_RecoveryConfigSingleton(t *testing.T) {

	// init recovery config
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.json"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}
	basicLogger := log.GetLogger()
	Init(config, basicLogger)

	recovery1 := GetInstance()

	// get same instance
	recovery2 := GetInstance()
	if recovery1.config.Url != recovery2.config.Url {
		t.Error("config url in two instances is not equal")
	}
	if recovery1.config.IntervalMinute != recovery2.config.IntervalMinute {
		t.Error("config IntervalMinute in two instances is not equal")
	}
}

func Test_Recovery404(t *testing.T) {
	logger = log.GetLogger()
	url := "http://http://www.xosdhjfglk.com/xxx/main.sh"
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)

	res, err := GetInstance().readUrl(url) //, "./boyar_recovery/")
	if err == nil {
		t.Error("404 url did not result an error")
	}
	if res != "" {
		t.Error("404 url returned a result")
	}

	// get same instance

}

func Test_RecoveryJson(t *testing.T) {
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0xTEST/main.json"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)

	r := GetInstance()
	r.tick()

	expect := "identical\nidentical\nidentical\n"
	if r.lastOutput != expect {
		t.Errorf("expect:\n%s got:\n%s", expect, r.lastOutput)
	}

}

func Test_ExecutionTimeout(t *testing.T) {
	// init recovery config
	config := Config{
		IntervalMinute: 1,
		TimeoutSec:     5,
		Url:            "",
	}
	logger = log.GetLogger()
	Init(config, logger)

	r := GetInstance()
	args := []string{"6"}
	err := r.runCommand("sleep", "", "", args)
	if err == nil {
		t.Error("timeout usecase did not return error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error is not timeout: %s", err.Error())
	}
}
