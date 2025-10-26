package execution

import (
	"fmt"
	"maps"
	"os/exec"
	"strings"
)

func ApplyContext(ctx Context, cmd *exec.Cmd) {
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
	inner := ""
	for _, s := range append(ctx.PrependCmd(), cmd...) {
		inner = fmt.Sprintf("%s%q ", inner, s)
	}
	inner = strings.TrimSpace(inner)

	outter := make([]string, 0)
	for k, v := range ctx.Envs() {
		outter = append(outter, fmt.Sprintf("%s=%s", k, v))
	}
	outter = append(outter, "sh", "-c", inner)

	out := ""
	for _, s := range outter {
		out = fmt.Sprintf("%s%q ", out, s)
	}
	out = strings.TrimSpace(out)

	return out
}

type Context interface {
	WithEnv(key, value string) Context
	WithPrependCmd(cmd ...string) Context
	Envs() map[string]string
	PrependCmd() []string
}

func NewContext() Context {
	return &context{}
}

func NewContextFrom(ctx Context) Context {
	return &context{
		prependCmd: ctx.PrependCmd(),
		envs:       ctx.Envs(),
	}
}

type context struct {
	prependCmd []string
	envs       map[string]string
}

// WithEnv implements Context.
func (c *context) WithEnv(key string, value string) Context {
	if c.envs == nil {
		c.envs = make(map[string]string)
	}
	c.envs[key] = value
	return c
}

// WithPrependCmd implements Context.
func (c *context) WithPrependCmd(prependCmd ...string) Context {
	c.prependCmd = prependCmd
	return c
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
