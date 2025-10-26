package ssh

import (
	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
)

// Runner defines the interface for executing commands on a remote host.
type Runner interface {
	Run(ctx execcontext.Context, cmd ...string) (stdout, stderr string, err error)
}
