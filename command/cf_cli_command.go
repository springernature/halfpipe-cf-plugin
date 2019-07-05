package command

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"log"
)

type CmdFunc func(cliConnection halfpipe_cf_plugin.CliInterface, logger *log.Logger) error

type CfCliCommand struct {
	commandFunc CmdFunc
}

func (c CfCliCommand) String() string {
	return "Checking that all app instances are in running state"
}

func (c CfCliCommand) Args() []string {
	panic("not used")
}

func (c CfCliCommand) CommandFunc() CmdFunc {
	return c.commandFunc
}

func NewCfCliCommand(commandFunc CmdFunc) Command {
	return CfCliCommand{
		commandFunc: commandFunc,
	}
}
