package resource_plan

import (
	"testing"
	"github.com/springernature/halfpipe-cf-plugin/resource/out"
	"github.com/stretchr/testify/assert"
)

func TestNewPushReturnsErrorForEmptyValue(t *testing.T) {
	_, err := NewPush().Plan(out.Request{})
	assert.IsType(t, ErrEmptyParamValue(""), err)
}

func TestReturnsAPlanForCorrectRequest(t *testing.T) {
	request := out.Request{
		Params: out.Params{
			Api:          "a",
			Org:          "b",
			Space:        "c",
			Username:     "d",
			Password:     "e",
			ManifestPath: "f",
			AppPath:      "",
		},
	}

	p, err := NewPush().Plan(request)

	assert.Nil(t, err)
	assert.Len(t, p, 2)
}
