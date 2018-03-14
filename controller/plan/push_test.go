package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/util/manifest"
	"strings"
)

func TestGivesBackAPushPlan(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
	}
	expectedApplicationName := strings.Join([]string{application.Name, "CANDIDATE"}, "-")

	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"
	testDomain := "domain.com"

	expectedPlan := Plan{
		NewCfCommand("push", expectedApplicationName, "-f", manifestPath, "-p", appPath, "-n", expectedApplicationName, "-d", testDomain),
	}

	push := NewPush(manifestPath, appPath, testDomain)

	commands, err := push.GetPlan(application)

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}
