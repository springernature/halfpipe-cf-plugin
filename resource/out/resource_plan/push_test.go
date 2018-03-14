package resource_plan

import (
	"testing"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"github.com/stretchr/testify/assert"
	"path"
	"code.cloudfoundry.org/cli/util/manifest"
	"code.cloudfoundry.org/cli/cf/errors"
)

var validRequest = out.Request{
	Source: out.Source{
		Api:      "a",
		Org:      "b",
		Space:    "c",
		Username: "d",
		Password: "e",
	},
	Params: out.Params{
		ManifestPath: "manifest.yml",
		AppPath:      "",
		TestDomain:   "kehe.com",
		Vars: map[string]string{
			"VAR2": "bb",
			"VAR4": "cc",
		},
	},
}

func TestNewPushReturnsErrorForEmptyValue(t *testing.T) {
	_, err := NewPush().Plan(out.Request{
		Source: out.Source{
			Api:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
	}, "")
	assert.Equal(t, NewErrEmptyParamValue("manifestPath").Error(), err.Error())

	_, err = NewPush().Plan(out.Request{
		Params: out.Params{
			ManifestPath: "f",
			AppPath:      "",
			TestDomain: "a",
		},
	}, "")
	assert.Equal(t, NewErrEmptySourceValue("space").Error(), err.Error())
}

func TestReturnsErrorIfWeFailToReadManifest(t *testing.T) {
	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	push := NewPush()
	push.manifestReader = func(pathToManifest string) (apps []manifest.Application, err error) {
		if pathToManifest == fullManifestPath {
			err = expectedError
		}
		return
	}

	_, err := push.Plan(validRequest, concourseRoot)
	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfWeFailToWriteManifest(t *testing.T) {
	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	push := NewPush()
	push.manifestReader = func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{{}}, nil
	}
	push.manifestWriter = func(application manifest.Application, filePath string) error {
		if filePath == fullManifestPath {
			return expectedError
		}
		return nil
	}

	_, err := push.Plan(validRequest, concourseRoot)

	assert.Equal(t, expectedError, err)
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
	push := NewPush()
	push.manifestReader = func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{applicationManifest}, nil
	}
	push.manifestWriter = func(application manifest.Application, filePath string) error {
		actualManifest = application
		return nil
	}

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, actualManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}
