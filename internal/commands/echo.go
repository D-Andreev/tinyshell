package commands

import (
	"io"
	"strings"
)

type Echo struct {
}

func NewEcho() *Echo {
	return &Echo{}
}

func (e *Echo) Execute(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	if len(args) == 0 {
		stdout.Write([]byte("\n"))
		return 0
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

	return 0
}
