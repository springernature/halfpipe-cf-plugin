package plan

import (
	"fmt"
	"log"
	"strings"
)

type Command struct {
	command string
	args    []string
}

func NewCfCommand(args ...string) Command {
	return Command{
		command: "cf",
		args:    args,
	}
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.command, strings.Join(c.args, " "))
}

type Plan []Command

func (c Plan) String() (s string) {
	s += "Planned execution\n"
	for _, p := range c {
		s += fmt.Sprintf("\t* %s\n", p)
	}
	return
}

func (c Plan) Execute(executor Executor, logger *log.Logger) (err error) {
	for _, p := range c {
		logger.Println("Executing", p)
		_, err = executor.CliCommand(p.args...)
		if err != nil {
			return
		}
		logger.Println("Suceeded!")
	}
	return
}

type Planner interface {
	Commands() (Plan, error)
}

type Executor interface {
	CliCommand(args ...string) ([]string, error)
}
