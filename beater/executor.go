package beat

import (
	"bytes"
	"github.com/christiangalsterer/execbeat/config"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/robfig/cron"
	"os/exec"
	"strings"
	"time"
)

type Executor struct {
	execbeat *Execbeat
	config   config.ExecConfig
	schedule string
}

func NewExecutor(execbeat *Execbeat, config config.ExecConfig) *Executor {
	executor := &Executor{
		execbeat: execbeat,
		config:   config,
	}

	return executor
}

func (e *Executor) Run() {

	// Setup DocumentType
	if e.config.DocumentType == "" {
		e.config.DocumentType = config.DefaultDocumentType
	}

	//init the cron schedule
	if e.config.Schedule != "" {
		e.schedule = e.config.Schedule
	} else {
		e.schedule = config.DefaultSchedule
	}

	cron := cron.New()
	cron.AddFunc(e.config.Schedule, func() { e.runOneTime() })
	cron.Start()
}

func (e *Executor) runOneTime() error {
	var cmd *exec.Cmd
	var err error
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmdName := strings.TrimSpace(e.config.Command)
	cmdArgs := strings.TrimSpace(e.config.Args)

	// execute command
	logp.Debug("Execbeat", "Executing command: [%v] with args [%w]", cmdName, cmdArgs)
	now := time.Now()

	if len(cmdArgs) > 0 {
		cmd = exec.Command(cmdName, cmdArgs)
	} else {
		cmd = exec.Command(cmdName)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Start()
	if err != nil {
		logp.Err("An error occured while executing command: %v", err)
	}

	err = cmd.Wait()
	if err != nil {
		logp.Err("An error occured while executing command: %v", err)
	}

	commandEvent := Exec{
		Command: cmdName,
		StdOut:  stdout.String(),
		StdErr:  stderr.String(),
	}

	event := ExecEvent{
		ReadTime:     now,
		DocumentType: e.config.DocumentType,
		Fields:       e.config.Fields,
		Exec:         commandEvent,
	}

	e.execbeat.client.PublishEvent(event.ToMapStr())

	return nil
}

func (e *Executor) Stop() {
}
