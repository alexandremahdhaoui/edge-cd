package ssh

// Runner defines the interface for executing commands on a remote host.
type Runner interface {
	Run(cmd string) (stdout, stderr string, err error)
}
