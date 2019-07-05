package executor

import (
	"github.com/springernature/halfpipe-cf-plugin/command"
	"log"
	"os"
	"os/exec"
)

type shellExecutor struct {
	logger *log.Logger
}

func NewShellExecutor(logger *log.Logger) CommandExecutor {
	return shellExecutor{logger: logger}
}

func (s shellExecutor) Execute(command command.Command) (err error) {
	cfPath, err := exec.LookPath("cf")
	if err != nil {
		return
	}

	cmd := exec.Cmd{
		Path:   cfPath,
		Args:   append([]string{cfPath}, command.Args()...),
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	return cmd.Run()
}