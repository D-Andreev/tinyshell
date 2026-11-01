package commands

import (
	"fmt"
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
) (string, func(), error) {
	var exitCode int
	var err error
	if len(args) > 0 {
		exitCode, err = strconv.Atoi(args[0])
		if err != nil {
			return "", constants.NOOP, fmt.Errorf(constants.INVALID_EXIT_CODE, args[0])
		}
	}
	effect := func() {
		os.Exit(exitCode)
	}
	return "", effect, nil
}
