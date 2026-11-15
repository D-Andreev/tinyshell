package commands

import (
	"fmt"
	"io"
	"strings"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/path"
)

type Type struct {
	builtin *Builtin
}

func NewType(builtin *Builtin) *Type {
	return &Type{builtin}
}

func (t *Type) Execute(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	var res strings.Builder
	status := 0
	p := path.NewPath()
	for _, arg := range args {
		_, err := t.builtin.Get(arg)
		if err != nil {
			// command is not builtin, check if it's external
			externalProgramPath, err := p.Get(arg)
			if err != nil {
				// command is neither builtin nor external
				fmt.Fprintf(&res, constants.COMMAND_NOT_FOUND, arg)
				status = 1
				continue
			}
			res.WriteString(arg + " is " + externalProgramPath + "\n")
			continue
		}
		res.WriteString(arg + " is /bin/" + arg + "\n")
	}

	stdout.Write([]byte(res.String()))
	return status
}
