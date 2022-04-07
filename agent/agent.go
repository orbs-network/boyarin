package agent

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const DDMMYYYYhhmmss = "2006-01-02 15:04:05"

type Config struct {
	IntervalMinute uint
	Url            string
}

type Agent struct {
	config *Config
	ticker *time.Ticker
}

var single *Agent

func Init(c *Config) {
	//initialize static instance on load
	single = &Agent{config: c}
}

//GetInstanceA - get singleton instance pre-initialized
func GetInstance() *Agent {
	return single
}

/////////////////////////////
func (a *Agent) Start(start bool) {
	if start {
		if a.ticker == nil {
			dlPath := getDownloadPath()

			// ensure download hash folder
			err := os.MkdirAll(dlPath, 0777)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println("start Agent v1.0")
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
		fmt.Println("stop Agent")
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
		fmt.Printf("read hash file [%s] failed  %s", hashFile, err)
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
		fmt.Printf("faile to write hash [%s] failed  %e", hashFile, err)
	}

	return true
}

/////////////////////////////////////////////////////////////
// write as it downloads and not load the whole file into memory.
func DownloadFile(targetPath, url, hashPath string) (string, error) {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	fmt.Printf("response status: %s\n", resp.Status)
	if resp.ContentLength == 0 {
		return "", errors.New("conten size is ZERO")
	}

	defer resp.Body.Close()

	// read body
	body := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
	_, err = io.Copy(body, resp.Body)

	if !isNewContent(hashPath, body.Bytes()) {
		return "", errors.New("file content is not new")
	}

	// ensure download folder
	err = os.MkdirAll(targetPath, 0777)
	if err != nil {
		log.Fatal(err)
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
		fmt.Printf("error %s", err)
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
	fmt.Println("tick")

	targetPath := getTargetPath(dlPath)
	fmt.Printf("Download target path: %s\n", targetPath)
	filePath, err := DownloadFile(targetPath, fileUrl, dlPath)

	if err != nil {
		fmt.Printf("Download file err: %s\n", err.Error())
		return
	}
	fmt.Println("Downloaded: " + fileUrl)

	// execute
	fmt.Println("executing " + filePath + "!")
	output := execBash(filePath)
	fmt.Println("output:")
	fmt.Println(output)

}
