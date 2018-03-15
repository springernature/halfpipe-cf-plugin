package plan

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/plan/plans"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/util/manifest"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/springernature/halfpipe-cf-plugin"
)

type MockPlan struct {
	plan  plans.Plan
	error error
}

func (m MockPlan) GetPlan(application manifest.Application) (plans.Plan, error) {
	return m.plan, m.error
}

type MockAppsGetter struct{}

func (MockAppsGetter) GetApps() ([]plugin_models.GetAppsModel, error) {
	panic("implement me")
}

func TestControllerReturnsErrorIfManifestReaderErrors(t *testing.T) {
	expectedError := errors.New("blurgh")
	command := halfpipe_cf_plugin.PUSH

	controller := NewPlanner("", "", "", MockAppsGetter{})
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{}, expectedError
	}

	_, err := controller.GetPlan(command)
	assert.Equal(t, expectedError, err)
}

func TestControllerReturnsErrorForBadManifest(t *testing.T) {
	controller := NewPlanner("", "", "", MockAppsGetter{})
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{}, nil
	}

	_, err := controller.GetPlan(halfpipe_cf_plugin.PUSH)
	assert.Equal(t, ErrBadManifest, err)

	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{
			{},
			{},
		}, nil
	}
	_, err = controller.GetPlan(halfpipe_cf_plugin.PROMOTE)
	assert.Equal(t, ErrBadManifest, err)
}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	expectedErr := errors.New("Meehp")

	controller := NewPlanner("", "", "", MockAppsGetter{})
	controller.pushPlan = MockPlan{error: expectedErr}
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	_, err := controller.GetPlan(halfpipe_cf_plugin.PUSH)

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"
	expectedErr := ErrUnknownCommand(command)

	controller := NewPlanner("", "", "", MockAppsGetter{})
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	_, err := controller.GetPlan(command)

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	expectedPlan := plans.Plan{}

	controller := NewPlanner("", "", "", MockAppsGetter{})
	controller.pushPlan = MockPlan{plan: expectedPlan}
	controller.manifestReader = func(pathToManifest string) ([]manifest.Application, error) {
		return []manifest.Application{{}}, nil
	}

	commands, err := controller.GetPlan(halfpipe_cf_plugin.PUSH)

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
