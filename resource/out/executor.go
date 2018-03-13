package out

import (
	"os/exec"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"os"
)

type cliExecutor struct {
}

func NewCliExecutor() plan.Executor {
	return cliExecutor{}
}

func (c cliExecutor) CliCommand(args ...string) (out []string, err error) {
	execCmd := exec.Command("cf", args...)
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
