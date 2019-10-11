package halfpipe_cf_plugin

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRequest(t *testing.T) {
	t.Run("halfpipe-push", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-push",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify())
		})

		t.Run("Missing appPath", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-push",
				ManifestPath: "path",
			}
			expectedError := ErrMissingArg(r.Command, "appPath")

			assert.Equal(t, expectedError, r.Verify())
		})

		t.Run("Missing testDomain", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-push",
				ManifestPath: "path",
				AppPath:      "path",
			}
			expectedError := ErrMissingArg(r.Command, "testDomain")

			assert.Equal(t, expectedError, r.Verify())
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

			assert.Equal(t, expectedError, r.Verify())
		})

		t.Run("Valid preStartCommand", func(t *testing.T) {
			r := Request{
				Command:         "halfpipe-push",
				ManifestPath:    "path",
				AppPath:         "path",
				TestDomain:      "test.domain",
				PreStartCommand: "cf something good",
			}

			assert.NoError(t, r.Verify())
		})

	})

	t.Run("halfpipe-check", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-check",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify())
		})
	})

	t.Run("halfpipe-promote", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-promote",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify())
		})

		t.Run("Missing testDomain", func(t *testing.T) {
			r := Request{
				Command:      "halfpipe-promote",
				ManifestPath: "path",
				AppPath:      "path",
			}
			expectedError := ErrMissingArg(r.Command, "testDomain")

			assert.Equal(t, expectedError, r.Verify())
		})

	})

	t.Run("halfpipe-cleanup", func(t *testing.T) {
		t.Run("Missing manifestPath", func(t *testing.T) {
			r := Request{
				Command: "halfpipe-cleanup",
			}
			expectedError := ErrMissingArg(r.Command, "manifestPath")

			assert.Equal(t, expectedError, r.Verify())
		})
	})
}
