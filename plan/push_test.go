package plan

import (
	"testing"

	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/cf/errors"
)

func TestErrorsOutIfGetCurrentSpaceFails(t *testing.T) {
	expectedError := errors.New("Wryyyy")

	push := NewPushPlanner(newMockCliConnection().WithSpaceError(expectedError))

	_, err := push.GetPlan(manifest.Application{}, Request{})
	assert.Equal(t, ErrGetCurrentSpace(expectedError), err)
}

func TestGivesBackAPushPlan(t *testing.T) {
	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"
	space := "dev"

	application := manifest.Application{
		Name: "my-app",
	}

	expectedApplicationName := createCandidateAppName(application.Name)
	expectedApplicationHostname := createCandidateHostname(application.Name, space)

	expectedPlan := Plan{
		NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "--no-route", "--no-start"),
		NewCfCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		//NewCfCommand("set-health-check", expectedApplicationName, "http"),
		NewCfCommand("start", expectedApplicationName),
	}

	push := NewPushPlanner(newMockCliConnection().WithSpace(space))

	commands, err := push.GetPlan(application, Request{
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
	expectedApplicationName := createCandidateAppName(application.Name)

	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"

	expectedPlan := Plan{
		NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath),
	}

	push := NewPushPlanner(newMockCliConnection())

	commands, err := push.GetPlan(application, Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
	})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}