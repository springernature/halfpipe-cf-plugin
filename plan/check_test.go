package plan

import (
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfDeleteAppNameIsThere(t *testing.T) {
	appName := "app"
	deleteAppName := createDeleteName(appName, 0)
	apps := []plugin_models.GetAppsModel{
		{Name: deleteAppName},
	}

	err := checkCFState(appName, "blah", "blah", newMockAppsGetter().WithApps(apps))

	assert.Equal(t, err, ErrAppNameExists(deleteAppName))
}

func TestFailsIfOldAppIsRunning(t *testing.T) {
	appName := "app"

	oldAppName := createOldAppName(appName)
	apps := []plugin_models.GetAppsModel{
		{Name: oldAppName, State: "running"},
	}

	err := checkCFState(appName, "blah", "blah", newMockAppsGetter().WithApps(apps))

	assert.Equal(t, err, ErrAppRunning(oldAppName))
}