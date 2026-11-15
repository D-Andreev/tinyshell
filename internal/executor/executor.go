package executor

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/tinyshell/internal/commands"
	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/parser"
	"github.com/tinyshell/internal/path"
)

type Executor struct {
	builtin *commands.Builtin
}

func NewExecutor(builtin *commands.Builtin) *Executor {
	return &Executor{builtin}
}

func (e *Executor) Execute(
	pipeline *parser.Pipeline,
	p *path.Path,
) {
	command, err := e.builtin.Get(pipeline.Commands[0].Args[0])
	if err != nil {
		e.RunExternalCmd(pipeline, p)
		return
	}

	args := pipeline.Commands[0].Args[1:]
	_ = command.Execute(args, os.Stdin, os.Stdout, os.Stderr)
}

func (e *Executor) RunExternalCmd(
	pipeline *parser.Pipeline,
	p *path.Path,
) {
	// command is not builtin, try to execute external program
	extProgramPath, err := p.Get(pipeline.Commands[0].Args[0])
	if err != nil {
		// command is neither builtin nor external
		fmt.Printf(constants.COMMAND_NOT_FOUND, pipeline.Commands[0].Args[0])
		return
	}

	info, err := os.Stat(extProgramPath)
	if err != nil {
		fmt.Printf(
			constants.ERROR_READING_EXTERNAL_PROGRAM, extProgramPath,
		)
		return
	}

	mode := info.Mode()

	// For the purposses of this project we're just checking if any execute permission exists, regardless of current user and owner, group or other.
	executePermissionFound := false
	for i := 1; i < 10; i++ {
		permissionBit := string(mode.String()[i])
		if permissionBit == constants.EXECUTE_PERMISSION_BIT {
			executePermissionFound = true
			break
		}
	}
	if !executePermissionFound {
		fmt.Printf(constants.PERMISSION_DENIED, extProgramPath)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()
	cmd := exec.CommandContext(ctx, extProgramPath, pipeline.Commands[0].Args[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf(
			constants.ERROR_FROM_EXT_PROGRAM,
			extProgramPath,
			err.Error(),
		)
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		fmt.Printf(
			constants.ERROR_FROM_EXT_PROGRAM,
			extProgramPath,
			err.Error(),
		)
		return
	}
	go func() {
		io.Copy(os.Stderr, stderr)
	}()
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, stdout)
	if err := cmd.Wait(); err != nil {
		fmt.Printf(
			constants.ERROR_FROM_EXT_PROGRAM,
			extProgramPath,
			err.Error(),
		)
		return
	}
}
