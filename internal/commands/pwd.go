package commands

import (
	"io"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/state"
)

type Pwd struct {
}

func NewPwd() *Pwd {
	return &Pwd{}
}

func (p *Pwd) Execute(
	state *state.State,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) func() {
	currentDir, err := state.Get("currentDir")

	if err != nil {
		stderr.Write([]byte(err.Error()))
		return constants.NOOP
	}

	stdout.Write([]byte(currentDir + "\n"))
	return constants.NOOP
}
