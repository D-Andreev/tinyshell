// Package testutil provides helpers shared by tests across packages.
package testutil

import "github.com/tinyshell/internal/token"

// W returns a Word token with the given value.
func W(s string) token.Token { return token.Token{Type: token.Word, Literal: s} }

var (
	PipeTok = token.Token{Type: token.Pipe, Literal: "|"}
	Gt      = token.Token{Type: token.RedirOut, Literal: ">"}
	Gt1     = token.Token{Type: token.RedirOut, Literal: "1>"}
	Dgt     = token.Token{Type: token.RedirAppend, Literal: ">>"}
	Dgt1    = token.Token{Type: token.RedirAppend, Literal: "1>>"}
	Gt2     = token.Token{Type: token.RedirErr, Literal: "2>"}
	Dgt2    = token.Token{Type: token.RedirErrAppend, Literal: "2>>"}
	EOFTok  = token.Token{Type: token.EOF, Literal: ""}
)
