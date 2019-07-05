package executor

import "github.com/springernature/halfpipe-cf-plugin/command"

type cliExecutor struct {

}

func (c cliExecutor) Execute(cmd command.Command) error {
	panic("implement me")
}

func NewCliExecutor() CommandExecutor {
	return cliExecutor{

	}
}
