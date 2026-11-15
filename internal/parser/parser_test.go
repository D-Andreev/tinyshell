package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	. "github.com/tinyshell/internal/testutil"
	"github.com/tinyshell/internal/token"
)

func toks(ts ...token.Token) []token.Token { return ts }

func cmd(args ...string) Command { return Command{Args: args} }

func (c Command) with(rs ...Redirect) Command {
	c.Redirs = append(c.Redirs, rs...)
	return c
}

func out(f string) Redirect    { return Redirect{FD: 1, Append: false, Target: f} }
func appnd(f string) Redirect  { return Redirect{FD: 1, Append: true, Target: f} }
func errOut(f string) Redirect { return Redirect{FD: 2, Append: false, Target: f} }
func errApp(f string) Redirect { return Redirect{FD: 2, Append: true, Target: f} }

func pl(cs ...Command) *Pipeline { return &Pipeline{Commands: cs} }

type parseCase struct {
	name  string
	input []token.Token
	want  *Pipeline
}

func runParseCases(t *testing.T, cases []parseCase) {
	t.Helper()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser()
			require.NoError(t, p.Parse(tc.input))
			assert.Equal(t, tc.want, p.Pipeline)
		})
	}
}

func commands(p *Pipeline) []Command {
	if p == nil {
		return nil
	}
	return p.Commands
}

// 1. Simple commands
func TestParseSimple(t *testing.T) {
	runParseCases(t, []parseCase{
		{"ls", toks(W("ls")), pl(cmd("ls"))},
		{"ls -la /tmp", toks(W("ls"), W("-la"), W("/tmp")), pl(cmd("ls", "-la", "/tmp"))},
		{`echo "hello world"`, toks(W("echo"), W("hello world")), pl(cmd("echo", "hello world"))},
		{`echo ""`, toks(W("echo"), W("")), pl(cmd("echo", ""))},
		{"trailing EOF token is ignored", toks(W("ls"), EOFTok), pl(cmd("ls"))},
	})
}

// 2. Empty input: no commands, no error
func TestParseEmpty(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input []token.Token
	}{
		{"no tokens", toks()},
		{"only EOF", toks(EOFTok)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser()
			require.NoError(t, p.Parse(tc.input))
			assert.Empty(t, commands(p.Pipeline))
		})
	}
}

// 3. Redirections
func TestParseRedirections(t *testing.T) {
	runParseCases(t, []parseCase{
		{"echo hi > out.txt",
			toks(W("echo"), W("hi"), Gt, W("out.txt")),
			pl(cmd("echo", "hi").with(out("out.txt")))},
		{"echo hi 1> out.txt",
			toks(W("echo"), W("hi"), Gt1, W("out.txt")),
			pl(cmd("echo", "hi").with(out("out.txt")))},
		{"echo hi >> log.txt",
			toks(W("echo"), W("hi"), Dgt, W("log.txt")),
			pl(cmd("echo", "hi").with(appnd("log.txt")))},
		{"ls /nope 2> err.txt",
			toks(W("ls"), W("/nope"), Gt2, W("err.txt")),
			pl(cmd("ls", "/nope").with(errOut("err.txt")))},
		{"ls /nope 2>> err.txt",
			toks(W("ls"), W("/nope"), Dgt2, W("err.txt")),
			pl(cmd("ls", "/nope").with(errApp("err.txt")))},
		{"ls /nope /tmp > out.txt 2> err.txt",
			toks(W("ls"), W("/nope"), W("/tmp"), Gt, W("out.txt"), Gt2, W("err.txt")),
			pl(cmd("ls", "/nope", "/tmp").with(out("out.txt"), errOut("err.txt")))},
		{`echo hi > "my file.txt"`,
			toks(W("echo"), W("hi"), Gt, W("my file.txt")),
			pl(cmd("echo", "hi").with(out("my file.txt")))},
		// both redirects are kept, in order (bash creates f1, then writes to f2)
		{"echo a > f1 > f2",
			toks(W("echo"), W("a"), Gt, W("f1"), Gt, W("f2")),
			pl(cmd("echo", "a").with(out("f1"), out("f2")))},
	})
}

// 4. A redirect can appear anywhere in a command
func TestParseRedirectPosition(t *testing.T) {
	runParseCases(t, []parseCase{
		{"> out.txt echo hi",
			toks(Gt, W("out.txt"), W("echo"), W("hi")),
			pl(cmd("echo", "hi").with(out("out.txt")))},
		{"echo > out.txt hi",
			toks(W("echo"), Gt, W("out.txt"), W("hi")),
			pl(cmd("echo", "hi").with(out("out.txt")))},
		{"2> err.txt ls /nope > out.txt",
			toks(Gt2, W("err.txt"), W("ls"), W("/nope"), Gt, W("out.txt")),
			pl(cmd("ls", "/nope").with(errOut("err.txt"), out("out.txt")))},
	})
}

// 5. Pipelines
func TestParsePipelines(t *testing.T) {
	runParseCases(t, []parseCase{
		{"ls | wc -l",
			toks(W("ls"), PipeTok, W("wc"), W("-l")),
			pl(cmd("ls"), cmd("wc", "-l"))},
		{"cat file | grep foo | sort | uniq",
			toks(W("cat"), W("file"), PipeTok, W("grep"), W("foo"), PipeTok, W("sort"), PipeTok, W("uniq")),
			pl(cmd("cat", "file"), cmd("grep", "foo"), cmd("sort"), cmd("uniq"))},
		{`echo "a|b" | wc -c`,
			toks(W("echo"), W("a|b"), PipeTok, W("wc"), W("-c")),
			pl(cmd("echo", "a|b"), cmd("wc", "-c"))},
		{"cat f | sort > sorted.txt",
			toks(W("cat"), W("f"), PipeTok, W("sort"), Gt, W("sorted.txt")),
			pl(cmd("cat", "f"), cmd("sort").with(out("sorted.txt")))},
		{"ls /nope 2> err.txt | wc -l",
			toks(W("ls"), W("/nope"), Gt2, W("err.txt"), PipeTok, W("wc"), W("-l")),
			pl(cmd("ls", "/nope").with(errOut("err.txt")), cmd("wc", "-l"))},
		{"grep x 2> e1 | sort > out 2> e2",
			toks(W("grep"), W("x"), Gt2, W("e1"), PipeTok, W("sort"), Gt, W("out"), Gt2, W("e2")),
			pl(cmd("grep", "x").with(errOut("e1")), cmd("sort").with(out("out"), errOut("e2")))},
	})
}

// 6. A command with only redirects and no words.
// bash accepts "> out.txt" on its own: it creates/truncates the file and runs
// nothing. This test expects the same. If you decide to make it an error
// instead, move these cases to TestParseErrors.
func TestParseRedirectOnly(t *testing.T) {
	runParseCases(t, []parseCase{
		{"> out.txt", toks(Gt, W("out.txt")), pl(cmd().with(out("out.txt")))},
	})
}

// 7. Syntax errors
func TestParseErrors(t *testing.T) {
	cases := []struct {
		name  string
		input []token.Token
	}{
		{"| ls (nothing before PipeTok)", toks(PipeTok, W("ls"))},
		{"ls | (nothing after PipeTok)", toks(W("ls"), PipeTok)},
		{"ls | then EOF", toks(W("ls"), PipeTok, EOFTok)},
		{"ls | | wc (empty middle stage)", toks(W("ls"), PipeTok, PipeTok, W("wc"))},
		{"| alone", toks(PipeTok)},
		{"echo > (no target)", toks(W("echo"), Gt)},
		{"echo > then EOF", toks(W("echo"), Gt, EOFTok)},
		{"echo 2>> (no target)", toks(W("echo"), Dgt2)},
		{"echo > | wc (target is a PipeTok)", toks(W("echo"), Gt, PipeTok, W("wc"))},
		{"echo hi >>> f (target is an operator)", toks(W("echo"), W("hi"), Dgt, Gt, W("f"))},
		{"echo > > f", toks(W("echo"), Gt, Gt, W("f"))},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := NewParser()
			assert.Error(t, p.Parse(tc.input))
		})
	}
}
