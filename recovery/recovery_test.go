package recovery

import (
	"testing"
)

func Test_BoyarRecoveryDummy(t *testing.T) {
	t.Log("ALL GOOD!")
}

// func Test_BoyarRecoveryConfigSingleton(t *testing.T) {
// 	// init recovery config
// 	url := "http://localhost:8080/node/0xTEST/main.sh"

// 	// init recovery config
// 	config := Config{
// 		IntervalMinute: 1,
// 		Url:            url,
// 	}
// 	Init(&config)

// 	recovery1 := GetInstance()

// 	// get same instance
// 	recovery2 := GetInstance()
// 	if recovery1.config.Url != recovery2.config.Url {
// 		t.Error("config url in two instances is not equal")
// 	}
// 	if recovery1.config.IntervalMinute != recovery2.config.IntervalMinute {
// 		t.Error("config IntervalMinute in two instances is not equal")
// 	}
// }

// func Test_BoyarRecoveryDownloadErr(t *testing.T) {
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

// func Test_BoyarRecoveryDownloadOK(t *testing.T) {
// 	logger = log.GetLogger()

// 	url := "https://deployment.orbs.network/boyar_recovery/node/0x9f0988Cd37f14dfe95d44cf21f9987526d6147Ba/main.sh"

// 	dlPath := getDownloadPath()
// 	targetPath := getTargetPath(dlPath)

// 	// delete hash file so content will be new
// 	hashFile := dlPath + "last_hash.txt"
// 	err := os.Remove(hashFile)
// 	if err != nil {
// 		t.Errorf("remove [%s] failed", hashFile)
// 	}

// 	// download
// 	res, err := DownloadFile(targetPath, url, dlPath)

// 	if res == "" {
// 		t.Errorf("res for url[%s] is empty", url)
// 	}
// 	if err != nil {
// 		t.Errorf("err for url[%s] should not be nil %s", url, err.Error())
// 	}

// 	// download again - expect content not new
// 	res, err = DownloadFile(targetPath, url, dlPath)

// 	if err.Error() != "file content is not new" {
// 		t.Errorf("file content should have been the same")
// 	}

// }
