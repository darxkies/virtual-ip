package pkg

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

type CommandRunner interface {
	Run(command string) error
}

type ShellCommandRunner struct {
	timeout time.Duration
}

func NewShellCommandRunner(timeout time.Duration) ShellCommandRunner {
	return ShellCommandRunner{timeout: timeout}
}

func (runner ShellCommandRunner) Run(command string) error {
	_context, cancel := context.WithTimeout(context.Background(), runner.timeout*time.Second)
	defer cancel()

	log.WithFields(log.Fields{"command": command}).Debug("Command started")

	cmd := exec.CommandContext(_context, "sh", "-c", command)

	output, error := cmd.CombinedOutput()
	if error != nil {
		log.WithFields(log.Fields{"command": command, "error": error}).Debug("Command failed")

		return fmt.Errorf("Command '%s' failed with error '%s' (Output: %s)", command, error, output)
	}

	log.WithFields(log.Fields{"command": command, "output": string(output)}).Debug("Command ended")

	return nil
}
