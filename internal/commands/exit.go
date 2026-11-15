package commands

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/state"
)

type Exit struct {
}

func NewExit() *Exit {
	return &Exit{}
}

func (e *Exit) Execute(
	state *state.State,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) int {
	var exitCode int
	var err error
	if len(args) > 0 {
		exitCode, err = strconv.Atoi(args[0])
		if err != nil {
			stderr.Write([]byte(fmt.Sprintf(constants.INVALID_EXIT_CODE+"\n", args[0])))
			return 1
		}
	}
	os.Exit(exitCode)
	return exitCode
}
