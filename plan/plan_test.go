package plan

import (
	"io/ioutil"
	"log"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"github.com/stretchr/testify/assert"
	"time"
)

var DevNullWriter = log.New(ioutil.Discard, "", 0)

type mockExecutor struct {
	err error
	f   func(command Command) error
}

func newMockExecutorWithError(err error) Executor {
	return mockExecutor{
		err: err,
	}
}

func newMockExecutorWithFunction(fun func(command Command) error) Executor {
	return mockExecutor{
		f: fun,
	}
}

func (m mockExecutor) Execute(cmd Command) error {
	if m.f != nil {
		return m.f(cmd)
	}
	return m.err
}

func TestPlan_String(t *testing.T) {
	p := Plan{
		NewCfCommand("push"),
		NewCfCommand("delete"),
	}

	expected := `# Planned execution
#	* cf push
#	* cf delete
`
	assert.Equal(t, expected, p.String())
}

func TestPlan_ExecutePassesOnError(t *testing.T) {
	expectedError := errors.New("Expected error")

	p := Plan{
		NewCfCommand("error"),
	}

	err := p.Execute(newMockExecutorWithError(expectedError), 10*time.Second, DevNullWriter)

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

	err := p.Execute(newMockExecutorWithFunction(func(command Command) error {
		numberOfCalls++
		if command.Args()[0] == "error" {
			return expectedError
		}
		return nil
	}), 10*time.Second, DevNullWriter)

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

	err := p.Execute(newMockExecutorWithFunction(func(command Command) error {
		numberOfCalls++
		return nil
	}), 10*time.Second, DevNullWriter)

	assert.Nil(t, err)
	assert.Equal(t, 4, numberOfCalls)
}

type timeExecutor struct {
	sleeptime time.Duration
	err       error
}

func newtimeExecutor() timeExecutor {
	return timeExecutor{}
}

func (m timeExecutor) withSleepSeconds(duration time.Duration) timeExecutor {
	m.sleeptime = duration * time.Second
	return m
}
func (m timeExecutor) withError(err error) timeExecutor {
	m.err = err
	return m
}

func (m timeExecutor) Execute(cmd Command) error {
	time.Sleep(m.sleeptime)
	return m.err
}

func TestPlanWithTimeout(t *testing.T) {
	t.Run("Errors out if call takes longer than configured timeout", func(t *testing.T) {
		command := NewCfCommand("push")
		timeout := 1 * time.Second
		expectedError := ErrTimeoutCommand(command, timeout)
		p := Plan{command}

		tE := newtimeExecutor().withSleepSeconds(2)
		err := p.Execute(tE, timeout, DevNullWriter)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Errors out if call errors out within the timeout", func(t *testing.T) {
		command := NewCfCommand("push")
		timeout := 1 * time.Second
		expectedError := errors.New("meehp")
		p := Plan{command}

		tE := newtimeExecutor().withSleepSeconds(0).withError(expectedError)
		err := p.Execute(tE, timeout, DevNullWriter)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error is nil if it doesnt timeout and the cf call doesnt error", func(t *testing.T) {
		command := NewCfCommand("push")
		timeout := 1 * time.Second
		p := Plan{command}

		tE := newtimeExecutor().withSleepSeconds(0)
		err := p.Execute(tE, timeout, DevNullWriter)
		assert.Nil(t, err)
	})

}
