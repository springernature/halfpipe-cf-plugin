package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGivesBackAPushPlanThatPushesDataWithoutSpecifyingAppBits(t *testing.T) {

	manifetPath := "path/to/manifest.yml"

	expectedPlan := Plan{
		NewCfCommand("push", "-f", manifetPath),
	}

	push := NewPush(manifetPath, "")

	commands, err := push.GetPlan()

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPushPlanThatPushesDataWithAppBits(t *testing.T) {

	manifestPath := "path/to/manifest.yml"
	appPath := "path/to/app.jar"

	expectedPlan := Plan{
		NewCfCommand("push", "-f", manifestPath, "-p", appPath),
	}

	push := NewPush(manifestPath, appPath)

	commands, err := push.GetPlan()

	assert.Nil(t, err)
	assert.Len(t, commands, 1)
	assert.Equal(t, expectedPlan, commands)
}
