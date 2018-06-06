package plugin

import (
	"testing"

	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
)

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

	expectedPlan := plan.Plan{
		plan.NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "--no-route"),
		plan.NewCfCommand("map-route", expectedApplicationName, testDomain, "-n", expectedApplicationHostname),
		plan.NewCfCommand("set-health-check", expectedApplicationName, "http"),
		plan.NewCfCommand("restart", expectedApplicationName),
	}

	push := NewPushPlanner(newMockAppsGetter())

	commands, err := push.GetPlan(application, Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
		Space:        space,
	})

	assert.Nil(t, err)
	assert.Len(t, commands, 4)
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

	expectedPlan := plan.Plan{
		plan.NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "--no-route"),
	}

	push := NewPushPlanner(newMockAppsGetter())

	commands, err := push.GetPlan(application, Request{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
	})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}
