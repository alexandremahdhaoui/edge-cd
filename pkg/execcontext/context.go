package execcontext

import (
	"fmt"
	"maps"
	"os/exec"
	"strings"
)

type Context interface {
	Envs() map[string]string
	PrependCmd() []string
}

func New(envs map[string]string, prependCmd []string) Context {
	return &context{
		prependCmd: prependCmd,
		envs:       envs,
	}
}

type context struct {
	envs       map[string]string
	prependCmd []string
}

// Envs implements Context.
func (c *context) Envs() map[string]string {
	out := make(map[string]string, len(c.envs))
	maps.Copy(out, c.envs)
	return out
}

// PrependCmd implements Context.
func (c *context) PrependCmd() []string {
	out := make([]string, len(c.prependCmd))
	copy(out, c.prependCmd)
	return out
}

func ApplyToCmd(ctx Context, cmd *exec.Cmd) {
	for k, v := range ctx.Envs() {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	prependCmd := ctx.PrependCmd()
	if len(prependCmd) < 1 {
		return
	}

	var tmpArgs []string
	if len(prependCmd) > 0 {
		tmpArgs = prependCmd[1:]
	}

	tmpCmd := exec.Command(prependCmd[0], tmpArgs...)
	cmd.Path = tmpCmd.Path
	cmd.Args = append(tmpCmd.Args, cmd.Args...)
}

func FormatCmd(ctx Context, cmd ...string) string {
	out := ""

	// Add environment variables first (without quoting the entire assignment)
	for k, v := range ctx.Envs() {
		envStr := fmt.Sprintf("%s=%q", k, v)
		out = fmt.Sprintf("%s%s ", out, envStr)
	}

	// Add prepend command
	for _, s := range ctx.PrependCmd() {
		out = safelyAppendToCmd(out, s)
	}

	// Add the actual command
	for _, s := range cmd {
		out = safelyAppendToCmd(out, s)
	}

	return strings.TrimSpace(out)
}

var unquottable = map[string]struct{}{
	"&&": {},
	"||": {},
	";":  {},
	":":  {},
	"&":  {},
}

func safelyAppendToCmd(cmd string, s string) string {
	if _, ok := unquottable[s]; ok {
		return fmt.Sprintf("%s%s ", cmd, s)
	}
	return fmt.Sprintf("%s%q ", cmd, s)
}
