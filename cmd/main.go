package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/tinyshell/internal/commands"
	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/path"
)

func main() {
	stdinR := bufio.NewReader(os.Stdin)
	builtinCommands := commands.NewBuiltin()
	p := path.NewPath()

	for {
		// read input
		fmt.Printf("%s ", constants.SHELL_TRAILING_SYMBOL)
		input, err := stdinR.ReadString('\n')
		if err != nil {
			fmt.Printf(constants.ERROR_READ_STDIN, err.Error())
			os.Exit(1)
		}
		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}
		command, err := builtinCommands.Get(parts[0])
		if err != nil {
			// command is not builtin, try to execute external program
			extProgramPath, err := p.Get(parts[0])
			if err != nil {
				// command is neither builtin nor external
				fmt.Printf(constants.COMMAND_NOT_FOUND, parts[0])
				continue
			}

			info, err := os.Stat(extProgramPath)
			if err != nil {
				fmt.Printf(
					constants.ERROR_READING_EXTERNAL_PROGRAM, extProgramPath,
				)
				continue
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
			cmd := exec.CommandContext(ctx, extProgramPath, parts[1:]...)
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				fmt.Printf(
					constants.ERROR_FROM_EXT_PROGRAM,
					extProgramPath,
					err.Error(),
				)
				continue
			}
			stderr, err := cmd.StderrPipe()
			if err != nil {
				fmt.Printf(
					constants.ERROR_FROM_EXT_PROGRAM,
					extProgramPath,
					err.Error(),
				)
				continue
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
				continue
			}

			continue
		}

		// execute
		args := parts[1:]
		res, effect, err := command.Execute(args)
		if err != nil {

		}
		// print output
		if len(res) > 0 {
			fmt.Printf("%s", res)
		}
		effect()
	}
}
