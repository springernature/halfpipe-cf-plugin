package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
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

type StubManifestReadWrite struct {
	manifest  manifest.Manifest
	readError error
}

var manifestWithOneApp = StubManifestReadWrite{manifest:manifest.Manifest{Applications: []manifest.Application{{}}}}

func (s StubManifestReadWrite) ReadManifest(path string) (manifest.Manifest, error) {
	return s.manifest, s.readError
}

func (StubManifestReadWrite) WriteManifest(path string, application manifest.Application) (error) {
	panic("implement me")
}

func TestControllerReturnsErrorIfManifestReaderErrors(t *testing.T) {
	expectedError := errors.New("blurgh")

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{readError: expectedError})

	_, err := controller.GetPlan(Request{Command: config.PUSH})
	assert.Equal(t, expectedError, err)
}

func TestControllerReturnsErrorForBadManifest(t *testing.T) {

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{manifest:manifest.Manifest{}})

	_, err := controller.GetPlan(Request{Command: config.PUSH})
	assert.Equal(t, ErrBadManifest, err)


	controller = NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{manifest:manifest.Manifest{Applications: []manifest.Application{{}, {}}}})
	_, err = controller.GetPlan(Request{Command: config.PROMOTE})
	assert.Equal(t, ErrBadManifest, err)
}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	expectedErr := errors.New("Meehp")

	controller := NewPlanner(newMockPlannerWithError(expectedErr), newMockPlanner(), newMockPlanner(), manifestWithOneApp)
	_, err := controller.GetPlan(Request{Command: config.PUSH})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"
	expectedErr := ErrUnknownCommand(command)

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestWithOneApp)

	_, err := controller.GetPlan(Request{Command: command})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	expectedPlan := plan.Plan{
		plan.NewCfCommand("blurgh"),
	}

	controller := NewPlanner(newMockPlannerWithPlan(expectedPlan), newMockPlanner(), newMockPlanner(), manifestWithOneApp)

	commands, err := controller.GetPlan(Request{Command: config.PUSH})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
