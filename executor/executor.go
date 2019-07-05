package executor

import (
	"errors"
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"log"
	"time"
)

func ErrTimeoutCommand(command command.Command, timeout time.Duration) error {
	return errors.New(fmt.Sprintf("'%s' timed out after %s", command, timeout))
}

type CommandExecutor interface {
	Execute(cmd command.Command) error
}

type PlanExecutor interface {
	Execute(plan plan.Plan) error
}

type planExecutor struct {
	shellExecutor CommandExecutor
	cfCliExecutor CommandExecutor
	timeout       time.Duration
	logger        *log.Logger
}

func NewExecutor(shellExecutor CommandExecutor, cfCliExecutor CommandExecutor, timeout time.Duration, logger *log.Logger) PlanExecutor {
	return planExecutor{
		shellExecutor: shellExecutor,
		cfCliExecutor: cfCliExecutor,
		timeout:       timeout,
		logger:        logger,
	}
}

func (p planExecutor) Execute(plan plan.Plan) (err error) {
	for _, cmd := range plan {
		p.logger.Println(fmt.Sprintf("$ %s", cmd))

		errChan := make(chan error, 1)
		go func() {
			switch typedCommand := cmd.(type) {
			case command.CfShellCommand:
				errChan <- p.shellExecutor.Execute(typedCommand)
			case command.CfCliCommand:
				errChan <- p.cfCliExecutor.Execute(typedCommand)
			}
		}()

		select {
		case err = <-errChan:
			if err != nil {
				return
			}
		case <-time.After(p.timeout):
			return ErrTimeoutCommand(cmd, p.timeout)
		}
		p.logger.Println()
	}
	return
}
