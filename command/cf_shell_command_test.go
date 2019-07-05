package command

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCommand_String(t *testing.T) {
	c := NewCfShellCommand("push", "-f", "man")

	expected := "cf push -f man"
	assert.Equal(t, expected, c.String())
}

func TestHidesPasswordIfLogin(t *testing.T) {
	c := NewCfShellCommand("login", "-a", "api", "-p", "password", "-u", "username")

	expected := "cf login -a api -p ******** -u username"
	assert.Equal(t, expected, c.String())
}

func TestDoesntHideValueToPFlagIfNotLogin(t *testing.T) {
	c := NewCfShellCommand("push", "appname", "-p", "path/to/app/bits")

	expected := "cf push appname -p path/to/app/bits"
	assert.Equal(t, expected, c.String())
}
