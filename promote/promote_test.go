package promote

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"errors"
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin"
	"github.com/springernature/halfpipe-cf-plugin/command"
	"github.com/springernature/halfpipe-cf-plugin/helpers"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/springernature/halfpipe-cf-plugin/plan"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReturnsErrorIfGetCurrentSpaceFails(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(helpers.NewMockCliConnection().WithSpaceError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, halfpipe_cf_plugin.Request{})

	assert.Equal(t, halfpipe_cf_plugin.ErrGetCurrentSpace(expectedError), err)
}


func TestReturnsErrorIfCandidateAppNotFound(t *testing.T) {
	applicationName := "kehe"
	expectedError := errors.New("error")
	promote := NewPromotePlanner(helpers.NewMockCliConnection().WithGetAppError(expectedError))

	_, err := promote.GetPlan(manifest.Application{Name: applicationName}, halfpipe_cf_plugin.Request{})

	assert.Equal(t, halfpipe_cf_plugin.ErrGetApp(applicationName, expectedError), err)
}

func TestReturnsErrorIfCandidateAppIsNotRunning(t *testing.T) {
	promote := NewPromotePlanner(helpers.NewMockCliConnection().WithApp(plugin_models.GetAppModel{
		Name:  "myApp-CANDIDATE",
		State: "stopped",
	}))

	_, err := promote.GetPlan(manifest.Application{}, halfpipe_cf_plugin.Request{})

	assert.Equal(t, ErrCandidateNotRunning, err)
}

func TestReturnsErrorIfGetAppsErrorsOut(t *testing.T) {
	expectedError := errors.New("mehp")

	promote := NewPromotePlanner(helpers.NewMockCliConnection().
		WithApp(plugin_models.GetAppModel{
		Name:  "myApp-CANDIDATE",
		State: "started",
	}).
		WithGetAppsError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, halfpipe_cf_plugin.Request{})
	assert.Equal(t, halfpipe_cf_plugin.ErrGetApps(expectedError), err)
}

func TestWorkerApp(t *testing.T) {
	t.Run("No previously deployed version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}))

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed stopped version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
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

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
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

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
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

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", helpers.CreateOldAppName(man.Name), helpers.CreateDeleteName(man.Name, 0)),
			command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version and a uncleaned up DELETE app", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
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
			{
				Name:  "myApp-DELETE",
				State: "stopped",
			},
		}))

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", helpers.CreateOldAppName(man.Name), helpers.CreateDeleteName(man.Name, 1)),
			command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version and a couple of uncleaned DELETE apps", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
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
			{
				Name:  "myApp-DELETE",
				State: "stopped",
			},
			{
				Name:  "myApp-DELETE-1",
				State: "stopped",
			},

			{
				Name:  "myApp-DELETE-2",
				State: "stopped",
			},

		}))

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", helpers.CreateOldAppName(man.Name), helpers.CreateDeleteName(man.Name, 3)),
			command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

}

func TestAppWithRoute(t *testing.T) {
	appName := "myApp"
	testDomain := "test.com"
	space := "dev"
	appCandidateHostname := fmt.Sprintf("myApp-%s-CANDIDATE", space)


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

	man := manifest.Application{
		Name: appName,
		Routes: []manifest.Route{
			{Route: route1},
			{Route: route2},
			{Route: route3},
			{Route: route4},
		},
	}

	candidateApp := plugin_models.GetAppModel{
		Name:  helpers.CreateCandidateAppName(appName),
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

	request := halfpipe_cf_plugin.Request{
		TestDomain: testDomain,
	}

	t.Run("Errors out if we cannot get domains in org", func(t *testing.T) {
		expectedError := errors.New("meeehp")
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithApp(candidateApp).
			WithCliError(expectedError))

		_, err := promote.GetPlan(man, request)
		assert.Equal(t, halfpipe_cf_plugin.ErrCliCommandWithoutTerminalOutput("cf domains", expectedError), err)
	})

	t.Run("No previously deployed version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithSpace(space).
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: helpers.CreateCandidateAppName(appName)},
		}))

		expectedPlan := plan.Plan{
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route1Domain, "-n", route1Host),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route2),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route4Domain, "--path", route4Path),
			command.NewCfShellCommand("unmap-route", helpers.CreateCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(man, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithSpace(space).
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: helpers.CreateCandidateAppName(appName), State: "started"},
			{Name: appName, State: "started"},
		}))

		expectedPlan := plan.Plan{
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route1Domain, "-n", route1Host),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route2),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route4Domain, "--path", route4Path),
			command.NewCfShellCommand("unmap-route", helpers.CreateCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			command.NewCfShellCommand("rename", appName, helpers.CreateOldAppName(appName)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(appName)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(man, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version and a stopped older version", func(t *testing.T) {
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithSpace(space).
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{
				Name:  helpers.CreateCandidateAppName(appName),
				State: "started",
			},
			{
				Name:  helpers.CreateOldAppName(appName),
				State: "stopped",
			},
			{
				Name:  appName,
				State: "started",
			},
		}))

		expectedPlan := plan.Plan{
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route1Domain, "-n", route1Host),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route2),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route4Domain, "--path", route4Path),
			command.NewCfShellCommand("unmap-route", helpers.CreateCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			command.NewCfShellCommand("rename", helpers.CreateOldAppName(appName), helpers.CreateDeleteName(appName, 0)),
			command.NewCfShellCommand("rename", appName, helpers.CreateOldAppName(appName)),
			command.NewCfShellCommand("stop", helpers.CreateOldAppName(appName)),
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(man, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})
}

func TestAppWithRouteWhenPreviousPromoteFailure(t *testing.T) {
	appName := "myApp"
	testDomain := "test.com"
	space := "dev"
	appCandidateHostname := fmt.Sprintf("myApp-%s-CANDIDATE", space)

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

	man := manifest.Application{
		Name: appName,
		Routes: []manifest.Route{
			{Route: route1},
			{Route: route2},
			{Route: route3},
			{Route: route4},
		},
	}

	candidateApp := plugin_models.GetAppModel{
		Name:  helpers.CreateCandidateAppName(appName),
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

	request := halfpipe_cf_plugin.Request{
		TestDomain: testDomain,
	}

	t.Run("No previously deployed version", func(t *testing.T) {
		/*
		Normal run would look like
				$ cf map-route appName-CANDIDATE domain.com, -n myRoute [1]
				$ cf map-route appName-CANDIDATE route2 bindToDomain.com [2]
				$ cf map-route appName-CANDIDATE domain.com, -n myRouteWithPath, --path yolo [3]
				$ cf map-route appName-CANDIDATE domain.com --path kehe/keho [4]
				$ cf unmap-route appName-CANDIDATE test.com -n myApp-dev-CANDIDATE [5]
				$ cf rename appName-CANDIDATE appName [6]
		 */

		t.Run("previous promote failed at step [4]", func(t *testing.T) {
			candidateApp.Routes = []plugin_models.GetApp_RouteSummary{
				{
					Host: appCandidateHostname,
					Domain: plugin_models.GetApp_DomainFields{
						Name: testDomain,
					},
				}, // From halfpipe-promote
				{
					Host: "myRoute",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[1]
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "bindToDomain.com",
					},
				}, //[2]
				{
					Host: "myRouteWithPath",
					Path: "yolo",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[3]
			}

			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithSpace(space).
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: helpers.CreateCandidateAppName(appName)},
			}))

			expectedPlan := plan.Plan{
				command.NewCfShellCommand("map-route", helpers.CreateCandidateAppName(appName), route4Domain, "--path", route4Path),
				command.NewCfShellCommand("unmap-route", helpers.CreateCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
			}

			plan, err := promote.GetPlan(man, request)
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [5]", func(t *testing.T) {
			candidateApp.Routes = []plugin_models.GetApp_RouteSummary{
				{
					Host: appCandidateHostname,
					Domain: plugin_models.GetApp_DomainFields{
						Name: testDomain,
					},
				}, // From halfpipe-promote
				{
					Host: "myRoute",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[1]
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "bindToDomain.com",
					},
				}, //[2]
				{
					Host: "myRouteWithPath",
					Path: "yolo",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[3]
				{
					Host: "",
					Path: "kehe/keho",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[4]
			}

			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithSpace(space).
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: helpers.CreateCandidateAppName(appName)},
			}))

			expectedPlan := plan.Plan{
				command.NewCfShellCommand("unmap-route", helpers.CreateCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
			}

			plan, err := promote.GetPlan(man, request)
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [6]", func(t *testing.T) {
			// Notice that there is no test-route, so [5] ran last promote run before failing at [6]
			candidateApp.Routes = []plugin_models.GetApp_RouteSummary{
				{
					Host: "myRoute",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[1]
				{
					Domain: plugin_models.GetApp_DomainFields{
						Name: "bindToDomain.com",
					},
				}, //[2]
				{
					Host: "myRouteWithPath",
					Path: "yolo",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[3]
				{
					Host: "",
					Path: "kehe/keho",
					Domain: plugin_models.GetApp_DomainFields{
						Name: "domain.com",
					},
				}, //[4]
			}
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: helpers.CreateCandidateAppName(appName)},
			}))

			expectedPlan := plan.Plan{
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(appName), appName),
			}

			plan, err := promote.GetPlan(man, request)
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

	})

}

func TestWorkerAppWithPreviousPromoteFailure(t *testing.T) {
	/* In TestAppWithRouteWhenPreviousPromoteFailure we test all the mapping and unmapping.
    These will always happen before the stopping and renaming of older apps thus we can
	test with worker app here as its easier to setup tests.
	*/

	t.Run("No previously deployed version", func(t *testing.T) {
		/*
		Normal run would look like
		$ cf rename appName-CANDIDATE appName [1]
		*/

		//
		promote := NewPromotePlanner(helpers.NewMockCliConnection().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}).
			WithApps([]plugin_models.GetAppsModel{
			{Name: "myApp-CANDIDATE", State: "started"},
		}))

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := plan.Plan{
			command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously running deployed version", func(t *testing.T) {
		/*
		Normal run would look like
		$ cf rename appName appName-OLD [1]
		$ cf stop appName-OLD [2]
		$ cf rename appName-CANDIDATE appName [3]
		*/

		t.Run("previous promote failed at step [2]", func(t *testing.T) {
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(plugin_models.GetAppModel{
				Name:  "myApp-CANDIDATE",
				State: "started",
			}).
				WithApps([]plugin_models.GetAppsModel{
				{Name: "myApp-CANDIDATE", State: "started"},
				{Name: "myApp-OLD", State: "started"},
			}))

			man := manifest.Application{
				Name:    "myApp",
				NoRoute: true,
			}
			expectedPlan := plan.Plan{
				command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [3]", func(t *testing.T) {
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(plugin_models.GetAppModel{
				Name:  "myApp-CANDIDATE",
				State: "started",
			}).
				WithApps([]plugin_models.GetAppsModel{
				{Name: "myApp-CANDIDATE", State: "started"},
				{Name: "myApp-OLD", State: "stopped"},
			}))

			man := manifest.Application{
				Name:    "myApp",
				NoRoute: true,
			}
			expectedPlan := plan.Plan{
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

	})

	t.Run("One previously deployed started version with an stopped old version", func(t *testing.T) {
		/*
		Normal run would look like
		$ cf rename appName-OLD appName-DELETE [1]
		$ cf rename appName appName-OLD [2]
		$ cf stop appName-OLD [3]
		$ cf rename appName-CANDIDATE appName [4]
		*/

		t.Run("previous promote failed at step [2]", func(t *testing.T) {
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(plugin_models.GetAppModel{
				Name:  "myApp-CANDIDATE",
				State: "started",
			}).
				WithApps([]plugin_models.GetAppsModel{
				{Name: "myApp-CANDIDATE", State: "started"},
				{Name: "myApp-DELETE", State: "stopped"},
				{Name: "myApp", State: "started"},
			}))

			man := manifest.Application{
				Name:    "myApp",
				NoRoute: true,
			}
			expectedPlan := plan.Plan{
				command.NewCfShellCommand("rename", man.Name, helpers.CreateOldAppName(man.Name)),
				command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [3]", func(t *testing.T) {
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(plugin_models.GetAppModel{
				Name:  "myApp-CANDIDATE",
				State: "started",
			}).
				WithApps([]plugin_models.GetAppsModel{
				{Name: "myApp-CANDIDATE", State: "started"},
				{Name: "myApp-DELETE", State: "stopped"},
				{Name: "myApp-OLD", State: "started"},
			}))

			man := manifest.Application{
				Name:    "myApp",
				NoRoute: true,
			}
			expectedPlan := plan.Plan{
				command.NewCfShellCommand("stop", helpers.CreateOldAppName(man.Name)),
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [4]", func(t *testing.T) {
			promote := NewPromotePlanner(helpers.NewMockCliConnection().
				WithApp(plugin_models.GetAppModel{
				Name:  "myApp-CANDIDATE",
				State: "started",
			}).
				WithApps([]plugin_models.GetAppsModel{
				{Name: "myApp-CANDIDATE", State: "started"},
				{Name: "myApp-DELETE", State: "stopped"},
				{Name: "myApp-OLD", State: "stopped"},
			}))

			man := manifest.Application{
				Name:    "myApp",
				NoRoute: true,
			}
			expectedPlan := plan.Plan{
				command.NewCfShellCommand("rename", helpers.CreateCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, halfpipe_cf_plugin.Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

	})

}
