package lexer

import (
	"strings"

	"github.com/tinyshell/internal/token"
)

var escapable = map[byte]bool{
	'\\': true,
	'$':  true,
	'"':  true,
}

type Lexer struct {
	input        string
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
	ch           byte // current char under examination
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	if l.ch == 0 {
		return token.Token{Type: token.EOF, Literal: ""}
	}

	res := strings.Builder{}
	for {
		switch l.ch {
		case 0, ' ', '\t', '\n', '\r':
			return token.Token{Type: token.Word, Literal: res.String()}
		case '\'':
			l.readChar()
			if !l.readInsideSingleQuotes(&res) {
				return newToken(token.ILLEGAL, '\'')
			}
		case '"':
			l.readChar()
			if !l.readInsideDoubleQuotes(&res) {
				return newToken(token.ILLEGAL, '"')
			}
		case '\\':
			l.readChar() // skip the backslash
			if l.ch == 0 {
				res.WriteByte('\\') // trailing backslash stays
				continue
			}
			res.WriteByte(l.ch) // next char is literal, whatever it is
			l.readChar()
		case '>':
			if res.Len() > 0 {
				return token.Token{Type: token.Word, Literal: res.String()}
			}
			if l.peekChar() == '>' {
				l.readChar()
				l.readChar()
				return token.Token{Type: token.RedirAppend, Literal: ">>"}
			}

			l.readChar()
			return token.Token{Type: token.RedirOut, Literal: ">"}
		case '1':
			if l.peekChar() == '>' {
				if res.Len() > 0 {
					literal := res.String() + string(l.ch)
					l.readChar()
					return token.Token{Type: token.Word, Literal: literal}
				}
				tok := token.Token{}
				l.readChar()
				if l.peekChar() == '>' {
					tok.Literal = "1>>"
					tok.Type = token.RedirAppend
					l.readChar()
					l.readChar()
					return tok
				}
				tok.Literal = "1>"
				tok.Type = token.RedirOut
				l.readChar()
				return tok
			} else {
				// not an operator, continue as normal
				res.WriteByte(l.ch)
				l.readChar()
			}
		case '2':
			if l.peekChar() == '>' {
				if res.Len() > 0 {
					literal := res.String() + string(l.ch)
					l.readChar()
					return token.Token{Type: token.Word, Literal: literal}
				}
				tok := token.Token{}
				l.readChar()
				if l.peekChar() == '>' {
					tok.Literal = "2>>"
					tok.Type = token.RedirErrAppend
					l.readChar()
					l.readChar()
					return tok
				}
				tok.Literal = "2>"
				tok.Type = token.RedirErr
				l.readChar()
				return tok
			} else {
				// not an operator, continue as normal
				res.WriteByte(l.ch)
				l.readChar()
			}
		case '|':
			if res.Len() > 0 {
				return token.Token{Type: token.Word, Literal: res.String()}
			}
			l.readChar()
			return token.Token{Type: token.Pipe, Literal: "|"}
		default:
			res.WriteByte(l.ch)
			l.readChar()
		}
	}
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	} else {
		return l.input[l.readPosition]
	}
}

func (l *Lexer) readInsideSingleQuotes(res *strings.Builder) bool {
	for l.ch != '\'' {
		if l.ch == 0 {
			return false
		}
		res.WriteByte(l.ch)
		l.readChar()
	}
	l.readChar() // skip closing '
	return true
}

func (l *Lexer) readInsideDoubleQuotes(res *strings.Builder) bool {
	for l.ch != '"' {
		if l.ch == 0 {
			return false
		}
		if l.ch == '\\' && escapable[l.peekChar()] {
			l.readChar() // drop the backslash, keep the next char
		}
		res.WriteByte(l.ch)
		l.readChar()
	}
	l.readChar() // skip closing "
	return true
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
