package executor

import (
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

var DevNullWriter = log.New(ioutil.Discard, "", 0)

type mockExecutor struct {
	err error
	f   func(command command.Command) error
}

func newMockExecutor() CommandExecutor {
	return mockExecutor{
	}
}

func newMockExecutorWithError(err error) CommandExecutor {
	return mockExecutor{
		err: err,
	}
}

func newMockExecutorWithFunction(fun func(command command.Command) error) CommandExecutor {
	return mockExecutor{
		f: fun,
	}
}

func (m mockExecutor) Execute(cmd command.Command) error {
	if m.f != nil {
		return m.f(cmd)
	}
	return m.err
}

func TestPlan_ExecutePassesOnError(t *testing.T) {
	expectedError := errors.New("Expected error")

	p := plan.Plan{
		command.NewCfShellCommand("error"),
	}

	err := NewExecutor(newMockExecutorWithError(expectedError), newMockExecutor(), 10*time.Second, DevNullWriter).Execute(p)

	assert.Equal(t, expectedError, err)
}

func TestPlan_ExecutePassesOnErrorIfItHappensInTheMiddleOfThePlan(t *testing.T) {

	t.Run("Cf Shell Command", func(t *testing.T) {
		expectedError := errors.New("Expected error")
		var numberOfCalls int

		p := plan.Plan{
			command.NewCfShellCommand("ok"),
			command.NewCfShellCommand("ok"),
			command.NewCfShellCommand("error"),
			command.NewCfShellCommand("ok"),
		}

		err := NewExecutor(newMockExecutorWithFunction(func(command command.Command) error {
			numberOfCalls++
			if command.Args()[0] == "error" {
				return expectedError
			}
			return nil
		}), newMockExecutor(), 10*time.Second, DevNullWriter).Execute(p)

		assert.Equal(t, 3, numberOfCalls)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Cf Cli Command", func(t *testing.T) {
		expectedError := errors.New("Expected error")
		p := plan.Plan{
			command.NewCfShellCommand("ok"),
			command.NewCfShellCommand("ok"),
			command.NewCfCliCommand(),
			command.NewCfShellCommand("ok"),
		}

		err := NewExecutor(newMockExecutor(), newMockExecutorWithError(expectedError), 10*time.Second, DevNullWriter).Execute(p)

		assert.Equal(t, expectedError, err)
	})
}

func TestPlan_Execute(t *testing.T) {
	var numberOfCfShellCalls int
	var numberOfCliShellCalls int

	p := plan.Plan{
		command.NewCfShellCommand("ok"),
		command.NewCfCliCommand(),
		command.NewCfShellCommand("ok"),
		command.NewCfCliCommand(),
	}

	err := NewExecutor(newMockExecutorWithFunction(func(command command.Command) error {
		numberOfCfShellCalls++
		return nil
	}), newMockExecutorWithFunction(func(command command.Command) error {
		numberOfCliShellCalls++
		return nil
	}), 10*time.Second, DevNullWriter).Execute(p)

	assert.Nil(t, err)
	assert.Equal(t, 2, numberOfCfShellCalls)
	assert.Equal(t, 2, numberOfCliShellCalls)
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

func (m timeExecutor) Execute(cmd command.Command) error {
	time.Sleep(m.sleeptime)
	return m.err
}

func TestPlanWithTimeout(t *testing.T) {
	t.Run("Errors out if call takes longer than configured timeout", func(t *testing.T) {
		command := command.NewCfShellCommand("push")
		timeout := 1 * time.Second
		expectedError := ErrTimeoutCommand(command, timeout)
		p := plan.Plan{command}

		tE := newtimeExecutor().withSleepSeconds(2)
		err := NewExecutor(tE, newMockExecutor(), timeout, DevNullWriter).Execute(p)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Errors out if call errors out within the timeout", func(t *testing.T) {
		command := command.NewCfShellCommand("push")
		timeout := 1 * time.Second
		expectedError := errors.New("meehp")
		p := plan.Plan{command}

		tE := newtimeExecutor().withSleepSeconds(0).withError(expectedError)
		err := NewExecutor(tE, newMockExecutor(), timeout, DevNullWriter).Execute(p)
		assert.Equal(t, expectedError, err)
	})

	t.Run("Error is nil if it doesnt timeout and the cf call doesnt error", func(t *testing.T) {
		command := command.NewCfShellCommand("push")
		timeout := 1 * time.Second
		p := plan.Plan{command}

		tE := newtimeExecutor().withSleepSeconds(0)
		err := NewExecutor(tE, newMockExecutor(), timeout, DevNullWriter).Execute(p)
		assert.Nil(t, err)
	})
}