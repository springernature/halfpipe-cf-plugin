package plan

import (
	"os"
	"os/exec"

)

type Executor interface {
	CliCommand(args ...string) ([]string, error)
}

type cfCLIExecutor struct {
}

// This executor differs from the executor used in the plugin in that it
// executes CF binary through the operating system rather than through the plugin system.
func NewCFCliExecutor() Executor {
	return cfCLIExecutor{}
}

func (c cfCLIExecutor) CliCommand(args ...string) (out []string, err error) {
	execCmd := exec.Command("cf", args...) // #nosec disables the gas warning for this line.
	execCmd.Stdout = os.Stderr
	execCmd.Stderr = os.Stderr

	if err = execCmd.Start(); err != nil {
		return
	}

	if err = execCmd.Wait(); err != nil {
		return
	}

	return
}
