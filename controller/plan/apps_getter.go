package plan

import "code.cloudfoundry.org/cli/plugin/models"

type AppsGetter interface {
	GetApps() ([]plugin_models.GetAppsModel, error)
}
