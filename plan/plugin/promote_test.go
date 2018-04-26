package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/plan"
)

type mockAppsGetter struct {
	apps      []plugin_models.GetAppsModel
	appsError error
	cliError  error
	cliOutput []string
}

func (m mockAppsGetter) CliCommandWithoutTerminalOutput(args ...string) ([]string, error) {
	return m.cliOutput, m.cliError
}

func (m mockAppsGetter) GetApps() ([]plugin_models.GetAppsModel, error) {
	return m.apps, m.appsError
}

func (m mockAppsGetter) WithGetAppsError(error error) mockAppsGetter {
	m.appsError = error
	return m
}

func (m mockAppsGetter) WithCliError(error error) mockAppsGetter {
	m.cliError = error
	return m
}

func (m mockAppsGetter) WithApps(apps []plugin_models.GetAppsModel) mockAppsGetter {
	m.apps = apps
	return m
}

func (m mockAppsGetter) WithCliOutput(cliOutput []string) mockAppsGetter {
	m.cliOutput = cliOutput
	return m
}

func newMockAppsGetter() mockAppsGetter {
	return mockAppsGetter{
	}
}

var domains = []string{
	"Getting domains in org myOrg as myUser...",
	"name                                 status   type",
	"domain1.com                          shared",
	"domain2.com                          shared",
	"this.should.be.without.hostname.com  owned",}

func TestGivesBackErrorIfGetAppsFails(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(newMockAppsGetter().WithGetAppsError(expectedError))

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
		plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-dev-CANDIDATE"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	promote := NewPromotePlanner(newMockAppsGetter().WithApps([]plugin_models.GetAppsModel{{Name: candidateAppName}}).WithCliOutput(domains))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
		Space:      "dev",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPromotePlanForAWorkerAppWhenThereIsNoOldApp(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}
	testDomain := "domain.com"

	candidateAppName := createCandidateAppName(application.Name)
	expectedPlan := plan.Plan{
		plan.NewCfCommand("rename", candidateAppName, application.Name),
	}

	promote := NewPromotePlanner(newMockAppsGetter().WithApps([]plugin_models.GetAppsModel{{Name: candidateAppName}}))

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
		plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-dev-CANDIDATE"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	apps := []plugin_models.GetAppsModel{
		{Name: "Ima a app"},
		{Name: "Me 2 lol"},
		{Name: candidateAppName},
	}
	promote := NewPromotePlanner(newMockAppsGetter().WithApps(apps).WithCliOutput(domains))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
		Space:      "dev",
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
		plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-live-CANDIDATE"),
		plan.NewCfCommand("rename", application.Name, "my-app-OLD"),
		plan.NewCfCommand("stop", "my-app-OLD"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	apps := []plugin_models.GetAppsModel{
		{Name: application.Name},
		{Name: candidateAppName},
	}
	promote := NewPromotePlanner(newMockAppsGetter().WithApps(apps).WithCliOutput(domains))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
		Space:      "live",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPromotePlanForAWorkerAppWhenThereIsAnOldApp(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}
	testDomain := "domain.com"

	candidateAppName := createCandidateAppName(application.Name)
	expectedPlan := plan.Plan{
		plan.NewCfCommand("rename", application.Name, "my-app-OLD"),
		plan.NewCfCommand("stop", "my-app-OLD"),
		plan.NewCfCommand("rename", candidateAppName, application.Name),
	}

	apps := []plugin_models.GetAppsModel{
		{Name: application.Name},
		{Name: candidateAppName},
	}
	promote := NewPromotePlanner(newMockAppsGetter().WithApps(apps))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestReturnsErrorFromCliCommandWithoutTerminalOutput(t *testing.T) {
	expectedError := errors.New("Meehp")
	promote := NewPromotePlanner(newMockAppsGetter().WithCliError(expectedError))

	application := manifest.Application{
		Name: "my-app",
		Routes: []string{
			"my-route1.domain1.com",
		},
	}

	_, err := promote.GetPlan(application, Request{})
	assert.Equal(t, expectedError, err)
}

func TestGivesBackAPromotePlanWhenThereIsAnOldAppAndAnEvenOlder(t *testing.T) {
	application := manifest.Application{
		Name: "my-app",
		Routes: []string{
			"my-route1.domain1.com",
			"this.should.be.without.hostname.com",
			"my-route2.domain2.com",
		},
	}
	testDomain := "domain.com"

	candidateAppName := createCandidateAppName(application.Name)
	expectedPlan := plan.Plan{
		plan.NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route1"),
		plan.NewCfCommand("map-route", candidateAppName, "this.should.be.without.hostname.com"),
		plan.NewCfCommand("map-route", candidateAppName, "domain2.com", "-n", "my-route2"),
		plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-yolo-CANDIDATE"),
		plan.NewCfCommand("rename", "my-app-OLD", "my-app-DELETE"),
		plan.NewCfCommand("rename", application.Name, "my-app-OLD"),
		plan.NewCfCommand("stop", "my-app-OLD"),
		plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
	}

	apps := []plugin_models.GetAppsModel{
		{Name: application.Name},
		{Name: candidateAppName},
		{Name: createOldAppName(application.Name)},
	}
	promote := NewPromotePlanner(newMockAppsGetter().WithApps(apps).WithCliOutput(domains))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
		Space:      "yolo",
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestGivesBackAPromotePlanForAWorkerAppWhenThereIsAnOldAppAndAnEvenOlder(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}
	testDomain := "domain.com"

	candidateAppName := createCandidateAppName(application.Name)
	oldAppName := createOldAppName(application.Name)
	deleteAppName := createDeleteName(application.Name, 0)

	expectedPlan := plan.Plan{
		plan.NewCfCommand("rename", oldAppName, deleteAppName),
		plan.NewCfCommand("rename", application.Name, oldAppName),
		plan.NewCfCommand("stop", oldAppName),
		plan.NewCfCommand("rename", candidateAppName, application.Name),
	}

	apps := []plugin_models.GetAppsModel{
		{Name: application.Name},
		{Name: candidateAppName},
		{Name: createOldAppName(application.Name)},
	}
	promote := NewPromotePlanner(newMockAppsGetter().WithApps(apps))

	commands, err := promote.GetPlan(application, Request{
		TestDomain: testDomain,
	})

	assert.Nil(t, err)
	assert.Equal(t, expectedPlan, commands)
}

func TestRenamesOldAppAndDeleteApp(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}

	oldAppName := createOldAppName(application.Name)

	apps := []plugin_models.GetAppsModel{
		{Name: "my-app-OLD"},
		{Name: "my-app-DELETE"},
		{Name: "my-app-DELETE-1"},
	}

	planRename := renameOldAppToDelete(apps, oldAppName, application.Name)

	expectedPlan := []plan.Command{
		plan.NewCfCommand("rename", oldAppName, "my-app-DELETE-2"),
	}
	assert.Equal(t, expectedPlan, planRename)
}

func TestRenamesOldAppWhenNoDeletePresent(t *testing.T) {
	application := manifest.Application{
		Name:    "my-app",
		NoRoute: true,
	}

	oldAppName := createOldAppName(application.Name)

	apps := []plugin_models.GetAppsModel{
		{Name: "my-app-OLD"},
	}

	planRename := renameOldAppToDelete(apps, oldAppName, application.Name)

	expectedPlan := []plan.Command{
		plan.NewCfCommand("rename", oldAppName, "my-app-DELETE"),
	}
	assert.Equal(t, expectedPlan, planRename)
}
