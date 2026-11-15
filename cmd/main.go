package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/tinyshell/internal/commands"
	"github.com/tinyshell/internal/constants"
	"github.com/tinyshell/internal/executor"
	"github.com/tinyshell/internal/lexer"
	"github.com/tinyshell/internal/parser"
	"github.com/tinyshell/internal/path"
	"github.com/tinyshell/internal/token"
)

func main() {
	stdinR := bufio.NewReader(os.Stdin)
	builtinCommands := commands.NewBuiltin()
	p := path.NewPath()

	for {
		parser := parser.NewParser()
		executor := executor.NewExecutor(builtinCommands)
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Printf(constants.ERROR_READ_STDIN, err.Error())
			os.Exit(1)
		}

		// read input
		fmt.Printf("%s ", constants.SHELL_TRAILING_SYMBOL+"~"+p.ReplaceHomeDir(currentDir))
		input, err := stdinR.ReadString('\n')
		if err != nil {
			fmt.Printf(constants.ERROR_READ_STDIN, err.Error())
			os.Exit(1)
		}
		l := lexer.New(input)
		var tokens []token.Token
		illegal := false
		for {
			tok := l.NextToken()
			if tok.Type == token.EOF {
				break
			}
			if tok.Type == token.ILLEGAL {
				fmt.Printf("syntax error: unterminated quote %s\n", tok.Literal)
				illegal = true
				break
			}
			tokens = append(tokens, tok)
		}
		if illegal || len(tokens) == 0 {
			continue
		}
		parser.Parse(tokens)
		executor.Execute(parser.Pipeline, p)
	}
}
