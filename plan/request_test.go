package plan

import (
	"testing"
	"github.com/stretchr/testify/assert"
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
