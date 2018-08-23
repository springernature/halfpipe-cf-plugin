package plan

import (
	"code.cloudfoundry.org/cli/plugin/models"
	"testing"
	"github.com/stretchr/testify/assert"
	"errors"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"fmt"
)

func TestReturnsErrorIfGetCurrentSpaceFails(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(newMockCliConnection().WithSpaceError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}


func TestReturnsErrorIfCandidateAppNotFound(t *testing.T) {
	expectedError := errors.New("error")
	promote := NewPromotePlanner(newMockCliConnection().WithGetAppError(expectedError))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfCandidateAppIsNotRunning(t *testing.T) {
	promote := NewPromotePlanner(newMockCliConnection().WithApp(plugin_models.GetAppModel{
		Name:  "myApp-CANDIDATE",
		State: "stopped",
	}))

	_, err := promote.GetPlan(manifest.Application{}, Request{})

	assert.Equal(t, ErrCandidateNotRunning, err)
}

func TestReturnsErrorIfGetAppsErrorsOut(t *testing.T) {
	expectedError := errors.New("mehp")

	promote := NewPromotePlanner(newMockCliConnection().
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
		promote := NewPromotePlanner(newMockCliConnection().
			WithApp(plugin_models.GetAppModel{
			Name:  "myApp-CANDIDATE",
			State: "started",
		}))

		man := manifest.Application{
			Name:    "myApp",
			NoRoute: true,
		}
		expectedPlan := Plan{
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed stopped version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
			NewCfCommand("stop", createOldAppName(man.Name)),
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", createOldAppName(man.Name), createDeleteName(man.Name, 0)),
			NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
			NewCfCommand("stop", createOldAppName(man.Name)),
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version and a uncleaned up DELETE app", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", createOldAppName(man.Name), createDeleteName(man.Name, 1)),
			NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
			NewCfCommand("stop", createOldAppName(man.Name)),
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started version with an stopped old version and a couple of uncleaned DELETE apps", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", createOldAppName(man.Name), createDeleteName(man.Name, 3)),
			NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
			NewCfCommand("stop", createOldAppName(man.Name)),
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
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
	}

	t.Run("Errors out if we cannot get domains in org", func(t *testing.T) {
		expectedError := errors.New("meeehp")
		promote := NewPromotePlanner(newMockCliConnection().
			WithApp(candidateApp).
			WithCliError(expectedError))

		_, err := promote.GetPlan(man, request)
		assert.Equal(t, expectedError, err)
	})

	t.Run("No previously deployed version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
			WithSpace(space).
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: createCandidateAppName(appName)},
		}))

		expectedPlan := Plan{
			NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			NewCfCommand("map-route", createCandidateAppName(appName), route2),
			NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			NewCfCommand("rename", createCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(man, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
			WithSpace(space).
			WithApp(candidateApp).
			WithCliOutput(cfDomains).
			WithApps([]plugin_models.GetAppsModel{
			{Name: createCandidateAppName(appName), State: "started"},
			{Name: appName, State: "started"},
		}))

		expectedPlan := Plan{
			NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			NewCfCommand("map-route", createCandidateAppName(appName), route2),
			NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			NewCfCommand("rename", appName, createOldAppName(appName)),
			NewCfCommand("stop", createOldAppName(appName)),
			NewCfCommand("rename", createCandidateAppName(appName), appName),
		}

		plan, err := promote.GetPlan(man, request)
		assert.Nil(t, err)
		assert.Equal(t, expectedPlan, plan)
	})

	t.Run("One previously deployed started live version and a stopped older version", func(t *testing.T) {
		promote := NewPromotePlanner(newMockCliConnection().
			WithSpace(space).
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

		expectedPlan := Plan{
			NewCfCommand("map-route", createCandidateAppName(appName), route1Domain, "-n", route1Host),
			NewCfCommand("map-route", createCandidateAppName(appName), route2),
			NewCfCommand("map-route", createCandidateAppName(appName), route3Domain, "-n", route3Host, "--path", route3Path),
			NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
			NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
			NewCfCommand("rename", createOldAppName(appName), createDeleteName(appName, 0)),
			NewCfCommand("rename", appName, createOldAppName(appName)),
			NewCfCommand("stop", createOldAppName(appName)),
			NewCfCommand("rename", createCandidateAppName(appName), appName),
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

			promote := NewPromotePlanner(newMockCliConnection().
				WithSpace(space).
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: createCandidateAppName(appName)},
			}))

			expectedPlan := Plan{
				NewCfCommand("map-route", createCandidateAppName(appName), route4Domain, "--path", route4Path),
				NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
				NewCfCommand("rename", createCandidateAppName(appName), appName),
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

			promote := NewPromotePlanner(newMockCliConnection().
				WithSpace(space).
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: createCandidateAppName(appName)},
			}))

			expectedPlan := Plan{
				NewCfCommand("unmap-route", createCandidateAppName(appName), testDomain, "-n", appCandidateHostname),
				NewCfCommand("rename", createCandidateAppName(appName), appName),
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
			promote := NewPromotePlanner(newMockCliConnection().
				WithApp(candidateApp).
				WithCliOutput(cfDomains).
				WithApps([]plugin_models.GetAppsModel{
				{Name: createCandidateAppName(appName)},
			}))

			expectedPlan := Plan{
				NewCfCommand("rename", createCandidateAppName(appName), appName),
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
		promote := NewPromotePlanner(newMockCliConnection().
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
		expectedPlan := Plan{
			NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
		}

		plan, err := promote.GetPlan(man, Request{})
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
			promote := NewPromotePlanner(newMockCliConnection().
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
			expectedPlan := Plan{
				NewCfCommand("stop", createOldAppName(man.Name)),
				NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [3]", func(t *testing.T) {
			promote := NewPromotePlanner(newMockCliConnection().
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
			expectedPlan := Plan{
				NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, Request{})
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
			promote := NewPromotePlanner(newMockCliConnection().
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
			expectedPlan := Plan{
				NewCfCommand("rename", man.Name, createOldAppName(man.Name)),
				NewCfCommand("stop", createOldAppName(man.Name)),
				NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [3]", func(t *testing.T) {
			promote := NewPromotePlanner(newMockCliConnection().
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
			expectedPlan := Plan{
				NewCfCommand("stop", createOldAppName(man.Name)),
				NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

		t.Run("previous promote failed at step [4]", func(t *testing.T) {
			promote := NewPromotePlanner(newMockCliConnection().
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
			expectedPlan := Plan{
				NewCfCommand("rename", createCandidateAppName(man.Name), man.Name),
			}

			plan, err := promote.GetPlan(man, Request{})
			assert.Nil(t, err)
			assert.Equal(t, expectedPlan, plan)
		})

	})

}
