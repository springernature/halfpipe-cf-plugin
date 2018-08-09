package plan

import (
	"fmt"
	"log"
	"time"
	"code.cloudfoundry.org/cli/cf/errors"
)

func ErrTimeoutCommand(command Command, timeout time.Duration) error {
	return errors.New(fmt.Sprintf("'%s' timed out after %s", command, timeout))
}

type Plan []Command

func (p Plan) String() (s string) {
	if p.IsEmpty() {
		s += "# Nothing to do!"
		return
	}

	s += "# Planned execution\n"
	for _, command := range p {
		s += fmt.Sprintf("#\t* %s\n", command)
	}
	return
}

func (p Plan) Execute(executor Executor, timeoutInSeconds time.Duration, logger *log.Logger) (err error) {
	for _, command := range p {
		logger.Println(fmt.Sprintf("$ %s", command))

		errChan := make(chan error, 1)
		go func() {
			_, err = executor.CliCommand(command.Args()...)
			errChan <- err
		}()

		select {
		case err = <-errChan:
			if err != nil {
				return
			}
		case <-time.After(timeoutInSeconds):
			return ErrTimeoutCommand(command, timeoutInSeconds)
		}
		logger.Println()
	}
	return
}

func (p Plan) IsEmpty() bool {
	return len(p) == 0
}
