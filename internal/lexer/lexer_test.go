package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/tinyshell/internal/token"
)

// w returns a Word token with the given value.
func w(s string) Token { return Token{Type: Word, Literal: s} }

var (
	pipe = Token{Type: Pipe, Literal: "|"}
	gt   = Token{Type: RedirOut, Literal: ">"}
	gt1  = Token{Type: RedirOut, Literal: "1>"}
	dgt  = Token{Type: RedirAppend, Literal: ">>"}
	dgt1 = Token{Type: RedirAppend, Literal: "1>>"}
	gt2  = Token{Type: RedirErr, Literal: "2>"}
	dgt2 = Token{Type: RedirErrAppend, Literal: "2>>"}
)

type lexCase struct {
	name  string
	input string
	want  []Token
}

// runLexCases runs each case as a subtest that expects the lexer to succeed.
func runLexCases(t *testing.T, cases []lexCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assertLex(t, tc.input, tc.want)
		})
	}
}

// lexAll calls NextToken until the input is exhausted and returns every token  in order.
func lexAll(input string) []Token {
	l := New(input)
	var toks []Token
	for {
		l.skipWhitespace()
		if l.ch == 0 {
			return toks
		}
		toks = append(toks, l.NextToken())
	}
}

// assertLex checks that input lexes to wanted tokens.
func assertLex(t *testing.T, input string, want []Token) {
	t.Helper()
	got := lexAll(input)
	if len(want) == 0 {
		assert.Empty(t, got, "input: %q", input)
		return
	}
	assert.Equal(t, want, got, "input: %q", input)
}

func TestLexWords(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"single word",
			`ls`,
			[]Token{w("ls")},
		},
		{
			"flags and path",
			`ls -la /tmp`,
			[]Token{w("ls"), w("-la"), w("/tmp")},
		},
		{
			"two args",
			`echo hello world`,
			[]Token{w("echo"), w("hello"), w("world")},
		},
		{
			"extra spaces collapse",
			`echo    hello      world`,
			[]Token{w("echo"), w("hello"), w("world")},
		},
		{
			"leading and trailing spaces",
			`   ls   `,
			[]Token{w("ls")},
		},
		{
			"tabs separate words",
			"echo\thello\tworld",
			[]Token{w("echo"), w("hello"), w("world")},
		},
		{
			"trailing newline ignored",
			"ls\n",
			[]Token{w("ls")},
		},
		{
			"multiple args to builtin",
			`type echo ls nonexistent`,
			[]Token{w("type"), w("echo"), w("ls"), w("nonexistent")},
		},
		{
			"numeric arg",
			`exit 3`,
			[]Token{w("exit"), w("3")},
		},
		{
			"relative path",
			`cd ../src`,
			[]Token{w("cd"), w("../src")},
		},
		{
			"no args",
			`pwd`,
			[]Token{w("pwd")},
		},
	})
}

func TestLexSingleQuotes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"space kept",
			`echo 'hello world'`,
			[]Token{w("echo"), w("hello world")},
		},
		{
			"many spaces kept",
			`echo 'hello    world'`,
			[]Token{w("echo"), w("hello    world")},
		},
		{
			"adjacent quoted parts join",
			`echo 'it''s'`,
			[]Token{w("echo"), w("its")},
		},
		{
			"backslash literal",
			`echo 'a\nb'`,
			[]Token{w("echo"), w(`a\nb`)},
		},
		{
			"dollar literal",
			`echo '$HOME'`,
			[]Token{w("echo"), w("$HOME")},
		},
		{
			"operators literal",
			`echo 'a | b > c'`,
			[]Token{w("echo"), w("a | b > c")},
		},
		{
			"double quote inside single",
			`echo 'say "hi"'`,
			[]Token{w("echo"), w(`say "hi"`)},
		},
		{
			"empty single quotes",
			`echo ''`,
			[]Token{w("echo"), w("")},
		},
	})
}

func TestLexDoubleQuotes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"space kept",
			`echo "hello world"`,
			[]Token{w("echo"), w("hello world")},
		},
		{
			"escaped quotes",
			`echo "say \"hi\""`,
			[]Token{w("echo"), w(`say "hi"`)},
		},
		{
			"backslash n stays",
			`echo "a\nb"`,
			[]Token{w("echo"), w(`a\nb`)},
		},
		{
			"escaped backslash",
			`echo "back\\slash"`,
			[]Token{w("echo"), w(`back\slash`)},
		},
		{
			"escaped dollar",
			`echo "\$HOME"`,
			[]Token{w("echo"), w("$HOME")},
		},
		{
			"single quote inside double",
			`echo "it's fine"`,
			[]Token{w("echo"), w("it's fine")},
		},
		{
			"operators literal",
			`echo "a | b > c"`,
			[]Token{w("echo"), w("a | b > c")},
		},
		{
			"empty string is a word",
			`echo ""`,
			[]Token{w("echo"), w("")},
		},
		{
			"two empty words",
			`echo "" ""`,
			[]Token{w("echo"), w(""), w("")},
		},
	})
}

func TestLexBackslashEscapes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"escaped space",
			`echo hello\ world`,
			[]Token{w("echo"), w("hello world")},
		},
		{
			"escaped single quotes",
			`echo \'quoted\'`,
			[]Token{w("echo"), w("'quoted'")},
		},
		{
			"escaped backslash",
			`echo a\\b`,
			[]Token{w("echo"), w(`a\b`)},
		},
		{
			"escaped pipe is a word",
			`echo \|`,
			[]Token{w("echo"), w("|")},
		},
		{
			"escaped gt is a word",
			`echo \>`,
			[]Token{w("echo"), w(">")},
		},
		{
			"escaped double quote",
			`echo \"x\"`,
			[]Token{w("echo"), w(`"x"`)},
		},
		{
			"escaped ordinary char",
			`echo \a`,
			[]Token{w("echo"), w("a")},
		},
		// bash -c keeps a trailing backslash literally. An interactive shell
		// would ask for a continuation line instead; that is a stretch goal.
		{
			"trailing backslash",
			`echo trailing\`,
			[]Token{w("echo"), w(`trailing\`)},
		},
	})
}

// Adjacent quoted and unquoted pieces form a single word.
func TestLexWordConcatenation(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"double single plain",
			`echo "foo"'bar'baz`,
			[]Token{w("echo"), w("foobarbaz")},
		},
		{
			"empty quotes in middle",
			`echo a""b`,
			[]Token{w("echo"), w("ab")},
		},
		{
			"alternating",
			`echo "a"b'c'"d"`,
			[]Token{w("echo"), w("abcd")},
		},
		{
			"quoted space joins",
			`echo hello" "world`,
			[]Token{w("echo"), w("hello world")},
		},
		{
			"quoted command name",
			`'my program' arg1`,
			[]Token{w("my program"), w("arg1")},
		},
	})
}

func TestLexRedirections(t *testing.T) {
	t.Skip("not implemented yet: lexer does not tokenize operators")
	runLexCases(t, []lexCase{
		{
			"stdout",
			`echo hi > out.txt`,
			[]Token{w("echo"), w("hi"), gt, w("out.txt")},
		},
		{
			"append",
			`echo hi >> out.txt`,
			[]Token{w("echo"), w("hi"), dgt, w("out.txt")},
		},
		{
			"explicit fd 1",
			`echo hi 1> out.txt`,
			[]Token{w("echo"), w("hi"), gt1, w("out.txt")},
		},
		{
			"explicit fd 1 append",
			`echo hi 1>> out.txt`,
			[]Token{w("echo"), w("hi"), dgt1, w("out.txt")},
		},
		{
			"stderr",
			`ls /nope 2> err.txt`,
			[]Token{w("ls"), w("/nope"), gt2, w("err.txt")},
		},
		{
			"stderr append",
			`ls /nope 2>> err.txt`,
			[]Token{w("ls"), w("/nope"), dgt2, w("err.txt")},
		},
		{
			"both streams",
			`ls /nope /tmp > out.txt 2> err.txt`,
			[]Token{w("ls"), w("/nope"), w("/tmp"), gt, w("out.txt"), gt2, w("err.txt")},
		},
		{
			"redirect first",
			`> out.txt echo hi`,
			[]Token{gt, w("out.txt"), w("echo"), w("hi")},
		},
		{
			"quoted target",
			`echo hi > "my file.txt"`,
			[]Token{w("echo"), w("hi"), gt, w("my file.txt")},
		},
		{
			"two redirects",
			`echo a > f1 > f2`,
			[]Token{w("echo"), w("a"), gt, w("f1"), gt, w("f2")},
		},
	})
}

// Spacing decides where a word ends. A digit is an fd number only when it is
// the whole word and directly precedes a redirect.
func TestLexOperatorSpacing(t *testing.T) {
	t.Skip("not implemented yet: lexer does not tokenize operators")
	runLexCases(t, []lexCase{
		{
			"no spaces",
			`echo hi>out.txt`,
			[]Token{w("echo"), w("hi"), gt, w("out.txt")},
		},
		{
			"space before only",
			`echo hi >out.txt`,
			[]Token{w("echo"), w("hi"), gt, w("out.txt")},
		},
		{
			"append no spaces",
			`echo hi>>out.txt`,
			[]Token{w("echo"), w("hi"), dgt, w("out.txt")},
		},
		{
			"a2 is a word",
			`echo a2> f`,
			[]Token{w("echo"), w("a2"), gt, w("f")},
		},
		{
			"2> attached",
			`echo 2>f`,
			[]Token{w("echo"), gt2, w("f")},
		},
		{
			"2 then space is a word",
			`echo 2 > f`,
			[]Token{w("echo"), w("2"), gt, w("f")},
		},
		{
			"quoted gt is a word",
			`echo ">" f`,
			[]Token{w("echo"), w(">"), w("f")},
		},
		{
			"quoted 2 is a word",
			`echo "2"> f`,
			[]Token{w("echo"), w("2"), gt, w("f")},
		},
	})
}

func TestLexPipelines(t *testing.T) {
	t.Skip("not implemented yet: lexer does not tokenize operators")
	runLexCases(t, []lexCase{
		{
			"two stages",
			`ls | wc -l`,
			[]Token{w("ls"), pipe, w("wc"), w("-l")},
		},
		{
			"four stages",
			`cat file | grep foo | sort | uniq`,
			[]Token{w("cat"), w("file"), pipe, w("grep"), w("foo"), pipe, w("sort"), pipe, w("uniq")},
		},
		{
			"no spaces",
			`ls|wc`,
			[]Token{w("ls"), pipe, w("wc")},
		},
		{
			"builtin in pipe",
			`echo hello | wc -c`,
			[]Token{w("echo"), w("hello"), pipe, w("wc"), w("-c")},
		},
		{
			"pipe then redirect",
			`cat f | sort > sorted.txt`,
			[]Token{w("cat"), w("f"), pipe, w("sort"), gt, w("sorted.txt")},
		},
		{
			"stderr redirect then pipe",
			`ls /nope 2> err.txt | wc -l`,
			[]Token{w("ls"), w("/nope"), gt2, w("err.txt"), pipe, w("wc"), w("-l")},
		},
		{
			"quoted pipe",
			`echo "a|b" | wc -c`,
			[]Token{w("echo"), w("a|b"), pipe, w("wc"), w("-c")},
		},
	})
}

// The lexer does not expand variables or tildes; expansion happens after
// parsing. These cases only check that the text is kept as written.
//
// Quoting is not recorded yet, so "$HOME" and '$HOME' produce the same Value.
// Add assertions for the quoting information once the Token design settles.
func TestLexUnexpandedWords(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"variable",
			`echo $HOME`,
			[]Token{w("echo"), w("$HOME")},
		},
		{
			"braced variable",
			`echo ${HOME}/bin`,
			[]Token{w("echo"), w("${HOME}/bin")},
		},
		{
			"double quoted variable",
			`echo "$HOME"`,
			[]Token{w("echo"), w("$HOME")},
		},
		{
			"single quoted variable",
			`echo '$HOME'`,
			[]Token{w("echo"), w("$HOME")},
		},
		{
			"exit status",
			`echo $?`,
			[]Token{w("echo"), w("$?")},
		},
		{
			"tilde",
			`cd ~`,
			[]Token{w("cd"), w("~")},
		},
		{
			"tilde path",
			`ls ~/docs`,
			[]Token{w("ls"), w("~/docs")},
		},
		{
			"tilde mid-word",
			`echo a~b`,
			[]Token{w("echo"), w("a~b")},
		},
		{
			"quoted tilde",
			`echo "~"`,
			[]Token{w("echo"), w("~")},
		},
		{
			"assignment",
			`export PATH=~/bin:$PATH`,
			[]Token{w("export"), w("PATH=~/bin:$PATH")},
		},
	})
}

// These inputs are valid for the lexer even though the parser would reject
// them. The lexer does not know the grammar.
func TestLexMalformedButTokenizable(t *testing.T) {
	t.Skip("not implemented yet: lexer does not tokenize operators")
	runLexCases(t, []lexCase{
		{
			"empty line",
			``,
			nil,
		},
		{
			"only spaces",
			`     `,
			nil,
		},
		{
			"only newline",
			"\n",
			nil,
		},
		{
			"trailing pipe",
			`ls |`,
			[]Token{w("ls"), pipe},
		},
		{
			"leading pipe",
			`| ls`,
			[]Token{pipe, w("ls")},
		},
		{
			"double pipe",
			`ls | | wc`,
			[]Token{w("ls"), pipe, pipe, w("wc")},
		},
		{
			"redirect without target",
			`echo >`,
			[]Token{w("echo"), gt},
		},
		{
			"redirect then pipe",
			`echo > | wc`,
			[]Token{w("echo"), gt, pipe, w("wc")},
		},
		// The lexer takes the longest operator it can: >> then >.
		{
			"triple gt",
			`echo hi >>> f`,
			[]Token{w("echo"), w("hi"), dgt, gt, w("f")},
		},
	})
}

// NextToken has no error return yet, so unterminated quotes cannot be
// reported. These cases are kept so they run once the lexer can report errors.
func TestLexUnterminatedQuotes(t *testing.T) {
	t.Skip("lexer has no error reporting: New and NextToken do not return errors")

	cases := []struct {
		name  string
		input string
	}{
		{
			"unclosed double quote",
			`echo "unclosed`,
		},
		{
			"unclosed single quote",
			`echo 'unclosed`,
		},
		{
			"escaped quote leaves it unclosed",
			`echo "abc\"`,
		},
		{
			"unclosed after closed",
			`echo "ok" 'not`,
		},
		{
			"single quote inside double does not close",
			`echo "it's`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lexAll(tc.input)
		})
	}
}
