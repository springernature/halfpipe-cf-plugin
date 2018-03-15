package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
)

type mockPlanner struct {
	plan  plan.Plan
	error error
}

func newMockPlanner() mockPlanner {
	return mockPlanner{}
}

func newMockPlannerWithError(err error) mockPlanner {
	return mockPlanner{
		error: err,
	}
}

func newMockPlannerWithPlan(plan plan.Plan) mockPlanner {
	return mockPlanner{
		plan: plan,
	}
}

func (m mockPlanner) GetPlan(application manifest.Application, request Request) (plan.Plan, error) {
	return m.plan, m.error
}

var manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
	return []manifest.Application{{}}, nil
}

func TestControllerReturnsErrorIfManifestReaderErrors(t *testing.T) {
	expectedError := errors.New("blurgh")
	manifestReaderWithError := func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, expectedError
	}

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestReaderWithError)

	_, err := controller.GetPlan(Request{Command: config.PUSH})
	assert.Equal(t, expectedError, err)
}

func TestControllerReturnsErrorForBadManifest(t *testing.T) {
	manifestReaderWithEmptyManifest := func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{}, nil
	}

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestReaderWithEmptyManifest)

	_, err := controller.GetPlan(Request{Command: config.PUSH})
	assert.Equal(t, ErrBadManifest, err)

	///

	manifestReaderWithManifestWithTwoApps := func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{
			{},
			{},
		}, nil
	}
	controller = NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestReaderWithManifestWithTwoApps)
	_, err = controller.GetPlan(Request{Command: config.PROMOTE})
	assert.Equal(t, ErrBadManifest, err)
}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	expectedErr := errors.New("Meehp")

	controller := NewPlanner(newMockPlannerWithError(expectedErr), newMockPlanner(), newMockPlanner(), manifestReader)
	_, err := controller.GetPlan(Request{Command: config.PUSH})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"
	expectedErr := ErrUnknownCommand(command)

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestReader)

	_, err := controller.GetPlan(Request{Command: command})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	expectedPlan := plan.Plan{
		plan.NewCfCommand("blurgh"),
	}

	controller := NewPlanner(newMockPlannerWithPlan(expectedPlan), newMockPlanner(), newMockPlanner(), manifestReader)

	commands, err := controller.GetPlan(Request{Command: config.PUSH})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
