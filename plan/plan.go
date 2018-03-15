package plan

import (
	"fmt"
	"log"
)

type Plan []Command

func (p Plan) String() (s string) {
	if p.IsEmpty() {
		s += "Nothing to do!"
		return
	}

	s += "Planned execution\n"
	for _, command := range p {
		s += fmt.Sprintf("\t* %s\n", command)
	}
	return
}

func (p Plan) Execute(executor Executor, logger *log.Logger) (err error) {
	for _, command := range p {
		logger.Println(fmt.Sprintf("=== Executing '%s' ===", command))
		_, err = executor.CliCommand(command.Args()...)
		if err != nil {
			return
		}
		logger.Println(fmt.Sprintf("=== Succeeded :D ==="))
		logger.Println()
	}
	return
}

func (p Plan) IsEmpty() bool {
	return len(p) == 0
}
