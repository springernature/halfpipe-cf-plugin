package executor

import (
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)
func TestBadCommand(t *testing.T) {
	assert.Equal(t, ErrBadCommand, NewCliExecutor(helpers.NewMockCliConnection(), DevNullWriter).Execute(command.CfShellCommand{}))
}

func TestReturnsError(t *testing.T) {
	expectedErr := errors.New("sup")

	cmd := command.NewCfCliCommand(func(cliConnection halfpipe_cf_plugin.CliInterface, logger *log.Logger) error {
		_, err := cliConnection.GetApp("appName")
		return err
	})

	err := NewCliExecutor(helpers.NewMockCliConnection().WithGetAppError(expectedErr), DevNullWriter).Execute(cmd)
	assert.Equal(t, expectedErr, err)
}

func TestAllesOk(t *testing.T) {
	cmd := command.NewCfCliCommand(func(cliConnection halfpipe_cf_plugin.CliInterface, logger *log.Logger) error {
		_, err := cliConnection.GetApp("appName")
		return err
	})

	err := NewCliExecutor(helpers.NewMockCliConnection().WithApp(plugin_models.GetAppModel{}), DevNullWriter).Execute(cmd)
	assert.NoError(t, err)
}