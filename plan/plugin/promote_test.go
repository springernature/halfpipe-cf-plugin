package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
)

func TestGivesBackAPromotePlan(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
		Routes: []string{
			"my-route1.domain1.com",
			"my-route2.domain2.com",
		},
	}
	testDomain := "domain.com"

	candidateAppName := createCandidateAppName(application.Name)
	expectedPlan := plan.Plan{
		plan.NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route1"),
		plan.NewCfCommand("map-route", candidateAppName, "domain2.com", "-n", "my-route2"),
		plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-CANDIDATE"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	promote := NewPromotePlanner()

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
