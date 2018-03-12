package controller

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/springernature/halfpipe-cf-plugin/controller/plan"
)

type MockPlan struct {
	plan  plan.Plan
	error error
}

func (m MockPlan) Commands() (plan.Plan, error) {
	return m.plan, m.error
}

func TestControllerReturnsErrorIfUnknownSubCommand(t *testing.T) {
	command := "not-supported"

	expectedErr := ErrUnknownCommand(command)

	_, err := NewController(command, "", "").Run()

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsErrorIfCallingOutToPlanFails(t *testing.T) {
	command := "halfpipe-push"
	expectedErr := errors.New("Meehp")

	controller := NewController(command, "", "")
	controller.pushPlan = MockPlan{error: expectedErr}

	_, err := controller.Run()

	assert.Equal(t, expectedErr, err)

}

func TestControllerReturnsTheCommandsForTheCommand(t *testing.T) {
	command := "halfpipe-push"
	expectedPlan := plan.Plan{}

	controller := NewController(command, "", "")
	controller.pushPlan = MockPlan{plan: expectedPlan}

	commands, err := controller.Run()

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
