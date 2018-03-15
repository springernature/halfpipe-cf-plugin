package resource

import (
	"os"
	"os/exec"

	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type cliExecutor struct {
}

func NewCliExecutor() plan.Executor {
	return cliExecutor{}
}

func (c cliExecutor) CliCommand(args ...string) (out []string, err error) {
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
