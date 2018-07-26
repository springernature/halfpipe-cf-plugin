package resource_test

import (
	"testing"
	"github.com/springernature/halfpipe-cf-plugin/plan/resource"
	"github.com/stretchr/testify/assert"
	"github.com/springernature/halfpipe-cf-plugin/config"
)

func TestVerifyErrorsIfNotAllSourceFieldsAreFilledOut(t *testing.T) {
	invalidSourceRequests := []resource.Source{
		{
			API:      "",
			Org:      "",
			Space:    "",
			Username: "",
			Password: "",
		},

		{
			API:      "a",
			Org:      "",
			Space:    "",
			Username: "",
			Password: "",
		},

		{
			API:      "a",
			Org:      "a",
			Space:    "",
			Username: "",
			Password: "",
		},

		{
			API:      "a",
			Org:      "a",
			Space:    "a",
			Username: "",
			Password: "",
		},

		{
			API:      "a",
			Org:      "a",
			Space:    "a",
			Username: "a",
			Password: "",
		},
	}

	for _, source := range invalidSourceRequests {
		assert.Error(t, resource.VerifyRequestSource(source))
	}

	validSource := resource.Source{
		API:      "a",
		Org:      "a",
		Space:    "a",
		Username: "a",
		Password: "c",
	}

	assert.Nil(t, resource.VerifyRequestSource(validSource))
}

func TestVerifyErrorsIfNotAllRequiredParamsFieldsAreFilledOut(t *testing.T) {
	missingCommand := resource.Params{
		Command: "",
	}
	assert.Equal(t, resource.ParamsMissingError("command"), resource.VerifyRequestParams(missingCommand))

	missingManifestPath := resource.Params{
		Command: "Something",
	}
	assert.Equal(t, resource.ParamsMissingError("manifestPath"), resource.VerifyRequestParams(missingManifestPath))

}

func TestVerifyErrorsIfNotAllRequiredParamsFieldsForPushFilledOut(t *testing.T) {
	missingTestDomain := resource.Params{
		Command:      config.PUSH,
		ManifestPath: "path",
		TestDomain:   "",
	}
	assert.Equal(t, resource.ParamsMissingError("testDomain"), resource.VerifyRequestParams(missingTestDomain))

	missingAppPath := resource.Params{
		Command:      config.PUSH,
		ManifestPath: "path",
		TestDomain:   "test.com",
		AppPath:      "",
	}
	assert.Equal(t, resource.ParamsMissingError("appPath"), resource.VerifyRequestParams(missingAppPath))

	missingGitRefPath := resource.Params{
		Command:      config.PUSH,
		ManifestPath: "path",
		TestDomain:   "test.com",
		AppPath:      "path",
		GitRefPath:   "",
	}
	assert.Equal(t, resource.ParamsMissingError("gitRefPath"), resource.VerifyRequestParams(missingGitRefPath))

	allesOk := resource.Params{
		Command:      config.PUSH,
		ManifestPath: "path",
		TestDomain:   "test.com",
		AppPath:      "path",
		GitRefPath:   "path",
	}
	assert.Nil(t, resource.VerifyRequestParams(allesOk))
}

func TestVerifyErrorsIfNotAllRequiredParamsFieldsForPromoteFilledOut(t *testing.T) {
	missingTestDomain := resource.Params{
		Command:      config.PROMOTE,
		ManifestPath: "path",
		TestDomain:   "",
	}
	assert.Equal(t, resource.ParamsMissingError("testDomain"), resource.VerifyRequestParams(missingTestDomain))

	allesOk := resource.Params{
		Command:      config.PROMOTE,
		ManifestPath: "path",
		TestDomain:   "test.com",
	}
	assert.Nil(t, resource.VerifyRequestParams(allesOk))
}

func TestVerifyErrorsIfNotAllRequiredParamsFieldsForCleanupFilledOut(t *testing.T) {
	allesOk := resource.Params{
		Command:      config.CLEANUP,
		ManifestPath: "path",
	}
	assert.Nil(t, resource.VerifyRequestParams(allesOk))
}