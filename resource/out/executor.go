package out

import (
	"os/exec"
	"log"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
)

type LogWriter log.Logger

func (w *LogWriter) Write(b []byte) (int, error) {
	(*log.Logger)(w).Print(string(b))
	return len(b), nil
}

type cliExecutor struct {
	Logger LogWriter
}

func NewCliExecutor(logger *log.Logger) plan.Executor {
	writer := (LogWriter)(*logger)

	return cliExecutor{
		writer,
	}
}

func (c cliExecutor) CliCommand(args ...string) (out []string, err error) {
	execCmd := exec.Command("cf", args...)
	execCmd.Stdout = &c.Logger
	execCmd.Stderr = &c.Logger
	err = execCmd.Run()
	if err != nil {
		return
	}

	return
}
