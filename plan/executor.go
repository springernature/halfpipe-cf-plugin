package plan

import (
	"log"
	"os"
	"os/exec"
)

type Executor interface {
	Execute(cmd Command) error
}

type shellExecutor struct {
	logger *log.Logger
}

func NewShellExecutor(logger *log.Logger) shellExecutor {
	return shellExecutor{logger: logger}
}

func (s shellExecutor) Execute(command Command) (err error) {
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
