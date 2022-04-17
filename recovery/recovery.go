package recovery

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/orbs-network/scribe/log"
)

//const DDMMYYYYhhmmss = "2006-01-02-15:04:05"
const (
	e_zero_content        = "e_zero_content"
	e_no_bash_prefix      = "e_no_bash_prefix"
	e_content_not_changed = "e_content_not_changed"
)

type Config struct {
	IntervalMinute uint
	Url            string
}

type Recovery struct {
	config      Config
	ticker      *time.Ticker
	tickCount   uint32
	lastTick    time.Time
	lastExec    time.Time
	lastScript  string
	lastOutput  string
	lastReadErr string
	status      map[string]interface{}
}

var single *Recovery
var logger log.Logger

func Init(c Config, _logger log.Logger) {
	//initialize static instance on load
	logger = _logger
	logger.Info("recovery - Init logger success")
	single = &Recovery{config: c, tickCount: 0}
}

//GetInstanceA - get singleton instance pre-initialized
func GetInstance() *Recovery {
	return single
}

/////////////////////////////
func (r *Recovery) Start(start bool) {
	if start {
		logger.Info("recovery::start()")
		if r.ticker == nil {
			//dlPath := getDownloadPath()

			// ensure download hash folder
			//err := os.MkdirAll(dlPath, 0777)

			// if err != nil {
			// 	logger.Error(err.Error())
			// }

			logger.Info("start boyar Recovery")
			r.ticker = time.NewTicker(5 * time.Second) // DEBUG every 5 sec
			//r.ticker = time.NewTicker(time.Duration(r.config.IntervalMinute) * time.Minute)

			go func() {
				// immediate
				// r.lastTick = time.Now()
				// tick(r.config.Url, dlPath)

				// delay for next tick
				for range r.ticker.C {
					r.tick()
				}
			}()
		}
	} else { // STOP
		logger.Info("stop boyar Recovery")
		if r.ticker != nil {
			r.ticker.Stop()
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
		logger.Error(fmt.Sprintf("read hash file [%s] failed %s", hashFile, err))
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

	// ensure folder exist
	err = os.MkdirAll(hashPath, 0777)
	if err != nil {
		logger.Error(fmt.Sprintf("MkdirAll failed[%s], %s", hashPath, err.Error()))
		return false
	}

	// write
	err = ioutil.WriteFile(hashFile, []byte(hashHex), 0644)
	if err != nil {
		logger.Error(fmt.Sprintf("faile to write hash [%s] failed  %e", hashFile, err))
	}

	return true
}

/////////////////////////////////////////////////////////////
// write as it downloads and not load the whole file into memory.
// func DownloadFile(targetPath, url, hashPath string) (string, error) {
// 	logger.Info("recovery downloadURL: " + url)
// 	client := http.Client{
// 		Timeout: 5 * time.Second,
// 	}

// 	// Get the data
// 	resp, err := client.Get(url)
// 	//resp, err := http.Get(url) //might take too long - no timeout

// 	if err != nil {
// 		logger.Info("download ERROR " + err.Error())
// 		return "", err
// 	}
// 	logger.Info("response status: " + resp.Status)
// 	if resp.StatusCode != 200 {
// 		stat := fmt.Sprintf("status: %d", resp.StatusCode)
// 		return "", errors.New(stat)
// 	}

// 	if resp.ContentLength == 0 {
// 		return "", errors.New("conten size is ZERO")
// 	}

// 	defer resp.Body.Close()

// 	// read body
// 	body := new(bytes.Buffer)
// 	body.ReadFrom(resp.Body)
// 	// return buf.Len()

// 	// body := bytes.NewBuffer(make([]byte, 0, resp.ContentLength))
// 	// _, err = io.Copy(body, resp.Body)

// 	if !isNewContent(hashPath, body.Bytes()) {
// 		return "", errors.New("file content is not new")
// 	}

// 	// ensure download folder
// 	err = os.MkdirAll(targetPath, 0777)
// 	if err != nil {
// 		return "", err
// 	}

// 	//Create executable write only file
// 	filePath := targetPath + "/main.sh"
// 	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0555)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Write the body to file
// 	_, err = io.Copy(out, body)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer out.Close()

// 	return filePath, nil
// }
/////////////////////////////////////////////////////////////
func readUrl(url, hashPath string) (string, error) {
	logger.Info("recovery downloadURL: " + url)
	client := http.Client{
		Timeout: 5 * time.Second,
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		logger.Info("url get ERROR " + err.Error())
		return "", err
	}
	logger.Info("response status: " + resp.Status)
	if resp.StatusCode != 200 {
		stat := fmt.Sprintf("status: %d", resp.StatusCode)
		return "", errors.New(stat)
	}

	if resp.ContentLength == 0 {
		return "", errors.New(e_zero_content)
	}

	defer resp.Body.Close()

	// read body
	body := new(bytes.Buffer)
	body.ReadFrom(resp.Body)

	// #!/ prefix check
	if body.String()[:3] != "#!/" {
		return "", errors.New(e_no_bash_prefix)
	}

	if !isNewContent(hashPath, body.Bytes()) {
		return "", errors.New(e_content_not_changed)
	}
	return body.String(), nil
}

/////////////////////////////////////////////////////////////
// func execBashFile(path string) string {
// 	cmd, err := exec.Command("/bin/sh", path).Output()
// 	if err != nil {
// 		logger.Error(err.Error())
// 	}
// 	output := string(cmd)
// 	return output
// }

/////////////////////////////////////////////////////////////
func getWDPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd + "/boyar_recovery/"
}

/////////////////////////////////////////////////////////////
// func getTargetPath(dlPath string) string {

// 	// format date
// 	now := time.Now().UTC()
// 	timeStr := now.Format(DDMMYYYYhhmmss)
// 	targetPath := dlPath + timeStr

// 	return targetPath
// }

/////////////////////////////////////////////////////////////
func (r *Recovery) tick() {
	logger.Info("Recovery tick")
	r.tickCount += 1
	r.lastTick = time.Now()

	// targetPath := getTargetPath(dlPath)
	// logger.Info("Download target path: " + targetPath)
	// filePath, err := DownloadFile(targetPath, fileUrl, dlPath)
	code, err := readUrl(r.config.Url, getWDPath())
	if err != nil {
		r.lastReadErr = err.Error()
		logger.Error(err.Error())
		return
	}
	// reset error
	r.lastReadErr = ""
	// keep code for status
	r.lastScript = code

	// execute
	logger.Info("Recovery about to execute code")
	logger.Info("------------------------------")
	logger.Info(code)
	logger.Info("------------------------------")
	out := execBashScript(code)
	r.lastExec = time.Now()
	r.lastOutput = out
	logger.Info("------------------------------")
	logger.Info("output")
	logger.Info(out)
	logger.Info("------------------------------")

	// logger.Info("Downloaded: " + fileUrl)

	// // execute
	// logger.Info("recovery execute " + filePath)
	// output := execBash(filePath)
	// logger.Info("recovery execute output:")
	// logger.Info(output)
}

func (r *Recovery) Status() interface{} {
	r.status = map[string]interface{}{
		"IntervalMinute": r.config.IntervalMinute,
		"Url":            r.config.Url,
		"tickCount":      r.tickCount,
		"lastTick":       r.lastTick,
		"lastExec":       r.lastExec,
		"lastScript":     r.lastScript,
		"lastOutput":     r.lastOutput,
		"lastReadError":  r.lastReadErr,
	}
	return r.status
}

func execBashScript(script string) string {
	out, err := exec.Command("bash", "-c", script).Output()
	if err != nil {
		logger.Error(err.Error())
		return ""
	}
	return string(out)
}
