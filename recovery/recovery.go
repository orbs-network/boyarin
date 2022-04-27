package recovery

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/orbs-network/scribe/log"
)

//const DDMMYYYYhhmmss = "2006-01-02-15:04:05"
const (
	e_zero_content   = "e_zero_content"
	e_no_bash_prefix = "e_no_bash_prefix"
	//e_content_not_changed = "e_content_not_changed"
)

/////////////////////////////////////////////////
// JSON
// {
// 	"shell": {
// 	"bin": "bash",
// 	"run": [
// 		"https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/shared/disk_cleanup_1",
// 		"https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/shared/docker_cleanup_1"
// 	]
//   }
// }
type Shell struct {
	Bin string   `json:"bin"`
	Run []string `json:"run"`
}

type Instructions struct {
	Shell Shell `json:"shell"`
}

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
	lastHash    string
	lastOutput  string
	lastReadErr string
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
			//r.ticker = time.NewTicker(5 * time.Second) // DEBUG every 5 sec
			r.ticker = time.NewTicker(time.Duration(r.config.IntervalMinute) * time.Minute)

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
func (r *Recovery) readUrl(url string) (string, error) {
	logger.Info("recovery readUrl: " + url)
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

	// if !r.isNewContent(hashPath, body.Bytes()) {
	// 	return "", errors.New(e_content_not_changed)
	// }
	return body.String(), nil
}

/////////////////////////////////////////////////////////////
func getWDPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return cwd + "/boyar_recovery/"
}

/////////////////////////////////////////////////////////////
func (r *Recovery) tick() {
	logger.Info("Recovery tick")
	r.tickCount += 1
	r.lastTick = time.Now()

	// read json
	jsnTxt, err := r.readUrl(r.config.Url) //, getWDPath())
	if err != nil {
		r.lastReadErr = err.Error()
		logger.Error(err.Error())
		return
	}
	var result Instructions
	//var result map[string]interface{}
	err = json.Unmarshal([]byte(jsnTxt), &result)
	if err != nil {
		r.lastReadErr = err.Error()
		logger.Error(err.Error())
	}
	//no scripts to run
	scriptArr := result.Shell.Run
	if len(scriptArr) == 0 {
		r.lastReadErr = "json run array came empty"
		logger.Error(r.lastReadErr)
		return
	}
	// no executable for bash
	if result.Shell.Bin == "" {
		r.lastReadErr = "bin for exec was not specified in json"
		logger.Error(r.lastReadErr)
		return
	}
	// clean last output for status
	r.lastOutput = ""

	// execute all scripts serial
	for _, url := range scriptArr {
		// read script
		script, err := r.readUrl(url) //, getWDPath())
		if err != nil {
			r.lastReadErr = err.Error()
			logger.Error(err.Error())
		} else {
			r.runScript(result.Shell.Bin, script)
		}
	}

}
func (r *Recovery) runScript(bin, script string) {
	// reset error
	r.lastReadErr = ""

	// no prefix
	if len(script) < 4 {
		r.lastReadErr = "script length < 4"
		logger.Error(r.lastReadErr)
		return
	}

	// #!/ prefix check
	if script[:3] != "#!/" {
		r.lastReadErr = e_no_bash_prefix
		logger.Error(r.lastReadErr)
		return
	}

	// execute
	logger.Info("Recovery about to execute script")
	logger.Info("------------------------------")
	logger.Info(script)
	logger.Info("------------------------------")

	out, err := execBashScript(bin, script)
	r.lastExec = time.Now()
	if len(out) > 0 {
		logger.Info("output")
		logger.Info(out)
		r.lastOutput += out
	} else {
		logger.Error("exec Error")
		logger.Error(err.Error())
		r.lastOutput = "ERROR: " + err.Error()
	}
	logger.Info("------------------------------")
}

func (r *Recovery) Status() interface{} {
	return map[string]interface{}{
		"IntervalMinute": r.config.IntervalMinute,
		"Url":            r.config.Url,
		"tickCount":      r.tickCount,
		"lastTick":       r.lastTick,
		"lastExec":       r.lastExec,
		"lastHash":       r.lastHash,
		"lastOutput":     r.lastOutput,
		"lastReadError":  r.lastReadErr,
	}
}

func execBashScript(bin, script string) (string, error) {
	shell := os.Getenv("SHELL")
	if len(shell) == 0 {
		shell = "bash"
	}

	if bin != shell {
		logger.Info(fmt.Sprintf("OS ENV [SHELL] = [%s] but main.json wants to work with bin=[%s]", shell, bin))
	}
	cmd := exec.Command(bin)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, script)
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
