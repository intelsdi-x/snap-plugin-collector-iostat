package command

import (
	"bytes"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io"
	"os/exec"
	"time"
)

type cmdRunner struct{}

func New() *cmdRunner {
	return &cmdRunner{}
}

func (c *cmdRunner) Run(cmd string, args []string) (io.Reader, error) {
	command := exec.Command(cmd, args...)
	resCh := make(chan []byte)
	errCh := make(chan error)
	go func() {
		out, err := command.CombinedOutput()
		if err != nil {
			errCh <- err
		}
		resCh <- out
	}()
	timer := time.After(2 * time.Second)
	select {
	case err := <-errCh:
		return nil, err
	case res := <-resCh:
		return bytes.NewReader(res), nil
	case <-timer:
		return nil, fmt.Errorf("time out (cmd:%v args:%v)", cmd, args)
	}
}

func (c *cmdRunner) Exec(cmd string, args []string) string {
	command := exec.Command(cmd, args...)
	outputBytes, err := command.CombinedOutput()
	if err != nil {
		log.Error(err)
	}
	return string(outputBytes[:])
}
