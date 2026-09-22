package commands

import (
	"fmt"
	"strings"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/path"
)

type Type struct {
}

func NewType() *Type {
	return &Type{}
}

func (t *Type) Execute(args []string) (string, func(), error) {
	var res strings.Builder
	builtinCommands := NewBuiltin()
	p := path.NewPath()
	for _, arg := range args {
		_, err := builtinCommands.Get(arg)
		if err != nil {
			// command is not builtin, check if it's external
			externalProgramPath, err := p.Get(arg)
			if err != nil {
				// command is neither builtin nor external
				fmt.Fprintf(&res, constants.COMMAND_NOT_FOUND, arg)
				continue
			}
			res.WriteString(arg + " is " + externalProgramPath + "\n")
			continue
		}
		res.WriteString(arg + " is /bin/" + arg + "\n")
	}

	return res.String(), constants.NOOP, nil
}
