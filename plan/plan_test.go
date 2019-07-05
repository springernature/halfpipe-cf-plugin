package plan

import (
	"github.com/springernature/halfpipe-cf-plugin/command"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlan_String(t *testing.T) {
	p := Plan{
		command.NewCfShellCommand("push"),
		command.NewCfShellCommand("delete"),
	}

	expected := `# Planned execution
#	* cf push
#	* cf delete
`
	assert.Equal(t, expected, p.String())
}