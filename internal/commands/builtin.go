package commands

import (
	"fmt"
	"io"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/state"
)

type Commander interface {
	Execute(
		state *state.State,
		args []string,
		stdin io.Reader,
		stdout io.Writer,
		stderr io.Writer,
	) func()
}

type Builtin struct {
	commands map[string]Commander
}

func NewBuiltin() *Builtin {
	b := &Builtin{
		commands: map[string]Commander{
			"exit": NewExit(),
			"echo": NewEcho(),
			"pwd":  NewPwd(),
			"cd":   NewCd(),
		},
	}
	b.commands["type"] = NewType(b)
	return b
}

func (b *Builtin) Get(commandName string) (Commander, error) {
	c, ok := b.commands[commandName]
	if !ok {
		return nil, fmt.Errorf(constants.COMMAND_NOT_FOUND, commandName)
	}

	return c, nil
}
