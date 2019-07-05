package plan

import (
	"fmt"
	"github.com/springernature/halfpipe-cf-plugin/command"
)

type Plan []command.Command

func (p Plan) String() (s string) {
	if p.IsEmpty() {
		s += "# Nothing to do!"
		return
	}

	s += "# Planned execution\n"
	for _, command := range p {
		s += fmt.Sprintf("#\t* %s\n", command)
	}
	return
}

func (p Plan) IsEmpty() bool {
	return len(p) == 0
}
