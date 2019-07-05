package command

import "github.com/springernature/halfpipe-cf-plugin/executor"

type CfCliCommand struct {
	commandFunc func(cliConnection executor.CliInterface) error
}

func (c CfCliCommand) String() string {
	return "Checking that all app instances are in running state"
}

func (c CfCliCommand) Args() []string {
	panic("not used")
}

func (c CfCliCommand) CommandFunc() func(cliConnection executor.CliInterface) error {
	return c.commandFunc
}

func NewCfCliCommand(commandFunc func(cliConnection executor.CliInterface) error) Command {
	return CfCliCommand{
		commandFunc: commandFunc,
	}
}
