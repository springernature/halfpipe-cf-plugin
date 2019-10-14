package push

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
)

func TestErrorsOutIfGetCurrentSpaceFails(t *testing.T) {
	expectedError := errors.New("Wryyyy")

	push := NewPushPlanner(helpers.NewMockCliConnection().WithSpaceError(expectedError))

	_, err := push.GetPlan(manifest.Application{}, halfpipe_cf_plugin.Request{})
	assert.Equal(t, halfpipe_cf_plugin.ErrGetCurrentSpace(expectedError), err)
}

func TestGivesBackAPushPlan(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"
	space := "dev"

	application := manifest.Application{
		Name: "my-app",
	}

	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)
	expectedApplicationHostname := helpers.CreateCandidateHostname(application.Name, space)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "--no-route", "--no-start"),
		command.NewCfShellCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		//NewCfShellCommand("set-health-check", expectedApplicationName, "http"),
		command.NewCfShellCommand("start", expectedApplicationName),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection().WithSpace(space))

	commands, err := push.GetPlan(application, halfpipe_cf_plugin.Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPushPlanForWorkerApp(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}
	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)

	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection())

	commands, err := push.GetPlan(application, halfpipe_cf_plugin.Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
	})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPushPlanIncludingPreStartCommands(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"
	space := "dev"
	preStartCommand := "cf add-network or something;cf something else"

	application := manifest.Application{
		Name: "my-app",
	}

	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)
	expectedApplicationHostname := helpers.CreateCandidateHostname(application.Name, space)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "--no-route", "--no-start"),
		command.NewCfShellCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		//NewCfShellCommand("set-health-check", expectedApplicationName, "http"),
		command.NewCfShellCommand("add-network", "or", "something"),
		command.NewCfShellCommand("something", "else"),
		command.NewCfShellCommand("start", expectedApplicationName),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection().WithSpace(space))

	commands, err := push.GetPlan(application, halfpipe_cf_plugin.Request{
		ManifestPath:    manifestPath,
		AppPath:         appPath,
		TestDomain:      testDomain,
		PreStartCommand: preStartCommand,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
