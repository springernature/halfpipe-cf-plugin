package command

import (
	"fmt"
)

type Command interface {
	fmt.Stringer
	Args() []string
}
