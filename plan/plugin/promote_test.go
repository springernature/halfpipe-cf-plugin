package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
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

func (m mockAppsGetter) WithAppError(err error) mockAppsGetter {
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

var domains = []string{
	"Getting domains in org myOrg as myUser...",
	"name                                 status   type",
	"domain1.com                          shared",
	"domain2.com                          shared",
	"this.should.be.without.hostname.com  owned",}

func TestPromote(t *testing.T) {
	t.Run("Gives back error if get apps fails", func(t *testing.T) {
		expectedError := errors.New("error")
		promote := NewPromotePlanner(newMockAppsGetter().WithGetAppsError(expectedError))

		_, err := promote.GetPlan(manifest.Application{
			NoRoute: true,
		}, Request{})

		assert.Equal(t, expectedError, err)
	})

	t.Run("Gives back a promote plan when there is no old app", func(t *testing.T) {
		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
				{"my-route2.domain2.com"},
				{"my-route3.domain3.com/path"},
				{"this.should.be.without.hostname.com"},
				{"this.should.be.without.hostname.com/some/other/path"},
			},
		}
		testDomain := "domain.com"

		candidateAppName := createCandidateAppName(application.Name)
		expectedPlan := plan.Plan{
			plan.NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route1"),
			plan.NewCfCommand("map-route", candidateAppName, "domain2.com", "-n", "my-route2"),
			plan.NewCfCommand("map-route", candidateAppName, "domain3.com", "-n", "my-route3", "--path", "path"),
			plan.NewCfCommand("map-route", candidateAppName, "this.should.be.without.hostname.com"),
			plan.NewCfCommand("map-route", candidateAppName, "this.should.be.without.hostname.com", "--path", "some/other/path"),
			plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-dev-CANDIDATE"),
			plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
		}

		promote := NewPromotePlanner(newMockAppsGetter().
			WithApps([]plugin_models.GetAppsModel{{Name: candidateAppName}}).
			WithApp(plugin_models.GetAppModel{Routes: []plugin_models.GetApp_RouteSummary{{Host: "my-app-dev-CANDIDATE", Domain:plugin_models.GetApp_DomainFields{Name:testDomain}}}}).
			WithCliOutput(domains))

		commands, err := promote.GetPlan(application, Request{
			TestDomain: testDomain,
			Space:      "dev",
		})

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, commands)
	})

	t.Run("Gives back a promote plan for a worker app when there is no old app", func(t *testing.T) {
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
	})

	t.Run("Give back a promote plan when there is an old app but with a different name", func(t *testing.T) {
		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
				{"my-route2.domain2.com"},
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
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApps(apps).
			WithApp(plugin_models.GetAppModel{Routes: []plugin_models.GetApp_RouteSummary{{Host: "my-app-dev-CANDIDATE", Domain:plugin_models.GetApp_DomainFields{Name:testDomain}}}}).
			WithCliOutput(domains))

		commands, err := promote.GetPlan(application, Request{
			TestDomain: testDomain,
			Space:      "dev",
		})

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, commands)
	})

	t.Run("Gives back a promote plan when there is an old app", func(t *testing.T) {
		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
				{"my-route2.domain2.com"},
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
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApps(apps).
			WithApp(plugin_models.GetAppModel{Routes: []plugin_models.GetApp_RouteSummary{{Host: "my-app-live-CANDIDATE", Domain:plugin_models.GetApp_DomainFields{Name:testDomain}}}}).
			WithCliOutput(domains))

		commands, err := promote.GetPlan(application, Request{
			TestDomain: testDomain,
			Space:      "live",
		})

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, commands)
	})

	t.Run("Gives back a promote plan for a worker app when there is an old app", func(t *testing.T) {
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
	})

	t.Run("Returns error from CliCommand without terminal output", func(t *testing.T) {
		expectedError := errors.New("Meehp")
		promote := NewPromotePlanner(newMockAppsGetter().WithGetAppsError(expectedError))

		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
			},
		}

		_, err := promote.GetPlan(application, Request{})
		assert.Equal(t, expectedError, err)
	})

	t.Run("Gives back a promote plan when there is an old app and an even older app", func(t *testing.T) {
		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
				{"this.should.be.without.hostname.com"},
				{"my-route2.domain2.com"},
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
		promote := NewPromotePlanner(newMockAppsGetter().
			WithApps(apps).
			WithApp(plugin_models.GetAppModel{Routes: []plugin_models.GetApp_RouteSummary{{Host: "my-app-yolo-CANDIDATE", Domain:plugin_models.GetApp_DomainFields{Name:testDomain}}}}).
			WithCliOutput(domains))

		commands, err := promote.GetPlan(application, Request{
			TestDomain: testDomain,
			Space:      "yolo",
		})

		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, commands)
	})

	t.Run("Gives back a promote plan for a worker app then there is an old app and an even older", func(t *testing.T) {
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
	})

	t.Run("Test renames old app and deleted app", func(t *testing.T) {
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
	})

	t.Run("Test renames old app when no delete present", func(t *testing.T) {
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
	})
}

func TestPromoteFromAPreviousPromoteFailure(t *testing.T) {
	t.Run("Candidate is not there should result in a error", func(t *testing.T) {
		application := manifest.Application{
			Name: "my-app",
			Routes: []manifest.Route{
				{"my-route1.domain1.com"},
				{"my-route2.domain1.com"},
			},
		}
		expectedError := errors.New("Mehp")

		promote := NewPromotePlanner(newMockAppsGetter().WithAppError(expectedError))

		_, err := promote.GetPlan(application, Request{})
		assert.Equal(t, expectedError, err)
	})

	t.Run("No previous app deployed", func(t *testing.T) {

		/*
		Normal run would look like
		[1] cf map-route my-app-CANDIDATE domain1.com -n my-route1
		[2] cf map-route my-app-CANDIDATE domain2.com
		[3] cf map-route my-app-CANDIDATE domain1.com -n my-route-path --path yolo1
		[4] cf map-route my-app-CANDIDATE domain1.com -n my-route2
		[5] cf map-route my-app-CANDIDATE this.should.be.without.hostname.com
		[6] cf map-route my-app-CANDIDATE domain1.com -n my-route-path --path yolo2
		[7] cf unmap-route my-app-CANDIDATE domain.com -n my-app-dev-CANDIDATE
		[8] cf rename halfpipe-example-nodejs-CANDIDATE halfpipe-example-nodejs
		 */

		t.Run("previous run failed on step 4", func(t *testing.T) {
			application := manifest.Application{
				Name: "my-app",
				Routes: []manifest.Route{
					{"my-route1.domain1.com"},
					{"domain2.com"},
					{"my-route-path.domain1.com/yolo1"},
					{"my-route2.domain1.com"},
					{"this.should.be.without.hostname.com"},
					{"my-route-path.domain1.com/yolo2"},
				},
			}
			testDomain := "domain.com"

			candidateAppName := createCandidateAppName(application.Name)
			expectedPlan := plan.Plan{
				plan.NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route2"),
				plan.NewCfCommand("map-route", candidateAppName, "this.should.be.without.hostname.com"),
				plan.NewCfCommand("map-route", candidateAppName, "domain1.com", "-n", "my-route-path", "--path", "yolo2"),
				plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-dev-CANDIDATE"),
				plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
			}

			app := plugin_models.GetAppModel{
				Name: candidateAppName,
				Routes: []plugin_models.GetApp_RouteSummary{
					{
						Host: "my-app-dev-CANDIDATE",
						Domain: plugin_models.GetApp_DomainFields{
							Name: testDomain,
						},
					},
					{
						Host: "my-route1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain2.com",
						},
					},
					{
						Host: "my-route-path",
						Path: "yolo1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
				},
			}

			promote := NewPromotePlanner(newMockAppsGetter().WithApp(app).WithCliOutput(domains))

			commands, err := promote.GetPlan(application, Request{
				TestDomain: testDomain,
				Space:      "dev",
			})

			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, commands)
		})

		t.Run("previous run failed on step 7", func(t *testing.T) {
			application := manifest.Application{
				Name: "my-app",
				Routes: []manifest.Route{
					{"my-route1.domain1.com"},
					{"domain2.com"},
					{"my-route-path.domain1.com/yolo1"},
					{"my-route2.domain1.com"},
					{"this.should.be.without.hostname.com"},
					{"my-route-path.domain1.com/yolo2"},
				},
			}
			testDomain := "domain.com"

			candidateAppName := createCandidateAppName(application.Name)

			app := plugin_models.GetAppModel{
				Name: candidateAppName,
				Routes: []plugin_models.GetApp_RouteSummary{
					{
						Host: "my-app-dev-CANDIDATE",
						Domain: plugin_models.GetApp_DomainFields{
							Name: testDomain,
						},
					},
					{
						Host: "my-route1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain2.com",
						},
					},
					{
						Host: "my-route2",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "this.should.be.without.hostname.com",
						},
					},
					{
						Host: "my-route-path",
						Path: "yolo1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "my-route-path",
						Path: "yolo2",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
				},
			}

			expectedPlan := plan.Plan{
				plan.NewCfCommand("unmap-route", candidateAppName, testDomain, "-n", "my-app-dev-CANDIDATE"),
				plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
			}

			promote := NewPromotePlanner(newMockAppsGetter().WithApp(app).WithCliOutput(domains))

			commands, err := promote.GetPlan(application, Request{
				TestDomain: testDomain,
				Space:      "dev",
			})

			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, commands)
		})

		t.Run("previous run failed on step 8", func(t *testing.T) {
			application := manifest.Application{
				Name: "my-app",
				Routes: []manifest.Route{
					{"my-route1.domain1.com"},
					{"domain2.com"},
					{"my-route-path.domain1.com/yolo1"},
					{"my-route2.domain1.com"},
					{"this.should.be.without.hostname.com"},
					{"my-route-path.domain1.com/yolo2"},
				},
			}
			testDomain := "domain.com"

			candidateAppName := createCandidateAppName(application.Name)
			expectedPlan := plan.Plan{
				plan.NewCfCommand("rename", "my-app-CANDIDATE", application.Name),
			}

			app := plugin_models.GetAppModel{
				Name: candidateAppName,
				Routes: []plugin_models.GetApp_RouteSummary{
					{
						Host: "my-route1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain2.com",
						},
					},
					{
						Host: "my-route2",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "this.should.be.without.hostname.com",
						},
					},
					{
						Host: "my-route-path",
						Path: "yolo1",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
					{
						Host: "my-route-path",
						Path: "yolo2",
						Domain: plugin_models.GetApp_DomainFields{
							Name: "domain1.com",
						},
					},
				},
			}
			promote := NewPromotePlanner(newMockAppsGetter().WithApp(app).WithCliOutput(domains))

			commands, err := promote.GetPlan(application, Request{
				TestDomain: testDomain,
				Space:      "dev",
			})

			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, commands)
		})
	})
}
