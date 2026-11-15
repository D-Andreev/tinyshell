package commands

import (
	"io"
	"os"
)

type Pwd struct {
}

func NewPwd() *Pwd {
	return &Pwd{}
}

func (p *Pwd) Execute(
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	currentDir, err := os.Getwd()

	if err != nil {
		stderr.Write([]byte(err.Error()))
		return 1
	}

	stdout.Write([]byte(currentDir + "\n"))
	return 0
}
