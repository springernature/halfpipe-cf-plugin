package command

import (
	"fmt"
	"regexp"
	"strings"
)

type CfShellCommand struct {
	command      string
	args         []string
}

func NewCfShellCommand(args ...string) Command {
	return CfShellCommand{
		command: "cf",
		args:    args,
	}
}

func (c CfShellCommand) Args() []string {
	return c.args
}

func (c CfShellCommand) String() string {
	var commandArgs = strings.Join(c.args, " ")

	if strings.HasPrefix(commandArgs, "login") {
		// If the command is login, a dirty replace of whatever comes after "-p "
		// to hide cf password from concourse console output
		cfLoginPasswordRegex := regexp.MustCompile(`-p ([a-zA-Z0-9_-]+)`)
		commandArgs = cfLoginPasswordRegex.ReplaceAllLiteralString(commandArgs, "-p ********")
	}

	return fmt.Sprintf("%s %s", c.command, commandArgs)
}
