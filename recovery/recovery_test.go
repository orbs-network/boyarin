package recovery

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/orbs-network/scribe/log"
)

func Test_RecoveryConfigSingleton(t *testing.T) {
	// init recovery config
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh"

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

// func Test_RecoveryExecution(t *testing.T) {
// 	path, _ := os.Getwd()
// 	code, _ := os.ReadFile(path + "/test.sh")

// 	out := execBashReader(string(code))
// 	expect := "recovery script"
// 	sz := len(expect)
// 	if out[:sz] != expect {
// 		t.Errorf("expect:\t%s\ngot:\t%s", expect, out)
// 	}
// }

// func Test_RecoveryDownloadErr(t *testing.T) {
// 	url := "http://www.notfound.com/main.sh"

// 	dlPath := getDownloadPath()
// 	targetPath := getTargetPath(dlPath)
// 	res, err := DownloadFile(targetPath, url, dlPath)

// 	if res != "" {
// 		t.Errorf("res for url[%s] should be nil", res)
// 	}
// 	if err == nil {
// 		t.Errorf("err for url[%s] should not be nil", res)
// 	}

// 	if err.Error() != "status: 404" {
// 		t.Errorf("expected [status: 404] got[%s]", err.Error())
// 	}
// }

func Test_RecoveryBashPrefix(t *testing.T) {
	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/0xDEV.txt"
	//url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh"

	// init recovery config
	config := Config{
		IntervalMinute: 1,
		Url:            url,
	}

	logger = log.GetLogger()
	Init(config, logger)
	// does not return script but txt = "this node is 0xDEV"
	_, err := GetInstance().readUrl(url, "./boyar_recovery/")
	if err == nil {
		t.Error("read text did not cause error")
		return
	}
	if err.Error() != e_no_bash_prefix {
		t.Errorf("exepect e_no_bash_prefix, got %s", err.Error())

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

	res, err := GetInstance().readUrl(url, "./boyar_recovery/")
	if err == nil {
		t.Error("404 url did not result an error")
	}
	if res != "" {
		t.Error("404 url returned a result")
	}

	// get same instance

}

// func Test_RecoveryOK(t *testing.T) {
// 	logger = log.GetLogger()
// 	url := "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh"

// 	hashPath := getWDPath()
// 	hashFile := hashPath + "last_hash.txt"

// 	// delete hash file so content will be new
// 	if _, err := os.Stat(hashFile); !errors.Is(err, os.ErrNotExist) {
// 		err = os.Remove(hashFile)
// 		if err != nil {
// 			t.Errorf("remove [%s] failed", hashFile)
// 		}
// 	}

// 	config := Config{
// 		IntervalMinute: 1,
// 		Url:            url,
// 	}

// 	logger = log.GetLogger()
// 	Init(config, logger)

// 	// download
// 	res, err := GetInstance().readUrl(url, hashPath) //DownloadFile(targetPath, url, dlPath)

// 	if res == "" {
// 		t.Errorf("res for url[%s] is empty", url)
// 	}
// 	if err != nil {
// 		t.Errorf("err for url[%s] should not be nil %s", url, err.Error())
// 	}

// 	// download again - expect content not new
// 	res, err = GetInstance().readUrl(url, hashPath)

// 	if err.Error() != e_content_not_changed {
// 		t.Errorf("file content should have been the same")
// 	}
// }

func Test_RecoveryExec(t *testing.T) {
	logger = log.GetLogger()
	// script := "#!/bin/bash\n"
	// script += "echo \"one\"\n"
	// script += "echo \"two\"\n"
	// script += "cat yyy.txt\n"
	// script += "touch xxx.txt\n"
	// script += "echo \"three\""
	// url := "https://deployment.orbs.network/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh"
	// res, _ := http.Get(url)
	wd, _ := os.Getwd()
	script, _ := ioutil.ReadFile(wd + "/test2.sh")

	//out, err := execBashScript(string(script))
	out, err := execBashScript(string(script))
	if err != nil {
		t.Error(err)
		return
	}
	expect := "one\ntwo\nthree\n"
	if out != expect {
		t.Errorf("expect:\n%s got:\n%s", expect, out)
	}

}
