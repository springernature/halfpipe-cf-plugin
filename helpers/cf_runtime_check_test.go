package helpers

import (
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfOldAppIsRunning(t *testing.T) {
	appName := "app"

	oldAppName := CreateOldAppName(appName)
	apps := []plugin_models.GetAppsModel{
		{Name: oldAppName, State: "running"},
	}

	err := CheckCFState(appName, NewMockCliConnection().WithApps(apps))

	assert.Equal(t, err, ErrAppRunning(oldAppName))
}