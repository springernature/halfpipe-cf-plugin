package check

import (
	"bytes"
	plugin_models "code.cloudfoundry.org/cli/plugin/models"
	"errors"
	halfpipe_cf_plugin "github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"testing"
	"time"
)

var DevNullWriter = log.New(ioutil.Discard, "", 0)

func TestPlan(t *testing.T) {
	appName := "app"
	plan, err := NewCheckPlanner(1 * time.Second).GetPlan(manifest.Application{Name: appName}, halfpipe_cf_plugin.Request{})
	assert.NoError(t, err)
	assert.Len(t, plan, 1)

	typedCmd := plan[0].(command.CfCliCommand)
	checkFunc := typedCmd.CommandFunc()

	t.Run("When the call to get app data fails we should return the error", func(t *testing.T) {
		expectedErr := errors.New("meehp")
		cliConnection := helpers.NewMockCliConnection().WithGetAppError(expectedErr)

		err := checkFunc(cliConnection, DevNullWriter)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("when there is only one instance and its running there should be no error", func(t *testing.T) {
		cliConnection := helpers.NewMockCliConnection().WithApp(plugin_models.GetAppModel{
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
			},
		})

		err := checkFunc(cliConnection, DevNullWriter)
		assert.NoError(t, err)
	})

	t.Run("when there two instance and both are running there should be no error", func(t *testing.T) {
		cliConnection := helpers.NewMockCliConnection().WithApp(plugin_models.GetAppModel{
			Instances: []plugin_models.GetApp_AppInstanceFields{
				{
					State: "running",
				},
				{
					State: "running",
				},
			},
		})

		var memLog bytes.Buffer
		err := checkFunc(cliConnection, log.New(&memLog, "", 0))

		assert.NoError(t, err)
		assert.Equal(t, `2/2 instances running
`, memLog.String())
	})

	t.Run("when the first call have half of the instances of the app and the second call fails we should return the error", func(t *testing.T) {
		expededErr := errors.New("meehp")
		numCalls := 0

		cliConnection := helpers.NewMockCliConnection().WithAppFun(func() (model plugin_models.GetAppModel, e error) {
			numCalls++
			if numCalls == 1 {
				model = plugin_models.GetAppModel{
					Instances: []plugin_models.GetApp_AppInstanceFields{
						{
							State: "running",
						},
						{
							State: "starting",
						},
					},
				}
				return
			} else if numCalls == 2 {
				e = expededErr
				return
			}
			panic("should never get here")
		})

		var memLog bytes.Buffer
		err := checkFunc(cliConnection, log.New(&memLog, "", 0))

		assert.Equal(t, 2, numCalls)
		assert.Equal(t, expededErr, err)
		assert.Equal(t, `1/2 instances running
`, memLog.String())
	})

	t.Run("when the first call have half of the instances of the app and the second call all instances are running there should be no error", func(t *testing.T) {
		numCalls := 0

		cliConnection := helpers.NewMockCliConnection().WithAppFun(func() (model plugin_models.GetAppModel, e error) {
			numCalls++
			if numCalls == 1 {
				model = plugin_models.GetAppModel{
					Instances: []plugin_models.GetApp_AppInstanceFields{
						{
							State: "running",
						},
						{
							State: "starting",
						},
					},
				}
				return
			} else if numCalls == 2 {
				model = plugin_models.GetAppModel{
					Instances: []plugin_models.GetApp_AppInstanceFields{
						{
							State: "running",
						},
						{
							State: "running",
						},
					},
				}
				return
			}
			panic("should never get here")
		})

		var memLog bytes.Buffer
		err := checkFunc(cliConnection, log.New(&memLog, "", 0))

		assert.Equal(t, 2, numCalls)
		assert.NoError(t, err)
		assert.Equal(t, `1/2 instances running
2/2 instances running
`, memLog.String())

	})

}
