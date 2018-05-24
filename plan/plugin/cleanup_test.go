package plugin

import (
	"testing"

	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/plugin/models"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

type mockAppGetter struct {
	app   plugin_models.GetAppModel
	error error
}

func (m mockAppGetter) GetApp(appName string) (plugin_models.GetAppModel, error) {
	return m.app, m.error
}

func newMockAppGetter(apps plugin_models.GetAppModel, error error) mockAppGetter {
	return mockAppGetter{
		app:   apps,
		error: error,
	}
}

func TestGivesBackErrorIfGetAppFails(t *testing.T) {
	expectedError := errors.New("error")
	del := NewCleanupPlanner(newMockAppGetter(plugin_models.GetAppModel{}, expectedError))

	_, err := del.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}

func TestEmptyPlanIfNoOldApp(t *testing.T) {
	del := NewCleanupPlanner(newMockAppGetter(plugin_models.GetAppModel{}, errors.New("App my-app-DELETE not found")))

	plan, err := del.GetPlan(manifest.Application{
		Name: "my-app",
	}, Request{})

	assert.Nil(t, err)
	assert.True(t, plan.IsEmpty())
}

func TestGivesBackADeletePlan(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	expectedApplicationName := createDeleteName(application.Name, 0)

	expectedPlan := plan.Plan{
		plan.NewCfCommand("delete", expectedApplicationName, "-f"),
	}

	del := NewCleanupPlanner(newMockAppGetter(plugin_models.GetAppModel{Name: "my-app-DELETE"}, nil))

	commands, err := del.GetPlan(application, Request{})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}
