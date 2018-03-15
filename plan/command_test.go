package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommand_String(t *testing.T) {
	c := NewCfCommand("push", "-f", "man")

	expected := "cf push -f man"
	assert.Equal(t, expected, c.String())
}

func TestHidesPasswordIfLogin(t *testing.T) {
	c := NewCfCommand("login", "-a", "api", "-p", "password", "-u", "username")

	expected := "cf login -a api -p ******** -u username"
	assert.Equal(t, expected, c.String())
}

func TestDoesntHideValueToPFlagIfNotLogin(t *testing.T) {
	c := NewCfCommand("push", "appname", "-p", "path/to/app/bits")

	expected := "cf push appname -p path/to/app/bits"
	assert.Equal(t, expected, c.String())
}
