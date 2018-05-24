package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"errors"
	"path"
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

type ManifestReadWriteStub struct {
	manifest      manifest.Manifest
	readError     error
	writeError    error
	savedManifest manifest.Manifest
}

func (m *ManifestReadWriteStub) ReadManifest(path string) (manifest.Manifest, error) {
	return m.manifest, m.readError
}

func (m *ManifestReadWriteStub) WriteManifest(path string, application manifest.Application) (error) {
	m.savedManifest = manifest.Manifest{
		Applications: []manifest.Application{application},
	}

	return m.writeError
}

func TestNewPushReturnsErrorForEmptyValue(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	_, err := NewPlanner(&ManifestReadWriteStub{}, fs).Plan(Request{
		Source: Source{
			API:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
	}, "")
	assert.Equal(t, NewErrEmptyParamValue("manifestPath").Error(), err.Error())

	_, err = NewPlanner(&ManifestReadWriteStub{}, fs).Plan(Request{
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

	push := NewPlanner(&ManifestReadWriteStub{readError: expectedError}, fs)

	_, err := push.Plan(validRequest, concourseRoot)
	assert.Equal(t, expectedError, err)
}

func TestReturnsErrorIfWeFailToWriteManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	expectedError := errors.New("Shiied")

	concourseRoot := "/tmp/some/path"

	manifest := manifest.Manifest{
		Applications: []manifest.Application{{}},
	}
	push := NewPlanner(&ManifestReadWriteStub{manifest: manifest, writeError: expectedError}, fs)

	_, err := push.Plan(validRequest, concourseRoot)

	assert.Equal(t, expectedError, err)
}

func TestDoesntWriteManifestIfNotPush(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	concourseRoot := "/tmp/some/path"

	push := NewPlanner(
		&ManifestReadWriteStub{
			readError:  errors.New("should not happen"),
			writeError: errors.New("should not happen")},
		fs)

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
}

func TestGivesACorrectPlanWhenManifestDoesNotHaveAnyEnvironmentVariables(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	applicationManifest := manifest.Manifest{
		Applications: []manifest.Application{
			{Name: "MyApp"},
		},
	}
	expectedManifest := manifest.Manifest{
		Applications: []manifest.Application{
			{
				Name:                 "MyApp",
				EnvironmentVariables: validRequest.Params.Vars,
			},
		},
	}

	manifestReadWrite := &ManifestReadWriteStub{manifest: applicationManifest}

	push := NewPlanner(manifestReadWrite, fs)

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, manifestReadWrite.savedManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}

func TestGivesACorrectPlanThatAlsoOverridesVariablesInManifest(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}

	applicationManifest := manifest.Manifest{
		Applications: []manifest.Application{
			{
				Name: "MyApp",
				EnvironmentVariables: map[string]string{
					"VAR1": "a",
					"VAR2": "b",
					"VAR3": "c",
				},
			},
		},
	}

	expectedManifest := manifest.Manifest{
		Applications: []manifest.Application{
			{
				Name: "MyApp",
				EnvironmentVariables: map[string]string{
					"VAR1": "a",
					"VAR2": "bb",
					"VAR3": "c",
					"VAR4": "cc",
				},
			},
		},
	}

	manifestReaderWriter := ManifestReadWriteStub{manifest: applicationManifest}
	push := NewPlanner(&manifestReaderWriter, fs)

	p, err := push.Plan(validRequest, "")

	assert.Nil(t, err)
	assert.Equal(t, expectedManifest, manifestReaderWriter.savedManifest)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}


func TestErrorsIfTheGitRefPathIsSpecifiedButDoesntExist(t *testing.T) {
	fs := afero.Afero{Fs: afero.NewMemMapFs()}


	push := NewPlanner(&ManifestReadWriteStub{
		manifest: manifest.Manifest{[]manifest.Application{{}}},
	}, fs)
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
	gitRef := "wiiiie\n"
	fs.WriteFile(path.Join(concourseRoot, gitRefPath), []byte(gitRef), 0700)

	applicationManifest := manifest.Manifest{
		Applications: []manifest.Application{
			{
				Name: "MyApp",
				EnvironmentVariables: map[string]string{
					"VAR1": "a",
					"VAR2": "b",
					"VAR3": "c",
				},
			},
		},
	}

	stub := ManifestReadWriteStub{manifest:applicationManifest}
	push := NewPlanner(&stub, fs)

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
	assert.Equal(t, stub.savedManifest.Applications[0].EnvironmentVariables["GIT_REVISION"], "wiiiie")
}

