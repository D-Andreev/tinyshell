package commands

import (
	"strings"

	"github.com/tinyshell/internal/constants"
)

type Echo struct {
}

func NewEcho() *Echo {
	return &Echo{}
}

func (e *Echo) Execute(args []string) (string, func(), error) {
	if len(args) == 0 {
		return "\n", constants.NOOP, nil
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

	return res.String(), constants.NOOP, nil
}
