package halfpipe_cf_plugin

import (
	"github.com/spf13/afero"
	"github.com/springernature/halfpipe-cf-plugin/manifest"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequest(t *testing.T) {
	manifestPath := "some/path"
	fs := afero.Afero{Fs: afero.NewMemMapFs()}
	man := `---
applications:
- name: yo
`
	fs.WriteFile(manifestPath, []byte(man), 777)
	fakeManifestReader := manifest.NewManifestReadWrite(fs)

	t.Run("halfpipe-push", func(t *testing.T) {

		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-push",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Missing appPath", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-push",
				ManifestPath: "path",
			}
			expectedError := ErrMissingArg(r.Command, "appPath")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Missing testDomain", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-push",
				ManifestPath: "path",
				AppPath:      "path",
			}
			expectedError := ErrMissingArg(r.Command, "testDomain")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Invalid preStartCommand", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "something bad",
			}
			expectedError := ErrInvalidPreStartCommand("something bad")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Valid preStartCommand", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "cf something good",
			}

			assert.NoError(t, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Invalid preStartCommand - semi-colon delimited", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "cf something good; something bad",
			}
			expectedError := ErrInvalidPreStartCommand("something bad")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Valid preStartCommand - semi-colon delimited", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "cf something good; cf something else good",
			}

			assert.NoError(t, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Empty preStartCommand", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "",
			}

			assert.NoError(t, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Missing appPath when its a docker push", func(t *testing.T) {
			f := afero.Afero{Fs: afero.NewMemMapFs()}
			m := `---
applications:
- name: yo
  docker:
    image: nginx
`
			f.WriteFile(manifestPath, []byte(m), 777)
			fMR := manifest.NewManifestReadWrite(f)

			r := Request{
				Command:      "halfpipe-push",
				ManifestPath: "path",
				TestDomain:   "Whoo",
			}

			assert.NoError(t, r.Verify(manifestPath, fMR))
		})

	})

	t.Run("halfpipe-check", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-check",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})
	})

	t.Run("halfpipe-promote", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-promote",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

		t.Run("Missing testDomain", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-promote",
				ManifestPath: "path",
				AppPath:      "path",
			}
			expectedError := ErrMissingArg(r.Command, "testDomain")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})

	})

	t.Run("halfpipe-cleanup", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-cleanup",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify(manifestPath, fakeManifestReader))
		})
	})
}
