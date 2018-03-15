package resource

import (
	"os"
	"os/exec"

	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type cfCLIExecutor struct {
}

// This executor differs from the executor used in the plugin in that it
// executes CF binary trough the operating system rather than trough the plugin system.
func NewCFCliExecutor() plan.Executor {
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
