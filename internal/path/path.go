package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tinyshell/internal/constants"
)

var PATH_SEP = ":"

type Path struct {
	paths string
}

func NewPath() *Path {
	return &Path{
		paths: os.Getenv("PATH"),
	}
}

func (p *Path) GetAll() []string {
	return strings.Split(os.Getenv("PATH"), PATH_SEP)
}

func (p *Path) Get(executable string) (string, error) {
	paths := p.GetAll()
	for _, path := range paths {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		for _, entry := range entries {
			if entry.Name() == executable {
				return path + "/" + entry.Name(), nil
			}
		}
	}

	return "", fmt.Errorf(constants.COMMAND_NOT_FOUND, executable)
}

func GetCurrentPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}
