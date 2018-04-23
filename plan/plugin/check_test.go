package plugin

import (
	"testing"

	"code.cloudfoundry.org/cli/plugin/models"
	"github.com/stretchr/testify/assert"
)

func TestFailsIfCandidateAppNameIsAlreadyInUse(t *testing.T) {

	appName := "app"

	candidateAppName := createCandidateAppName("app")
	apps := []plugin_models.GetAppsModel{
		{Name: candidateAppName},
	}

	err := checkCFState(appName, "blah", "blah", newMockAppsGetter(apps, nil))

	assert.Equal(t, err, ErrAppNameExists(candidateAppName))
}

func TestFailsIfDeleteAppNameIsThere(t *testing.T) {
	appName := "app"
	deleteAppName := createDeleteName(appName, 0)
	apps := []plugin_models.GetAppsModel{
		{Name: deleteAppName},
	}

	err := checkCFState(appName, "blah", "blah", newMockAppsGetter(apps, nil))

	assert.Equal(t, err, ErrAppNameExists(deleteAppName))
}

func TestFailsIfOldAppIsRunning(t *testing.T) {
	appName := "app"

	oldAppName := createOldAppName(appName)
	apps := []plugin_models.GetAppsModel{
		{Name: oldAppName, State: "running"},
	}

	err := checkCFState(appName, "blah", "blah", newMockAppsGetter(apps, nil))

	assert.Equal(t, err, ErrAppRunning(oldAppName))
}

func TestFailsIfCandidateRouteIsAlreadyInUse(t *testing.T) {
	appName := "my-app"
	candidateHost := createCandidateHostname(appName, "dev")

	apps := []plugin_models.GetAppsModel{
		{Name: "app1", Routes: []plugin_models.GetAppsRouteSummary{{
			Host:   candidateHost,
			Domain: plugin_models.GetAppsDomainFields{Name: "testdomain.com"},
		}}},
	}

	err := checkCFState(appName, "testdomain.com", candidateHost, newMockAppsGetter(apps, nil))

	assert.Equal(t, err, ErrRouteInUse(candidateHost, "testdomain.com"))
}
