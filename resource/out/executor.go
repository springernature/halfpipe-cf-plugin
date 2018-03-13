package out

import (
	"os/exec"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"os"
	"github.com/kr/pty"
	"syscall"
	"io"
)

type cliExecutor struct {
}

func NewCliExecutor() plan.Executor {
	return cliExecutor{}
}

func (c cliExecutor) CliCommand(args ...string) (out []string, err error) {
	execCmd := exec.Command("cf", args...)

	if err = runInFakeTTY(execCmd); err != nil {
		return
	}

	return
}

func runInFakeTTY(cmd *exec.Cmd) (err error) {
	pty, tty, err := pty.Open()
	if err != nil {
		return err
	}
	defer tty.Close()

	cmd.Stdout = tty
	cmd.Stdin = tty
	cmd.Stderr = tty
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.Setctty = true
	cmd.SysProcAttr.Setsid = true
	if err = cmd.Start(); err != nil {
		pty.Close()
		return
	}
	io.Copy(os.Stderr, pty)

	if err = cmd.Wait(); err != nil {
		pty.Close()
		return
	}
	return
}
