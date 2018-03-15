package plans

import (
	"testing"

	"io/ioutil"
	"log"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/stretchr/testify/assert"
)

var DevNullWriter = log.New(ioutil.Discard, "", 0)

type MockExecutor struct {
	fun func(args ...string) ([]string, error)
}

func (m MockExecutor) CliCommand(args ...string) ([]string, error) {
	return m.fun(args...)
}

func TestCommand_String(t *testing.T) {
	c := NewCfCommand("push", "-f", "man")

	expected := "cf push -f man"
	assert.Equal(t, expected, c.String())
}

func TestPlan_String(t *testing.T) {
	p := Plan{
		NewCfCommand("push"),
		NewCfCommand("delete"),
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
		NewCfCommand("error"),
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
		NewCfCommand("ok"),
		NewCfCommand("ok"),
		NewCfCommand("error"),
		NewCfCommand("ok"),
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
		NewCfCommand("ok"),
		NewCfCommand("ok"),
		NewCfCommand("ok"),
		NewCfCommand("ok"),
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