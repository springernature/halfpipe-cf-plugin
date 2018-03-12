package plan

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/stretchr/testify/assert"
	"log"
	"io/ioutil"
)

var DevNullWriter = log.New(ioutil.Discard, "", 0)

type MockExecutor struct {
	fun func(args ...string) ([]string, error)
}

func (m MockExecutor) CliCommand(args ...string) ([]string, error) {
	return m.fun(args...)
}

func TestCommand_String(t *testing.T) {
	c := Command{
		args: []string{"push", "-f", "man"},
	}

	expected := "cf push -f man"
	assert.Equal(t, expected, c.String())
}

func TestPlan_String(t *testing.T) {
	p := Plan{
		{[]string{"push"}},
		{[]string{"delete"}},
	}

	expected := `Planned execution
	* cf push
	* cf delete
`
	assert.Equal(t, expected, p.String())
}

func TestPlan_ExecutePassesOnError(t *testing.T) {
	expectedError := errors.New("Expected error")

	p := Plan{
		{[]string{"error"}},
	}

	err := p.Execute(MockExecutor{
		func(args ...string) ([]string, error) {
			return []string{}, expectedError
		},
	}, DevNullWriter)

	assert.Equal(t, expectedError, err)
}

func TestPlan_ExecutePassesOnErrorIfItHappensInTheMiddleOfThePlan(t *testing.T) {
	expectedError := errors.New("Expected error")
	var numberOfCalls int

	p := Plan{
		{[]string{"ok"}},
		{[]string{"ok"}},
		{[]string{"error"}},
		{[]string{"ok"}},
	}

	err := p.Execute(MockExecutor{
		func(args ...string) ([]string, error) {
			numberOfCalls += 1
			if args[0] == "error" {
				return []string{}, expectedError
			}
			return []string{}, nil
		},
	}, DevNullWriter)

	assert.Equal(t, 3, numberOfCalls)
	assert.Equal(t, expectedError, err)
}

func TestPlan_Execute(t *testing.T) {
	var numberOfCalls int

	p := Plan{
		{[]string{"ok"}},
		{[]string{"ok"}},
		{[]string{"ok"}},
		{[]string{"ok"}},
	}

	err := p.Execute(MockExecutor{
		func(args ...string) ([]string, error) {
			numberOfCalls += 1
			return []string{}, nil
		},
	}, DevNullWriter)

	assert.Nil(t, err)
	assert.Equal(t, 4, numberOfCalls)
}
