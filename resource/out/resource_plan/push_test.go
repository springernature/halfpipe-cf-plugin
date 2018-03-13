package resource_plan

import (
	"testing"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"github.com/stretchr/testify/assert"
)

func TestNewPushReturnsErrorForEmptyValue(t *testing.T) {
	_, err := NewPush().Plan(out.Request{
		Source: out.Source{
			Api:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
	})
	assert.Equal(t, NewErrEmptyParamValue("manifestPath").Error(), err.Error())

	_, err = NewPush().Plan(out.Request{
		Params: out.Params{
			ManifestPath: "f",
			AppPath:      "",
		},
	})
	assert.Equal(t, NewErrEmptySourceValue("space").Error(), err.Error())

}

func TestReturnsAPlanForCorrectRequest(t *testing.T) {
	request := out.Request{
		Source: out.Source{
			Api:      "a",
			Org:      "b",
			Space:    "c",
			Username: "d",
			Password: "e",
		},
		Params: out.Params{
			ManifestPath: "f",
			AppPath:      "",
		},
	}

	p, err := NewPush().Plan(request)

	assert.Nil(t, err)
	assert.Len(t, p, 2)
	assert.Contains(t, p[0].String(), "cf login")
	assert.Contains(t, p[1].String(), "cf halfpipe-push")
}
