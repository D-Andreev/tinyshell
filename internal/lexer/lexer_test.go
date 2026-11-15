package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/tinyshell/internal/testutil"
	. "github.com/tinyshell/internal/token"
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
			[]Token{W("ls")},
		},
		{
			"flags and path",
			`ls -la /tmp`,
			[]Token{W("ls"), W("-la"), W("/tmp")},
		},
		{
			"two args",
			`echo hello world`,
			[]Token{W("echo"), W("hello"), W("world")},
		},
		{
			"extra spaces collapse",
			`echo    hello      world`,
			[]Token{W("echo"), W("hello"), W("world")},
		},
		{
			"leading and trailing spaces",
			`   ls   `,
			[]Token{W("ls")},
		},
		{
			"tabs separate words",
			"echo\thello\tworld",
			[]Token{W("echo"), W("hello"), W("world")},
		},
		{
			"trailing newline ignored",
			"ls\n",
			[]Token{W("ls")},
		},
		{
			"multiple args to builtin",
			`type echo ls nonexistent`,
			[]Token{W("type"), W("echo"), W("ls"), W("nonexistent")},
		},
		{
			"numeric arg",
			`exit 3`,
			[]Token{W("exit"), W("3")},
		},
		{
			"relative path",
			`cd ../src`,
			[]Token{W("cd"), W("../src")},
		},
		{
			"no args",
			`pwd`,
			[]Token{W("pwd")},
		},
	})
}

func TestLexSingleQuotes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"space kept",
			`echo 'hello world'`,
			[]Token{W("echo"), W("hello world")},
		},
		{
			"many spaces kept",
			`echo 'hello    world'`,
			[]Token{W("echo"), W("hello    world")},
		},
		{
			"adjacent quoted parts join",
			`echo 'it''s'`,
			[]Token{W("echo"), W("its")},
		},
		{
			"backslash literal",
			`echo 'a\nb'`,
			[]Token{W("echo"), W(`a\nb`)},
		},
		{
			"dollar literal",
			`echo '$HOME'`,
			[]Token{W("echo"), W("$HOME")},
		},
		{
			"operators literal",
			`echo 'a | b > c'`,
			[]Token{W("echo"), W("a | b > c")},
		},
		{
			"double quote inside single",
			`echo 'say "hi"'`,
			[]Token{W("echo"), W(`say "hi"`)},
		},
		{
			"empty single quotes",
			`echo ''`,
			[]Token{W("echo"), W("")},
		},
	})
}

func TestLexDoubleQuotes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"space kept",
			`echo "hello world"`,
			[]Token{W("echo"), W("hello world")},
		},
		{
			"escaped quotes",
			`echo "say \"hi\""`,
			[]Token{W("echo"), W(`say "hi"`)},
		},
		{
			"backslash n stays",
			`echo "a\nb"`,
			[]Token{W("echo"), W(`a\nb`)},
		},
		{
			"escaped backslash",
			`echo "back\\slash"`,
			[]Token{W("echo"), W(`back\slash`)},
		},
		{
			"escaped dollar",
			`echo "\$HOME"`,
			[]Token{W("echo"), W("$HOME")},
		},
		{
			"single quote inside double",
			`echo "it's fine"`,
			[]Token{W("echo"), W("it's fine")},
		},
		{
			"operators literal",
			`echo "a | b > c"`,
			[]Token{W("echo"), W("a | b > c")},
		},
		{
			"empty string is a word",
			`echo ""`,
			[]Token{W("echo"), W("")},
		},
		{
			"two empty words",
			`echo "" ""`,
			[]Token{W("echo"), W(""), W("")},
		},
	})
}

func TestLexBackslashEscapes(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"escaped space",
			`echo hello\ world`,
			[]Token{W("echo"), W("hello world")},
		},
		{
			"escaped single quotes",
			`echo \'quoted\'`,
			[]Token{W("echo"), W("'quoted'")},
		},
		{
			"escaped backslash",
			`echo a\\b`,
			[]Token{W("echo"), W(`a\b`)},
		},
		{
			"escaped PipeTok is a word",
			`echo \|`,
			[]Token{W("echo"), W("|")},
		},
		{
			"escaped Gt is a word",
			`echo \>`,
			[]Token{W("echo"), W(">")},
		},
		{
			"escaped double quote",
			`echo \"x\"`,
			[]Token{W("echo"), W(`"x"`)},
		},
		{
			"escaped ordinary char",
			`echo \a`,
			[]Token{W("echo"), W("a")},
		},
		// bash -c keeps a trailing backslash literally. An interactive shell
		// would ask for a continuation line instead; that is a stretch goal.
		{
			"trailing backslash",
			`echo trailing\`,
			[]Token{W("echo"), W(`trailing\`)},
		},
	})
}

// Adjacent quoted and unquoted pieces form a single word.
func TestLexWordConcatenation(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"double single plain",
			`echo "foo"'bar'baz`,
			[]Token{W("echo"), W("foobarbaz")},
		},
		{
			"empty quotes in middle",
			`echo a""b`,
			[]Token{W("echo"), W("ab")},
		},
		{
			"alternating",
			`echo "a"b'c'"d"`,
			[]Token{W("echo"), W("abcd")},
		},
		{
			"quoted space joins",
			`echo hello" "world`,
			[]Token{W("echo"), W("hello world")},
		},
		{
			"quoted command name",
			`'my program' arg1`,
			[]Token{W("my program"), W("arg1")},
		},
	})
}

func TestLexRedirections(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"stdout",
			`echo hi > out.txt`,
			[]Token{W("echo"), W("hi"), Gt, W("out.txt")},
		},
		{
			"append",
			`echo hi >> out.txt`,
			[]Token{W("echo"), W("hi"), Dgt, W("out.txt")},
		},
		{
			"explicit fd 1",
			`echo hi 1> out.txt`,
			[]Token{W("echo"), W("hi"), Gt1, W("out.txt")},
		},
		{
			"explicit fd 1 append",
			`echo hi 1>> out.txt`,
			[]Token{W("echo"), W("hi"), Dgt1, W("out.txt")},
		},
		{
			"stderr",
			`ls /nope 2> err.txt`,
			[]Token{W("ls"), W("/nope"), Gt2, W("err.txt")},
		},
		{
			"stderr append",
			`ls /nope 2>> err.txt`,
			[]Token{W("ls"), W("/nope"), Dgt2, W("err.txt")},
		},
		{
			"both streams",
			`ls /nope /tmp > out.txt 2> err.txt`,
			[]Token{W("ls"), W("/nope"), W("/tmp"), Gt, W("out.txt"), Gt2, W("err.txt")},
		},
		{
			"redirect first",
			`> out.txt echo hi`,
			[]Token{Gt, W("out.txt"), W("echo"), W("hi")},
		},
		{
			"quoted target",
			`echo hi > "my file.txt"`,
			[]Token{W("echo"), W("hi"), Gt, W("my file.txt")},
		},
		{
			"two redirects",
			`echo a > f1 > f2`,
			[]Token{W("echo"), W("a"), Gt, W("f1"), Gt, W("f2")},
		},
	})
}

// Spacing decides where a word ends. A digit is an fd number only when it is
// the whole word and directly precedes a redirect.
func TestLexOperatorSpacing(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"no spaces",
			`echo hi>out.txt`,
			[]Token{W("echo"), W("hi"), Gt, W("out.txt")},
		},
		{
			"space before only",
			`echo hi >out.txt`,
			[]Token{W("echo"), W("hi"), Gt, W("out.txt")},
		},
		{
			"append no spaces",
			`echo hi>>out.txt`,
			[]Token{W("echo"), W("hi"), Dgt, W("out.txt")},
		},
		{
			"a2 is a word",
			`echo a2> f`,
			[]Token{W("echo"), W("a2"), Gt, W("f")},
		},
		{
			"2> attached",
			`echo 2>f`,
			[]Token{W("echo"), Gt2, W("f")},
		},
		{
			"2 then space is a word",
			`echo 2 > f`,
			[]Token{W("echo"), W("2"), Gt, W("f")},
		},
		{
			"quoted Gt is a word",
			`echo ">" f`,
			[]Token{W("echo"), W(">"), W("f")},
		},
		{
			"quoted 2 is a word",
			`echo "2"> f`,
			[]Token{W("echo"), W("2"), Gt, W("f")},
		},
	})
}

func TestLexPipelines(t *testing.T) {
	runLexCases(t, []lexCase{
		{
			"two stages",
			`ls | wc -l`,
			[]Token{W("ls"), PipeTok, W("wc"), W("-l")},
		},
		{
			"four stages",
			`cat file | grep foo | sort | uniq`,
			[]Token{W("cat"), W("file"), PipeTok, W("grep"), W("foo"), PipeTok, W("sort"), PipeTok, W("uniq")},
		},
		{
			"no spaces",
			`ls|wc`,
			[]Token{W("ls"), PipeTok, W("wc")},
		},
		{
			"builtin in pipe",
			`echo hello | wc -c`,
			[]Token{W("echo"), W("hello"), PipeTok, W("wc"), W("-c")},
		},
		{
			"pipe then redirect",
			`cat f | sort > sorted.txt`,
			[]Token{W("cat"), W("f"), PipeTok, W("sort"), Gt, W("sorted.txt")},
		},
		{
			"stderr redirect then pipe",
			`ls /nope 2> err.txt | wc -l`,
			[]Token{W("ls"), W("/nope"), Gt2, W("err.txt"), PipeTok, W("wc"), W("-l")},
		},
		{
			"quoted pipe",
			`echo "a|b" | wc -c`,
			[]Token{W("echo"), W("a|b"), PipeTok, W("wc"), W("-c")},
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
			[]Token{W("echo"), W("$HOME")},
		},
		{
			"braced variable",
			`echo ${HOME}/bin`,
			[]Token{W("echo"), W("${HOME}/bin")},
		},
		{
			"double quoted variable",
			`echo "$HOME"`,
			[]Token{W("echo"), W("$HOME")},
		},
		{
			"single quoted variable",
			`echo '$HOME'`,
			[]Token{W("echo"), W("$HOME")},
		},
		{
			"exit status",
			`echo $?`,
			[]Token{W("echo"), W("$?")},
		},
		{
			"tilde",
			`cd ~`,
			[]Token{W("cd"), W("~")},
		},
		{
			"tilde path",
			`ls ~/docs`,
			[]Token{W("ls"), W("~/docs")},
		},
		{
			"tilde mid-word",
			`echo a~b`,
			[]Token{W("echo"), W("a~b")},
		},
		{
			"quoted tilde",
			`echo "~"`,
			[]Token{W("echo"), W("~")},
		},
		{
			"assignment",
			`export PATH=~/bin:$PATH`,
			[]Token{W("export"), W("PATH=~/bin:$PATH")},
		},
	})
}

// These inputs are valid for the lexer even though the parser would reject
// them. The lexer does not know the grammar.
func TestLexMalformedButTokenizable(t *testing.T) {
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
			[]Token{W("ls"), PipeTok},
		},
		{
			"leading pipe",
			`| ls`,
			[]Token{PipeTok, W("ls")},
		},
		{
			"double pipe",
			`ls | | wc`,
			[]Token{W("ls"), PipeTok, PipeTok, W("wc")},
		},
		{
			"redirect without target",
			`echo >`,
			[]Token{W("echo"), Gt},
		},
		{
			"redirect then pipe",
			`echo > | wc`,
			[]Token{W("echo"), Gt, PipeTok, W("wc")},
		},
		// The lexer takes the longest operator it can: >> then >.
		{
			"triple gt",
			`echo hi >>> f`,
			[]Token{W("echo"), W("hi"), Dgt, Gt, W("f")},
		},
	})
}

// NextToken has no error return yet, so unterminated quotes cannot be
// reported. These cases are kept so they run once the lexer can report errors.
func TestLexUnterminatedQuotes(t *testing.T) {
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
