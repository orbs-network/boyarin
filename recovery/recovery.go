package recovery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/orbs-network/scribe/log"
)

const DDMMYYYYhhmmss = "2006-01-02 15:04:05"

type Config struct {
	IntervalMinute uint
	Url            string
}

type Recovery struct {
	config *Config
	ticker *time.Ticker
}

var single *Recovery
var logger log.Logger

func Init(c *Config) {
	//initialize static instance on load
	single = &Recovery{config: c}
}

//GetInstanceA - get singleton instance pre-initialized
func GetInstance() *Recovery {
	return single
}

/////////////////////////////
func (a *Recovery) Start(start bool) {
	if start {
		if a.ticker == nil {
			dlPath := getDownloadPath()

			// ensure download hash folder
			err := os.MkdirAll(dlPath, 0777)
			if err != nil {
				log.Error(err)
			}

			logger.Info("start boyar Recovery")
			tick(a.config.Url, dlPath)
			a.ticker = time.NewTicker(5 * time.Second) // DEBUG
			//a.ticker = time.NewTicker(time.Duration(a.config.IntervalMinute) * time.Minute)

			go func() {
				for range a.ticker.C {
					tick(a.config.Url, dlPath)
				}
			}()
		}
	} else { // STOP
		logger.Info("stop boyar Recovery")
		if a.ticker != nil {
			a.ticker.Stop()
		}
	}
}

/////////////////////////////////////////////////////////////
// write as it downloads and not load the whole file into memory.
func isNewContent(hashPath string, body []byte) bool {
	hashFile := hashPath + "last_hash.txt"

	// load last hash
	lastHash, err := ioutil.ReadFile(hashFile)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Error(errors.New(fmt.Sprintf("read hash file [%s] failed  %s", hashFile, err)))
		return false
	}

	// sha256 on body
	sha := sha256.Sum256(body)

	// save hash 256 = 64 chars
	hashHex := make([]byte, 64)
	hex.Encode(hashHex, sha[:])

	// file content hasnt changed
	if lastHash != nil && string(hashHex) == string(lastHash) {
		return false
	}

	// write
	err = ioutil.WriteFile(hashFile, []byte(hashHex), 0644)
	if err != nil {
		log.Error(errors.New(fmt.Sprintf("faile to write hash [%s] failed  %e", hashFile, err)))
	}

	return true
}

/////////////////////////////////////////////////////////////
// write as it downloads and not load the whole file into memory.
func DownloadFile(targetPath, url, hashPath string) (string, error) {
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Get the data
	resp, err := client.Get(url)
	//resp, err := http.Get(url) //might take too long - no timeout

	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		stat := fmt.Sprintf("status: %d", resp.StatusCode)
		return "", errors.New(stat)
	}

	logger.Info("response status: " + resp.Status)
	if resp.ContentLength == 0 {
		return "", errors.New("conten size is ZERO")
	}

	defer resp.Body.Close()

	// read body
	body := new(bytes.Buffer)
	body.ReadFrom(resp.Body)
	// return buf.Len()

	// body := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
	// _, err = io.Copy(body, resp.Body)

	if !isNewContent(hashPath, body.Bytes()) {
		return "", errors.New("file content is not new")
	}

	// ensure download folder
	err = os.MkdirAll(targetPath, 0777)
	if err != nil {
		return "", err
	}

	//Create executable write only file
	filePath := targetPath + "/main.sh"
	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0555)
	if err != nil {
		return "", err
	}

	// Write the body to file
	_, err = io.Copy(out, body)
	if err != nil {
		return "", err
	}
	defer out.Close()

	return filePath, nil
}

func execBash(path string) string {
	cmd, err := exec.Command("/bin/sh", path).Output()
	if err != nil {
		log.Error(err)
	}
	output := string(cmd)
	return output
}

/////////////////////////////////////////////////////////////
func getDownloadPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd + "/download/"
}

/////////////////////////////////////////////////////////////
func getTargetPath(dlPath string) string {

	// format date
	now := time.Now().UTC()
	timeStr := now.Format(DDMMYYYYhhmmss)
	targetPath := dlPath + timeStr

	return targetPath
}

/////////////////////////////////////////////////////////////
func tick(fileUrl, dlPath string) {
	logger.Info("Recovery tick")

	targetPath := getTargetPath(dlPath)
	logger.Info("Download target path: " + targetPath)
	filePath, err := DownloadFile(targetPath, fileUrl, dlPath)

	if err != nil {
		log.Error(err)
		return
	}
	logger.Info("Downloaded: " + fileUrl)

	// execute
	logger.Info("executing " + filePath + "!")
	output := execBash(filePath)
	logger.Info("output:")
	logger.Info(output)

}
