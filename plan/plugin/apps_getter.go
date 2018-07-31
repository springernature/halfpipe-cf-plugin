package plugin

import "code.cloudfoundry.org/cli/plugin/models"

type AppsGetter interface {
	GetApps() ([]plugin_models.GetAppsModel, error)
	GetApp(appName string) (plugin_models.GetAppModel, error)
	CliCommandWithoutTerminalOutput(args ...string) ([]string, error)
}
