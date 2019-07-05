package check

import (
	"fmt"
	halfpipe_cf_plugin "github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"log"
	"time"
)

type check struct {
	candidateAppName string
	sleepBetweenRuns time.Duration
}

func NewCheckPlanner(sleepBetweenRuns time.Duration) plan.Planner {
	return &check{
		sleepBetweenRuns: sleepBetweenRuns,
	}
}

func (c *check) GetPlan(application manifest.Application, request halfpipe_cf_plugin.Request) (pl plan.Plan, err error) {
	c.candidateAppName = helpers.CreateCandidateAppName(application.Name)

	pl = append(pl, command.NewCfCliCommand(c.createCheckFun()))
	return
}

func (c check) createCheckFun() command.CmdFunc {
	return func(cliConnection halfpipe_cf_plugin.CliInterface, logger *log.Logger) (err error) {
		for {
			app, err := cliConnection.GetApp(c.candidateAppName)
			if err != nil {
				return err
			}

			var numRunning int
			for _, instance := range app.Instances {
				if instance.State == "running" {
					numRunning++
				}
			}

			logger.Println(fmt.Sprintf(`%d/%d instances running`, numRunning, len(app.Instances)))

			if len(app.Instances) != numRunning {
				time.Sleep(c.sleepBetweenRuns)
				continue
			}

			return nil
		}
	}
}
