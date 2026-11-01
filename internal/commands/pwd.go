package commands

import (
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
) (string, func(), error) {
	currentDir, error := state.Get("currentDir")

	return currentDir + "\n", constants.NOOP, error
}
