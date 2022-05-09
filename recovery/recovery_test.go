package recovery

import (
	"testing"
	"time"

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

	Init(config, logger)
	logger = log.GetLogger()

	GetInstance().tick()
	res, err := GetInstance().readUrl(url) //, "./boyar_recovery/")
	if err == nil {
		t.Error("404 url did not result an error")
	}
	if res != "" {
		t.Error("404 url returned a result")
	}
}

func Test_RecoveryJsonHappy(t *testing.T) {
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

func Test_RecoveryEmptyJson(t *testing.T) {
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0xTEST/empty.json"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)

	r := GetInstance()
	r.tick()

	if r.lastError != e_json_no_binary {
		t.Errorf("expect:\n%s got:\n%s", e_json_no_binary, r.lastError)
	}
}

func Test_RecoveryJsonInvalid(t *testing.T) {
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0xTEST/invalid.json"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)

	r := GetInstance()
	r.tick()

	e := "invalid character"
	if r.lastError[:len(e)] != e {
		t.Errorf("expect:\n%s got:\n%s", e, r.lastError)
	}
}

func Test_RecoveryTimeout(t *testing.T) {
	// init recovery config
	config := Config{
		IntervalMinute: 5,
		TimeoutMinute:  1,
		Url:            "",
	}
	logger = log.GetLogger()
	Init(config, logger)

	// happy path
	r := GetInstance()
	t.Logf("sleeping 5 %s", time.Now())
	args := []string{"2"} // 2 seconds = happy path
	err := r.runCommand("sleep", "", "", args)
	if err != nil {
		t.Error(err)
	}
}

func Test_RecoveryStderr(t *testing.T) {
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0xTEST/stderr.json"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)

	r := GetInstance()
	r.tick()

	e := "invalid character"
	if r.lastError[:len(e)] != e {
		t.Errorf("expect:\n%s got:\n%s", e, r.lastError)
	}
}

// this part doesnt work in minutes
// {
// 	// timeout exceeded
// 	args = []string{"120"} // seconds = more than a minute
// 	t.Logf("sleeping 120 %s", time.Now())
// 	err = r.runCommand("sleep", "", "", args)
// 	if err == nil {
// 		t.Error("timeout usecase did not return error")
// 	}
// 	if !errors.Is(err, context.DeadlineExceeded) {
// 		t.Errorf("error is not timeout: %s", err.Error())
// 	}
// }
