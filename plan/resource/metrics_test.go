package resource

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"strings"

	"github.com/stretchr/testify/assert"
)

func TestNewPrometheusMetrics(t *testing.T) {
	var path string
	var counter int
	gateway := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		counter++
	}))
	defer gateway.Close()

	m := NewMetrics(Request{
		Source: Source{
			PrometheusGatewayURL: gateway.URL,
			API:                  "some.cf.api",
			Org:                  "some-cf-org",
		},
		Params: Params{
			Command: "promote",
		},
	})

	m.Success()
	assert.Equal(t, 1, counter)
	assert.True(t, strings.HasPrefix(path, "/metrics/job/promote/"), path)
	assert.Contains(t, path, "cf_api/some.cf.api")
	assert.Contains(t, path, "cf_org/some-cf-org")

	//m.Failure()
	//assert.Equal(t, 2, counter)
}
