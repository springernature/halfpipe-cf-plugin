package executor

import "code.cloudfoundry.org/cli/plugin/models"

type CliInterface interface {
	GetApps() ([]plugin_models.GetAppsModel, error)
	GetApp(appName string) (plugin_models.GetAppModel, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
	GetCurrentSpace() (plugin_models.Space, error)
}
