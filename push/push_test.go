package push

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"strconv"
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

func TestGivesBackAPushPlanForDockerImage(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	testDomain := "domain.com"
	space := "dev"

	request := halfpipe_cf_plugin.Request{
		ManifestPath:   manifestPath,
		TestDomain:     testDomain,
		DockerUsername: "wryyyy",
	}

	application := manifest.Application{
		Name: "my-app",
		Docker: manifest.DockerInfo{
			Image: "someImage",
		},
	}

	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)
	expectedApplicationHostname := helpers.CreateCandidateHostname(application.Name, space)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "--no-route", "--no-start", "--docker-image", application.Docker.Image, "--docker-username", request.DockerUsername),
		command.NewCfShellCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		//NewCfShellCommand("set-health-check", expectedApplicationName, "http"),
		command.NewCfShellCommand("start", expectedApplicationName),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection().WithSpace(space))

	commands, err := push.GetPlan(application, request)

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPushPlanForWorkerDockerImage(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	testDomain := "domain.com"

	request := halfpipe_cf_plugin.Request{
		ManifestPath:   manifestPath,
		TestDomain:     testDomain,
		DockerUsername: "kehe",
	}

	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
		Docker: manifest.DockerInfo{
			Image: "yay",
		},
	}
	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "--docker-image", application.Docker.Image, "--docker-username", request.DockerUsername),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection())

	commands, err := push.GetPlan(application, request)

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPushPlanWithInstances(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"
	space := "dev"
	instances := 1337

	application := manifest.Application{
		Name: "my-app",
	}

	expectedApplicationName := helpers.CreateCandidateAppName(application.Name)
	expectedApplicationHostname := helpers.CreateCandidateHostname(application.Name, space)

	expectedPlan := plan.Plan{
		command.NewCfShellCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "-i", strconv.Itoa(instances), "--no-route", "--no-start"),
		command.NewCfShellCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		//NewCfShellCommand("set-health-check", expectedApplicationName, "http"),
		command.NewCfShellCommand("start", expectedApplicationName),
	}

	push := NewPushPlanner(helpers.NewMockCliConnection().WithSpace(space))

	request := halfpipe_cf_plugin.Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
		Instances:    instances,
	}
	commands, err := push.GetPlan(application, request)

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
