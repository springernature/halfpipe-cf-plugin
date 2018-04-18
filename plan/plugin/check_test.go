package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfCandidateAppNameIsAlreadyInUse(t *testing.T) {

	appName := "app"

	apps := []plugin_models.GetAppsModel{
		{Name: createCandidateAppName("app")},
	}

	check := NewCheck(newMockAppsGetter(apps, nil))

	ok, _ := check.IsCFInAGoodState(appName, "blah", "blah")

	assert.False(t, ok)
}

func TestFailsIfDeleteAppnameIsThere(t *testing.T) {

	appName := "app"

	apps := []plugin_models.GetAppsModel{
		{Name: createDeleteName(appName, 0)},
	}

	check := NewCheck(newMockAppsGetter(apps, nil))

	ok, _ := check.IsCFInAGoodState(appName, "blah", "blah")

	assert.False(t, ok)
}

func TestFailsIfOldAppnameIsRunning(t *testing.T) {

	appName := "app"

	apps := []plugin_models.GetAppsModel{
		{Name: createOldAppName(appName), State: "running"},
	}

	check := NewCheck(newMockAppsGetter(apps, nil))

	ok, _ := check.IsCFInAGoodState(appName, "blah", "blah")

	assert.False(t, ok)
}

func TestFailsIfCandidateRouteIsAlreadyInUse(t *testing.T) {
	appName := "my-app"
	candidateAppName := createCandidateAppName(appName)

	candidateHost := createCandidateHostname(appName, "dev")
	apps := []plugin_models.GetAppsModel{
		{Name: "app1", Routes: []plugin_models.GetAppsRouteSummary{{
			Host:   candidateHost,
			Domain: plugin_models.GetAppsDomainFields{Name: "testdomain.com"},
		}}},
	}

	check := NewCheck(newMockAppsGetter(apps, nil))

	ok, _ := check.IsCFInAGoodState(candidateAppName, "testdomain.com", candidateHost)

	assert.False(t, ok)
}
