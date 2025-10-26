package ssh

import (
	"github.com/alexandremahdhaoui/edge-cd/pkg/execution"
)

// Runner defines the interface for executing commands on a remote host.
type Runner interface {
	Run(cmd string) (stdout, stderr string, err error)
	RunWithEnvs(cmd string) (stdout, stderr string, err error)
	RunWithBuilder(builder *execution.CommandBuilder) (stdout, stderr string, err error)
}
