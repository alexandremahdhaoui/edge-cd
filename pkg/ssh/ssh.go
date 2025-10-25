package ssh

import (
	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
)

// Runner defines the interface for executing commands on a remote host.
type Runner interface {
	Run(cmd string) (stdout, stderr string, err error)
}

// BuilderRunner defines the interface for executing commands built with CommandBuilder
// on a remote host. This allows proper composition of prepend commands and environment
// variables for remote execution over SSH.
type BuilderRunner interface {
	RunWithBuilder(builder *execution.CommandBuilder) (stdout, stderr string, err error)
}
