package parser

import (
	"fmt"

	"github.com/tinyshell/internal/token"
)

type Pipeline struct {
	Commands []Command
}

type Command struct {
	Args   []string
	Redirs []Redirect
}

type Redirect struct {
	FD     int    // 1 or 2 (stdout or stderr)
	Append bool   // > vs >>
	Target string // filename
}

type Parser struct {
	Pipeline *Pipeline
}

func NewParser() *Parser {
	return &Parser{
		Pipeline: &Pipeline{
			Commands: make([]Command, 0),
		},
	}
}

func (p *Parser) Parse(tokens []token.Token) error {
	if len(tokens) > 0 && tokens[0].Type == token.Pipe {
		return fmt.Errorf("Exptected something before pipe")
	}

	if len(tokens) == 0 || (len(tokens) > 0 && tokens[0].Type == token.EOF) {
		return nil
	}

	var currC Command
	currCIdx := 0
	for i := 0; i < len(tokens); i++ {
		t := tokens[i]
		if t.Type == token.Word {
			currC.Args = append(currC.Args, t.Literal)
			continue
		}
		if p.IsRedirect(t.Type) {
			if i+1 >= len(tokens) || tokens[i+1].Type != token.Word {
				return fmt.Errorf("Exptected word after redirect")
			}
			// Both tokens are consumed
			currC.Redirs = append(currC.Redirs, Redirect{
				FD:     p.RedirectFD(t.Type),
				Append: p.IsRedirectAppend(t.Type),
				Target: tokens[i+1].Literal,
			})
			i++
			continue
		}
		if t.Type == token.Pipe {
			if i+1 >= len(tokens) || tokens[i+1].Type != token.Word {
				return fmt.Errorf("Exptected word after pipe")
			}
			p.Pipeline.Commands = append(p.Pipeline.Commands, currC)
			currC = Command{}
			currCIdx++
			continue
		}
	}

	p.Pipeline.Commands = append(p.Pipeline.Commands, currC)
	return nil
}

func (p *Parser) IsRedirect(tokenType token.TokenType) bool {
	return tokenType == token.RedirOut ||
		tokenType == token.RedirAppend ||
		tokenType == token.RedirErr ||
		tokenType == token.RedirErrAppend
}

func (p *Parser) RedirectFD(tokenType token.TokenType) int {
	switch tokenType {
	case token.RedirErr, token.RedirErrAppend:
		return 2
	default:
		return 1
	}
}

func (p *Parser) IsRedirectAppend(tokenType token.TokenType) bool {
	return tokenType == token.RedirAppend ||
		tokenType == token.RedirErrAppend
}
