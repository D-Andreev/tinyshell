package commands

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/state"
)

type Cd struct {
}

func NewCd() *Cd {
	return &Cd{}
}

/*
cd /abs/path works.
cd rel/path, cd .. and cd ./x resolve against the current directory.
cd ~ and cd with no argument go to $HOME. cd ~/foo also works.
A missing target prints cd: <path>: No such file or directory and leaves the directory unchanged.
A target that is a file, not a directory, prints cd: <path>: Not a directory.
Stretch: cd - goes back to the previous directory ($OLDPWD) and prints it.
The change must persist: after cd /tmp, both pwd and an external ls see /tmp.
*/
func (cd *Cd) Execute(
	state *state.State,
	args []string,
	stdin io.Reader,
	stdout io.Writer,
	stderr io.Writer,
) func() {
	targetPath := os.Getenv("HOME")
	effect := func() {
		state.Set("currentDir", targetPath)
	}

	// Go to home page
	if len(args) == 0 || (len(args) == 1 && args[0] == constants.TILDE) {
		return effect
	}

	args[0] = strings.ReplaceAll(args[0], "~", targetPath)

	// Resolve path
	path, err := filepath.Abs(args[0])
	if err != nil {
		return constants.NOOP
	}

	// Check if path is valid
	fileInfo, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		stderr.Write([]byte(fmt.Sprintf(constants.ERROR_NO_SUCH_FILE_OR_DIR, path)))
		return constants.NOOP
	} else if !fileInfo.IsDir() {
		stderr.Write([]byte(fmt.Sprintf(constants.ERROR_NOT_A_DIRECTORY, path)))
		return constants.NOOP
	}

	return effect
}
