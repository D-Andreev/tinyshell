package commands

import (
	"fmt"

	"github.com/tinyshell/internal/constants"
)

type Commander interface {
	Execute(args []string) (string, func(), error)
}

type Builtin struct {
	commands map[string]Commander
}

func NewBuiltin() *Builtin {
	return &Builtin{
		commands: map[string]Commander{
			"exit": NewExit(),
			"echo": NewEcho(),
			"type": NewType(),
		},
	}
}

func (b *Builtin) Get(commandName string) (Commander, error) {
	c, ok := b.commands[commandName]
	if !ok {
		return nil, fmt.Errorf(constants.COMMAND_NOT_FOUND, commandName)
	}

	return c, nil
}
