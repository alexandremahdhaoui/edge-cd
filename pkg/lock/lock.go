package lock

import (
	"errors"
	"fmt"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/alexandremahdhaoui/tooling/pkg/flaterrors"
)

const (
	lockFilePath = "/tmp/edgectl.lock"
)

var (
	// ErrLockHeld is returned when an attempt is made to acquire a lock that is already held.
	ErrLockHeld      = fmt.Errorf("lock already held at %s", lockFilePath)
	errAcquireLock   = errors.New("failed to acquire lock")
	errReleaseLock   = errors.New("failed to release lock")
)

// Acquire attempts to acquire a remote file-based lock.
// It returns ErrLockHeld if the lock is already held.
func Acquire(execCtx execcontext.Context, runner ssh.Runner) error {
	_, stderr, err := runner.Run(execCtx, "mkdir", lockFilePath)
	if err != nil {
		if strings.Contains(stderr, "File exists") || strings.Contains(stderr, "cannot create directory") {
			return ErrLockHeld
		}
		return flaterrors.Join(err, errAcquireLock)
	}
	return nil
}

// Release attempts to release a remote file-based lock.
// It succeeds even if the lock does not exist.
func Release(execCtx execcontext.Context, runner ssh.Runner) error {
	_, stderr, err := runner.Run(execCtx, "rmdir", lockFilePath) // Capture stderr
	if err != nil {
		// If the directory doesn't exist, it's already released, so we don't treat it as an error.
		if strings.Contains(stderr, "No such file or directory") || strings.Contains(stderr, "not a directory") {
			return nil
		}
		return flaterrors.Join(err, errReleaseLock)
	}
	return nil
}
