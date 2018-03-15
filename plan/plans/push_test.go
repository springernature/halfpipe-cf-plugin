package plans

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/util/manifest"
)

func TestGivesBackAPushPlan(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	expectedApplicationName := createCandidateAppName(application.Name)

	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"

	expectedPlan := Plan{
		NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "-n", expectedApplicationName, "-d", testDomain),
	}

	push := NewPush()

	commands, err := push.GetPlan(application, PluginRequest{
		ManifestPath: manifestPath,
		AppPath:      appPath,
		TestDomain:   testDomain,
	})

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}
