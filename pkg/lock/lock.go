package lock

import (
	"fmt"
	"strings"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

const (
	lockFilePath = "/tmp/edgectl.lock"
)

// ErrLockHeld is returned when an attempt is made to acquire a lock that is already held.
var ErrLockHeld = fmt.Errorf("lock already held at %s", lockFilePath)

// Acquire attempts to acquire a remote file-based lock.
// It returns ErrLockHeld if the lock is already held.
func Acquire(execCtx execcontext.Context, runner ssh.Runner) error {
	cmd := fmt.Sprintf("mkdir %s", lockFilePath)
	_, stderr, err := runner.Run(execCtx, cmd)
	if err != nil {
		if strings.Contains(stderr, "File exists") || strings.Contains(stderr, "cannot create directory") {
			return ErrLockHeld
		}
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	return nil
}

// Release attempts to release a remote file-based lock.
// It succeeds even if the lock does not exist.
func Release(execCtx execcontext.Context, runner ssh.Runner) error {
	cmd := fmt.Sprintf("rmdir %s", lockFilePath)
	_, stderr, err := runner.Run(execCtx, cmd) // Capture stderr
	if err != nil {
		// If the directory doesn't exist, it's already released, so we don't treat it as an error.
		if strings.Contains(stderr, "No such file or directory") || strings.Contains(stderr, "not a directory") {
			return nil
		}
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}
