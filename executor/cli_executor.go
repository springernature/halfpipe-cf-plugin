package executor

import (
	"errors"
	halfpipe_cf_plugin "github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"log"
)

var ErrBadCommand = errors.New("this executor can only execute CfCliCommand commands")

type cliExecutor struct {
	cliConnection halfpipe_cf_plugin.CliInterface
	logger        *log.Logger
}

func (c cliExecutor) Execute(cmd command.Command) (err error) {
	switch typedCmd := cmd.(type) {
	case command.CfCliCommand:
		err = typedCmd.CommandFunc()(c.cliConnection, c.logger)
	default:
		err = ErrBadCommand
	}
	return
}

func NewCliExecutor(cliConnection halfpipe_cf_plugin.CliInterface, logger *log.Logger) CommandExecutor {
	return cliExecutor{
		cliConnection: cliConnection,
		logger:        logger,
	}
}
