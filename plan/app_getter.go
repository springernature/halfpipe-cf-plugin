package plan

import "code.cloudfoundry.org/cli/plugin/models"

type AppGetter interface {
	GetApp(string) (plugin_models.GetAppModel, error)
}
