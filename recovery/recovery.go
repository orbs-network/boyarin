package recovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/orbs-network/scribe/log"
)

const (
	e_zero_content    = "e_zero_content"
	e_no_bash_prefix  = "e_no_bash_prefix"
	e_no_code_or_args = "e_no_code_or_args"
	e_json_no_binary  = "e_json_no_binary"
	//e_content_not_changed = "e_content_not_changed"
	DDMMYYYYhhmmss = "2006-01-02 15:04:05"
)

/////////////////////////////////////////////////
// INSTRUCTIONS JSON
// {
//     "bin": "/bin/bash",
//     "args": [],
//     "dir": null,
//     "stdins": [
//         "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/shared/disk_cleanup_1.sh",
//         "https://raw.githubusercontent.com/amihaz/staging-deployment/main/boyar_recovery/shared/docker_cleanup_1.sh"
//     ]
// }

///////////////////////////////////////////////
type Instructions struct {
	Bin    string   `json:"bin"`
	Args   []string `json:"args"`
	Dir    string   `json:dir`
	Stdins []string `json:"stdins"`
}

type Config struct {
	IntervalMinute uint
	TimeoutMinute  uint
	Url            string
}

type Recovery struct {
	config     Config
	ticker     *time.Ticker
	tickCount  uint32
	lastTick   time.Time
	lastExec   time.Time
	lastOutput string
	lastError  string
}

/////////////////////////////////////////////////
var single *Recovery
var logger log.Logger

/////////////////////////////////////////////////
func Init(c Config, _logger log.Logger) {
	//initialize static instance on load
	logger = _logger
	logger.Info("recovery - Init logger success")
	// default
	if c.TimeoutMinute == 0 {
		c.TimeoutMinute = 30
	}
	if c.IntervalMinute == 0 {
		c.IntervalMinute = 60 * 6
	}
	single = &Recovery{config: c, tickCount: 0}
}

//GetInstanceA - get singleton instance pre-initialized
func GetInstance() *Recovery {
	return single
}

/////////////////////////////////////////////////
func (r *Recovery) Start(start bool) {
	if start {
		logger.Info("recovery::start()")
		if r.ticker == nil {
			logger.Info("start boyar Recovery")
			//r.ticker = time.NewTicker(5 * time.Second) // DEBUG every 5 sec
			r.ticker = time.NewTicker(time.Duration(r.config.IntervalMinute) * time.Minute)

			go func() {
				// immediate
				r.tick()

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
		r.lastError = err.Error()
		logger.Error(err.Error())
		return
	}

	// read JSON
	var inst Instructions
	err = json.Unmarshal([]byte(jsnTxt), &inst)
	if err != nil {
		r.lastError = err.Error()
		logger.Error(err.Error())
		return
	}

	// mandatory
	if len(inst.Bin) == 0 {
		r.lastError = e_json_no_binary
		logger.Error(r.lastError)
		return
	}
	// optional - if no std in, args may be executed
	if len(inst.Stdins) == 0 {
		logger.Info("no stdins provided")
	}
	// read all code
	fullCode := ""
	for _, url := range inst.Stdins {
		// append code
		code, err := r.readUrl(url)
		if err != nil {
			r.lastError = err.Error()
			logger.Error(err.Error())
			return
		} else {
			fullCode += code + "\n"
		}
	}

	// execute all with timeout
	err = r.runCommand(inst.Bin, inst.Dir, fullCode, inst.Args)
	if err != nil {
		r.lastError = err.Error()
		logger.Error(r.lastError)
	}
}

/////////////////////////////////////////////////
func (r *Recovery) runCommand(bin, dir, code string, args []string) error {
	// reset error for status
	r.lastError = ""
	r.lastOutput = ""
	r.lastExec = time.Now()

	// no prefix
	if len(code) < 4 && len(args) == 0 {
		return errors.New(e_no_code_or_args)
	}

	// execute
	if code != "" {
		logger.Info("about to execute recovery code:" + code)
	} else {
		logger.Info("about to execute recovery args:" + strings.Join(args, ", "))
	}

	// timeout 5 minutes
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*time.Duration(r.config.TimeoutMinute))
	defer cancel()

	cmd := exec.CommandContext(ctx, bin, args...)

	// working dir
	if len(dir) > 0 {
		cmd.Dir = dir
	}

	// stdin code execution
	if code != "" {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}

		// stream code stdin
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, code)
		}()
	}

	// execute
	out, err := cmd.CombinedOutput()

	// keep stdout+err despite context or cmd errors
	if len(out) > 0 {
		r.lastOutput = string(out)
		logger.Info("output: " + string(out))
	}

	// context error such timeout
	if ctx.Err() != nil {
		return ctx.Err()
	}
	// cmd error
	if err != nil {
		return err
	}

	return nil
}

/////////////////////////////////////////////////
func (r *Recovery) Status() interface{} {
	nextTickTime := time.Time(r.lastTick)
	nextTickTime = nextTickTime.Add(time.Minute * time.Duration(r.config.IntervalMinute))

	if r.tickCount == 0 {
		return map[string]interface{}{
			"intervalMinute": r.config.IntervalMinute,
			"url":            r.config.Url,
			"tickCount":      "before first tick",
		}
	}
	return map[string]interface{}{
		"intervalMinute":    r.config.IntervalMinute,
		"url":               r.config.Url,
		"tickCount":         r.tickCount,
		"lastTick":          r.lastTick,
		"nextTickTime":      nextTickTime.Format(DDMMYYYYhhmmss),
		"lastExec":          r.lastExec,
		"lastOutput":        r.lastOutput,
		"lastError":         r.lastError,
		"execTimeoutMinute": r.config.TimeoutMinute,
	}
}
