package plan

import (
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
)

type mockPlanner struct {
	plan  Plan
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

func newMockPlannerWithPlan(plan Plan) mockPlanner {
	return mockPlanner{
		plan: plan,
	}
}

func (m mockPlanner) GetPlan(application manifest.Application, request halfpipe_cf_plugin.Request) (Plan, error) {
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

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{readError: expectedError})

	_, err := controller.GetPlan(halfpipe_cf_plugin.Request{Command: config.PUSH})
	assert.Equal(t, expectedError, err)
}

func TestControllerReturnsErrorForBadManifest(t *testing.T) {

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{manifest:manifest.Manifest{}})

	_, err := controller.GetPlan(halfpipe_cf_plugin.Request{Command: config.PUSH})
	assert.Equal(t, ErrBadManifest, err)


	controller = NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), newMockPlanner(), StubManifestReadWrite{manifest:manifest.Manifest{Applications: []manifest.Application{{}, {}}}})
	_, err = controller.GetPlan(halfpipe_cf_plugin.Request{Command: config.PROMOTE})
	assert.Equal(t, ErrBadManifest, err)
}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	expectedErr := errors.New("Meehp")

	controller := NewPlanner(newMockPlannerWithError(expectedErr), newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestWithOneApp)
	_, err := controller.GetPlan(halfpipe_cf_plugin.Request{Command: config.PUSH})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"
	expectedErr := ErrUnknownCommand(command)

	controller := NewPlanner(newMockPlanner(), newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestWithOneApp)

	_, err := controller.GetPlan(halfpipe_cf_plugin.Request{Command: command})

	assert.Equal(t, expectedErr, err)
}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	expectedPlan := Plan{
		command.NewCfShellCommand("blurgh"),
	}

	controller := NewPlanner(newMockPlannerWithPlan(expectedPlan), newMockPlanner(), newMockPlanner(), newMockPlanner(), manifestWithOneApp)

	commands, err := controller.GetPlan(halfpipe_cf_plugin.Request{Command: config.PUSH})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
