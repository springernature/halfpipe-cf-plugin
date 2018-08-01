package plugin

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"testing"
	"github.com/stretchr/testify/assert"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"fmt"
)

type mockAppsGetter struct {
	apps      []plugin_models.GetAppsModel
	appsError error
	app       plugin_models.GetAppModel
	appError  error
	cliError  error
	cliOutput []string
}

func (m mockAppsGetter) GetApp(appName string) (plugin_models.GetAppModel, error) {
	return m.app, m.appError
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

func (m mockAppsGetter) WithApp(app plugin_models.GetAppModel) mockAppsGetter {
	m.app = app
	return m
}

func (m mockAppsGetter) WithGetAppError(err error) mockAppsGetter {
	m.appError = err
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

func TestReturnsErrorIfCandidateAppNotFound(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(newMockAppsGetter().WithGetAppError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfCandidateAppIsNotRunning(t *testing.T) {
	promote := NewPromotePlanner(newMockAppsGetter().WithApp(plugin_models.GetAppModel{
		Name:  "myApp-CANDIDATE",
		State: "stopped",
	}))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, ErrCandidateNotRunning, err)
}

func TestReturnsErrorIfGetAppsErrorsOut(t *testing.T) {
	expectedError := errors.New("Mehp")

	promote := NewPromotePlanner(newMockAppsGetter().
		WithApp(plugin_models.GetAppModel{
		Name:  "myApp-CANDIDATE",
		State: "started",
	}).
		WithGetAppsError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, Request{})
	assert.Equal(t, expectedError, err)
}

func TestWorkerApp(t *testing.T) {
	t.Run("No previously deployed version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}))

		manifest := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			plan.NewCfCommand("rename", createCandidateAppName(manifest.Name), manifest.Name),
		}

		plan, err := promote.GetPlan(manifest, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed stopped version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}).
			WithApps([]plugin_models.GetAppsModel{
			{
				Name:  "myApp",
				State: "stopped",
			},
		}))

		manifest := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			plan.NewCfCommand("rename", manifest.Name, createOldAppName(manifest.Name)),
			plan.NewCfCommand("rename", createCandidateAppName(manifest.Name), manifest.Name),
		}

		plan, err := promote.GetPlan(manifest, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}).
			WithApps([]plugin_models.GetAppsModel{
			{
				Name:  "myApp",
				State: "started",
			},
		}))

		manifest := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			plan.NewCfCommand("rename", manifest.Name, createOldAppName(manifest.Name)),
			plan.NewCfCommand("stop", createOldAppName(manifest.Name)),
			plan.NewCfCommand("rename", createCandidateAppName(manifest.Name), manifest.Name),
		}

		plan, err := promote.GetPlan(manifest, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}).
			WithApps([]plugin_models.GetAppsModel{
			{
				Name:  "myApp",
				State: "started",
			},
			{
				Name:  "myApp-OLD",
				State: "stopped",
			},
		}))

		manifest := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			plan.NewCfCommand("rename", createOldAppName(manifest.Name), createDeleteName(manifest.Name, 0)),
			plan.NewCfCommand("rename", manifest.Name, createOldAppName(manifest.Name)),
			plan.NewCfCommand("stop", createOldAppName(manifest.Name)),
			plan.NewCfCommand("rename", createCandidateAppName(manifest.Name), manifest.Name),
		}

		plan, err := promote.GetPlan(manifest, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})
}

func TestAppWithRoute(t *testing.T) {
	appName := "myApp"
	testDomain := "test.com"
	appCandidateHostname := "myApp-dev-CANDIDATE"
	space := "dev"

	cfDomains := []string{
		"Getting domains in org myOrg as myUser...",
		"name                                 status   type",
		"test.com                             shared",
		"domain.com                           shared",
		"bindToDomain.com                     owned",
	}

	route1Host := "myRoute"
	route1Domain := "domain.com"
	route1 := fmt.Sprintf("%s.%s", route1Host, route1Domain)

	route2 := "bindToDomain.com"

	route3Host := "myRouteWithPath"
	route3Domain := "domain.com"
	route3Path := "yolo"
	route3 := fmt.Sprintf("%s.%s/%s", route3Host, route3Domain, route3Path)

	route4Domain := "domain.com"
	route4Path := "kehe/keho"
	route4 := fmt.Sprintf("%s/%s", route4Domain, route4Path)

	manifest := manifest.Application{
		Name: appName,
		Routes: []manifest.Route{
			{Route: route1},
			{Route: route2},
			{Route: route3},
			{Route: route4},
		},
	}

	candidateApp := plugin_models.GetAppModel{
		Name:  createCandidateAppName(appName),
		State: "started",
		Routes: []plugin_models.GetApp_RouteSummary{
			{
				Host: appCandidateHostname,
				Domain: plugin_models.GetApp_DomainFields{
					Name: testDomain,
				},
			},
		},
	}

	request := Request{
		TestDomain: testDomain,
		Space:      space,
	}

	t.Run("Errors out if we cannot get domains in org", func(t *testing.T) {
		expectedError := errors.New("Meeehp")
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(candidateApp).
			WithCliError(expectedError))

		_, err := promote.GetPlan(manifest, request)
		assert.Equal(t, expectedError, err)
	})

	t.Run("No previously deployed version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: createCandidateAppName(appName)},
		}))

		expectedPlan := plan.Plan{
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route2),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			plan.NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			plan.NewCfCommand("rename", createCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(manifest, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: createCandidateAppName(appName), State: "started"},
			{Name: appName, State: "started"},
		}))

		expectedPlan := plan.Plan{
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route2),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			plan.NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			plan.NewCfCommand("rename", appName, createOldAppName(appName)),
			plan.NewCfCommand("stop", createOldAppName(appName)),
			plan.NewCfCommand("rename", createCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(manifest, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version and a stopped older version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{
				Name:  createCandidateAppName(appName),
				State: "started",
			},
			{
				Name:  createOldAppName(appName),
				State: "stopped",
			},
			{
				Name:  appName,
				State: "started",
			},
		}))

		expectedPlan := plan.Plan{
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route2),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			plan.NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			plan.NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			plan.NewCfCommand("rename", createOldAppName(appName), createDeleteName(appName, 0)),
			plan.NewCfCommand("rename", appName, createOldAppName(appName)),
			plan.NewCfCommand("stop", createOldAppName(appName)),
			plan.NewCfCommand("rename", createCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(manifest, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})
}
