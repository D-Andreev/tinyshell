package commands

import (
	"io"
	"strings"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/state"
)

type Echo struct {
}

func NewEcho() *Echo {
	return &Echo{}
}

func (e *Echo) Execute(
	state *state.State,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) func() {
	if len(args) == 0 {
		stdout.Write([]byte("\n"))
		return constants.NOOP
	}

	var res strings.Builder
	for i, a := range args {
		if i == len(args)-1 {
			res.WriteString(a)
			continue
		}
		res.WriteString(a + " ")
	}
	res.WriteString("\n")

	stdout.Write([]byte(res.String()))

	return constants.NOOP
}
