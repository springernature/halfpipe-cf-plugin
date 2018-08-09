package plan

import (
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfOldAppIsRunning(t *testing.T) {
	appName := "app"

	oldAppName := createOldAppName(appName)
	apps := []plugin_models.GetAppsModel{
		{Name: oldAppName, State: "running"},
	}

	err := checkCFState(appName, newMockAppsGetter().WithApps(apps))

	assert.Equal(t, err, ErrAppRunning(oldAppName))
}