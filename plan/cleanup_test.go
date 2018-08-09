package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"code.cloudfoundry.org/cli/plugin/models"
)

func TestGivesBackErrorIfGetAppFails(t *testing.T) {
	expectedError := errors.New("error")
	del := NewCleanupPlanner(newMockAppsGetter().WithGetAppsError(expectedError))

	_, err := del.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}

func TestEmptyPlanIfNoOldApp(t *testing.T) {
	del := NewCleanupPlanner(newMockAppsGetter())

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

	expectedPlan := Plan{
		NewCfCommand("delete", expectedApplicationName, "-f"),
	}

	del := NewCleanupPlanner(newMockAppsGetter().WithApps(
		[]plugin_models.GetAppsModel{
			{
				Name: createDeleteName(application.Name, 0),
			},
		}))

	commands, err := del.GetPlan(application, Request{})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackADeletePlanWithMultipleDeleteApps(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	del := NewCleanupPlanner(newMockAppsGetter().WithApps(
		[]plugin_models.GetAppsModel{
			{
				Name: createDeleteName(application.Name, 0),
			},
			{
				Name: application.Name,
			},
			{
				Name: createDeleteName(application.Name, 1),
			},
			{
				Name: createDeleteName("some-other-app", 0),
			},
		}))

	commands, err := del.GetPlan(application, Request{})

	expectedPlan := Plan{
		NewCfCommand("delete", createDeleteName(application.Name, 0), "-f"),
		NewCfCommand("delete", createDeleteName(application.Name, 1), "-f"),
	}

	assert.Nil(t, err)
	assert.Len(t, commands, 2)
	assert.Equal(t, expectedPlan, commands)
}
