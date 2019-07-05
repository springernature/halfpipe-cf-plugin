package cleanup

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/executor"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
)

func TestGivesBackErrorIfGetAppFails(t *testing.T) {
	expectedError := errors.New("error")
	del := NewCleanupPlanner(helpers.NewMockCliConnection().WithGetAppsError(expectedError))

	_, err := del.GetPlan(manifest.Application{}, halfpipe_cf_plugin.Request{})

	assert.Equal(t, executor.ErrGetApps(expectedError), err)
}

func TestEmptyPlanIfNoOldApp(t *testing.T) {
	del := NewCleanupPlanner(helpers.NewMockCliConnection())

	plan, err := del.GetPlan(manifest.Application{
		Name: "my-app",
	}, halfpipe_cf_plugin.Request{})

	assert.Nil(t, err)
	assert.True(t, plan.IsEmpty())
}

func TestGivesBackADeletePlan(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	expectedApplicationName := helpers.CreateDeleteName(application.Name, 0)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("delete", expectedApplicationName, "-f"),
	}

	del := NewCleanupPlanner(helpers.NewMockCliConnection().WithApps(
		[]plugin_models.GetAppsModel{
			{
				Name: helpers.CreateDeleteName(application.Name, 0),
			},
		}))

	commands, err := del.GetPlan(application, halfpipe_cf_plugin.Request{})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackADeletePlanWithMultipleDeleteApps(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	del := NewCleanupPlanner(helpers.NewMockCliConnection().WithApps(
		[]plugin_models.GetAppsModel{
			{
				Name: helpers.CreateDeleteName(application.Name, 0),
			},
			{
				Name: application.Name,
			},
			{
				Name: helpers.CreateDeleteName(application.Name, 1),
			},
			{
				Name: helpers.CreateDeleteName("some-other-app", 0),
			},
		}))

	commands, err := del.GetPlan(application, halfpipe_cf_plugin.Request{})

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("delete", helpers.CreateDeleteName(application.Name, 0), "-f"),
		command.NewCfShellCommand("delete", helpers.CreateDeleteName(application.Name, 1), "-f"),
	}

	assert.Nil(t, err)
	assert.Len(t, commands, 2)
	assert.Equal(t, expectedPlan, commands)
}
