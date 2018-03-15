package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/cf/errors"
)

type mockAppsGetter struct {
	apps  []plugin_models.GetAppsModel
	error error
}

func (m mockAppsGetter) GetApps() ([]plugin_models.GetAppsModel, error) {
	return m.apps, m.error
}

func newMockAppsGetter(apps []plugin_models.GetAppsModel, error error) mockAppsGetter {
	return mockAppsGetter{
		apps:  apps,
		error: error,
	}
}

func TestGivesBackErrorIfGetAppsFails(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(newMockAppsGetter([]plugin_models.GetAppsModel{}, expectedError))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}

func TestGivesBackAPromotePlanWhenThereIsNoOldApp(t *testing.T) {
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

	promote := NewPromotePlanner(newMockAppsGetter([]plugin_models.GetAppsModel{}, nil))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}


func TestGivesBackAPromotePlanWhenThereIsAnOldAppButWithDifferentName(t *testing.T) {
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

	promote := NewPromotePlanner(newMockAppsGetter([]plugin_models.GetAppsModel{
		{Name: "Ima a app"},
		{Name: "Me 2 lol"},
	}, nil))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPromotePlanWhenThereIsAnOldApp(t *testing.T) {
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
		plan.NewCfCommand("rename", application.Name, "my-app-OLD"),
		plan.NewCfCommand("stop", "my-app-OLD"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	promote := NewPromotePlanner(newMockAppsGetter([]plugin_models.GetAppsModel{
		{
			Name: application.Name,
		},
	}, nil))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPromotePlanWhenThereIsAnOldAppAndAnEvenOlder(t *testing.T) {
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
		plan.NewCfCommand("rename", "my-app-OLD", "my-app-DELETE"),
		plan.NewCfCommand("rename", application.Name, "my-app-OLD"),
		plan.NewCfCommand("stop", "my-app-OLD"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	promote := NewPromotePlanner(newMockAppsGetter([]plugin_models.GetAppsModel{
		{
			Name: application.Name,
		},
		{
			Name: createOldAppName(application.Name),
		},
	}, nil))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}
