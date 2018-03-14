package controller

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/util/manifest"
)

type MockPlan struct {
	plan  plan.Plan
	error error
}

func (m MockPlan) GetPlan(application manifest.Application) (plan.Plan, error) {
	return m.plan, m.error
}

func TestControllerReturnsErrorIfManifestReaderErrors(t *testing.T) {
	expectedError := errors.New("blurgh")
	command := "halfpipe-push"

	controller := NewController(command, "", "", "")
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{}, expectedError
	}

	_, err := controller.GetPlan()
	assert.Equal(t, expectedError, err)
}

func TestControllerReturnsErrorForBadManifest(t *testing.T) {
	command := "halfpipe-push"

	controller := NewController(command, "", "", "")
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{}, nil
	}

	_, err := controller.GetPlan()
	assert.Equal(t, ErrBadManifest, err)

	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{
			{},
			{},
		}, nil
	}
	_, err = controller.GetPlan()
	assert.Equal(t, ErrBadManifest, err)
}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	command := "halfpipe-push"
	expectedErr := errors.New("Meehp")

	controller := NewController(command, "", "", "")
	controller.pushPlan = MockPlan{error: expectedErr}
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	_, err := controller.GetPlan()

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"

	expectedErr := ErrUnknownCommand(command)

	controller := NewController(command, "", "", "")
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	_, err := controller.GetPlan()

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	command := "halfpipe-push"
	expectedPlan := plan.Plan{}

	controller := NewController(command, "", "", "")
	controller.pushPlan = MockPlan{plan: expectedPlan}
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	commands, err := controller.GetPlan()

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
