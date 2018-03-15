package plans

import (
	"testing"

	"code.cloudfoundry.org/cli/util/manifest"
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
	expectedPlan := Plan{

		NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route1"),
		NewCfCommand("map-route", candidateAppName, "domain2.com", "-n", "my-route2"),
		NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-CANDIDATE"),
		NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	promote := NewPromote()

	commands, err := promote.GetPlan(application, PluginRequest{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
