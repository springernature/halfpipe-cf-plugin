package resource

import (
	"path"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
)

var validRequest = Request{
	Source: Source{
		API:      "a",
		Org:      "b",
		Space:    "c",
		Username: "d",
		Password: "e",
	},
	Params: Params{
		ManifestPath: "manifest.yml",
		AppPath:      "",
		TestDomain:   "kehe.com",
		Command:      config.PUSH,
		Vars: map[string]string{
			"VAR2": "bb",
			"VAR4": "cc",
		},
	},
}
var manifestReaderWithOneApp = func(pathToManifest string) (apps []manifest.Application, err error) {
	return []manifest.Application{{}}, nil
}

var manifestWriterWithoutError = func(application manifest.Application, filePath string) error {
	return nil
}

func TestNewPushReturnsErrorForEmptyValue(t *testing.T) {
	_, err := NewPlanner(manifestReaderWithOneApp, manifestWriterWithoutError).Plan(Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
	}, "")
	assert.Equal(t, NewErrEmptyParamValue("manifestPath").Error(), err.Error())

	_, err = NewPlanner(manifestReaderWithOneApp, manifestWriterWithoutError).Plan(Request{
		Params: Params{
			Command:      config.PUSH,
			ManifestPath: "f",
			AppPath:      "",
			TestDomain:   "a",
		},
	}, "")
	assert.Equal(t, NewErrEmptySourceValue("space").Error(), err.Error())
}

func TestReturnsErrorIfWeFailToReadManifest(t *testing.T) {
	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		if pathToManifest == fullManifestPath {
			err = expectedError
		}
		return
	}

	push := NewPlanner(manifestReader, manifestWriterWithoutError)

	_, err := push.Plan(validRequest, concourseRoot)
	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfWeFailToWriteManifest(t *testing.T) {
	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	manifestWriter := func(application manifest.Application, filePath string) error {
		if filePath == fullManifestPath {
			return expectedError
		}
		return nil
	}
	push := NewPlanner(manifestReaderWithOneApp, manifestWriter)

	_, err := push.Plan(validRequest, concourseRoot)

	assert.Equal(t, expectedError, err)
}

func TestDoesntWriteManifestIfNotPush(t *testing.T) {
	concourseRoot := "/tmp/some/path"

	var manifestReaderCalled = false
	var manifestWriterCalled = false

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		manifestReaderCalled = true
		return []manifest.Application{}, nil
	}
	manifestWriter := func(application manifest.Application, filePath string) error {
		manifestWriterCalled = true
		return nil
	}

	push := NewPlanner(manifestReader, manifestWriter)

	validPromoteRequest := Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
		Params: Params{
			ManifestPath: "manifest.yml",
			AppPath:      "",
			TestDomain:   "kehe.com",
			Command:      config.PROMOTE,
			Vars: map[string]string{
				"VAR2": "bb",
				"VAR4": "cc",
			},
		},
	}

	_, err := push.Plan(validPromoteRequest, concourseRoot)

	assert.Nil(t, err)
	assert.False(t, manifestReaderCalled)
	assert.False(t, manifestWriterCalled)
}

func TestGivesACorrectPlanThatAlsoOverridesVariablesInManifest(t *testing.T) {
	applicationManifest := manifest.Application{
		Name: "MyApp",
		EnvironmentVariables: map[string]string{
			"VAR1": "a",
			"VAR2": "b",
			"VAR3": "c",
		},
	}

	expectedManifest := manifest.Application{
		Name: "MyApp",
		EnvironmentVariables: map[string]string{
			"VAR1": "a",
			"VAR2": "bb",
			"VAR3": "c",
			"VAR4": "cc",
		},
	}
	var actualManifest manifest.Application

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{applicationManifest}, nil
	}
	manifestWriter := func(application manifest.Application, filePath string) error {
		actualManifest = application
		return nil
	}
	push := NewPlanner(manifestReader, manifestWriter)

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, actualManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}
