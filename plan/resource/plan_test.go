package resource

import (
	"path"
	"testing"

	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/util/manifest"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/spf13/afero"
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
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	_, err := NewPlanner(manifestReaderWithOneApp, manifestWriterWithoutError, fs).Plan(Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
	}, "")
	assert.Equal(t, NewErrEmptyParamValue("manifestPath").Error(), err.Error())

	_, err = NewPlanner(manifestReaderWithOneApp, manifestWriterWithoutError, fs).Plan(Request{
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
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		if pathToManifest == fullManifestPath {
			err = expectedError
		}
		return
	}

	push := NewPlanner(manifestReader, manifestWriterWithoutError, fs)

	_, err := push.Plan(validRequest, concourseRoot)
	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfWeFailToWriteManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"
	fullManifestPath := path.Join(concourseRoot, validRequest.Params.ManifestPath)

	manifestWriter := func(application manifest.Application, filePath string) error {
		if filePath == fullManifestPath {
			return expectedError
		}
		return nil
	}
	push := NewPlanner(manifestReaderWithOneApp, manifestWriter, fs)

	_, err := push.Plan(validRequest, concourseRoot)

	assert.Equal(t, expectedError, err)
}

func TestDoesntWriteManifestIfNotPush(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

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

	push := NewPlanner(manifestReader, manifestWriter, fs)

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

func TestGivesACorrectPlanWhenManifestDoesNotHaveAnyEnvironmentVariables(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	applicationManifest := manifest.Application{
		Name: "MyApp",
	}

	expectedManifest := manifest.Application{
		Name:                 "MyApp",
		EnvironmentVariables: validRequest.Params.Vars,
	}

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{applicationManifest}, nil
	}

	var actualManifest manifest.Application
	manifestWriter := func(application manifest.Application, filePath string) error {
		actualManifest = application

		return nil
	}

	push := NewPlanner(manifestReader, manifestWriter, fs)

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, actualManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}

func TestGivesACorrectPlanThatAlsoOverridesVariablesInManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

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
	push := NewPlanner(manifestReader, manifestWriter, fs)

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, actualManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}

func TestErrorsIfTheGitRefPathIsSpecifiedButDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{{}}, nil
	}
	manifestWriter := func(application manifest.Application, filePath string) error {
		return nil
	}

	push := NewPlanner(manifestReader, manifestWriter, fs)
	request := Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
		Params: Params{
			ManifestPath: "manifest.yml",
			GitRefPath:   "git/.git/ref",
			AppPath:      "",
			TestDomain:   "kehe.com",
			Command:      config.PUSH,
		},
	}
	_, err := push.Plan(request, "/some/path")

	assert.Error(t, err)
}

func TestPutsGitRefInTheManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	concourseRoot := "/some/path"
	gitRefPath := "git/.git/ref"
	gitRef := "wiiiie"
	fs.WriteFile(path.Join(concourseRoot, gitRefPath), []byte(gitRef), 0700)

	manifestReader := func(pathToManifest string) (apps []manifest.Application, err error) {
		return []manifest.Application{{}}, nil
	}

	var writtenManifest manifest.Application
	manifestWriter := func(application manifest.Application, filePath string) error {
		writtenManifest = application
		return nil
	}

	push := NewPlanner(manifestReader, manifestWriter, fs)

	request := Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
		Params: Params{
			ManifestPath: "manifest.yml",
			GitRefPath:   gitRefPath,
			AppPath:      "",
			TestDomain:   "kehe.com",
			Command:      config.PUSH,
		},
	}

	_, err := push.Plan(request, concourseRoot)

	assert.Nil(t, err)
	assert.Equal(t, writtenManifest.EnvironmentVariables["GIT_REVISION"], gitRef)
}

