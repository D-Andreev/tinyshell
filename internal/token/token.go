package token

// TokenType identifies the kind of a Token.
type TokenType int

const (
	// Word is a plain word such as echo, or a quoted string such as "a b".
	// Its Value is the unquoted text, so "a b" yields the value a b.
	Word TokenType = iota

	// Pipe is the | operator. It connects the left command's stdout to the right command's stdin.
	//
	// Example: cat app.log | grep error
	Pipe

	// RedirOut is the > operator. It sends stdout to a file, creating the file
	// or truncating it (emptying it) if it already exists.
	//
	// Example: cat app.log > app-tmp.log
	RedirOut

	// RedirAppend is the >> operator, or 1>> with fd 1 written out. It sends
	// stdout to a file, creating the file if needed and appending to it.
	//
	// Example: cat app.log >> app-tmp.log
	RedirAppend

	// RedirErr is the 2> operator. It sends stderr to a file, creating the file
	// or truncating it if it already exists.
	//
	// Example: ls /nope 2> err.txt
	RedirErr

	// RedirErrAppend is the 2>> operator. It sends stderr to a file, creating
	// the file if needed and appending to it.
	//
	// Example: ls /nope 2>> err.txt
	RedirErrAppend

	// EOF
	EOF

	// Illegal
	ILLEGAL
)

type Token struct {
	Type    TokenType
	Literal string
}
