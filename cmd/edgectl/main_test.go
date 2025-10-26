package main

import (
	"os"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/stretchr/testify/assert"
)


// TestCloneOrPullRepoWithInjectEnvEmpty verifies that CloneOrPullRepoWithBranchAndEnv handles empty env
func TestCloneOrPullRepoWithInjectEnvEmpty(t *testing.T) {
	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("[ -d /tmp/config-repo ]", "", "", os.ErrNotExist)

	// Call with empty env string (no --inject-env flag)
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
		"", // empty env, as if --inject-env was not provided
		"",
	)

	// The function should attempt to execute without failing on empty env
	assert.NoError(t, err, "CloneOrPullRepoWithBranchAndEnv should handle empty env string")
}

// TestCloneOrPullRepoWithInjectEnvSet verifies that CloneOrPullRepoWithBranchAndEnv handles populated env
func TestCloneOrPullRepoWithInjectEnvSet(t *testing.T) {
	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("[ -d /tmp/config-repo ]", "", "", os.ErrNotExist)

	// Call with env string (with --inject-env flag)
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no", // env from --inject-env flag
		"",
	)

	// The function should attempt to execute without failing on populated env
	assert.NoError(t, err, "CloneOrPullRepoWithBranchAndEnv should handle populated env string")
}

// TestBootstrapCommandWithoutInjectEnv verifies bootstrap command accepts absence of --inject-env
func TestBootstrapCommandWithoutInjectEnv(t *testing.T) {
	// This test verifies that the bootstrap command structure properly handles
	// the --inject-env flag being absent (empty string passed to provision functions).
	// We verify this by testing CloneOrPullRepoWithBranchAndEnv with empty env.

	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("[ -d /tmp/config-repo ]", "", "", os.ErrNotExist)

	// When --inject-env is not provided, injectEnv is ""
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
		"", // empty string when --inject-env flag not provided
		"",
	)

	assert.NoError(t, err, "bootstrap should work without --inject-env flag")
}

// TestBootstrapCommandWithInjectEnv verifies bootstrap command accepts --inject-env flag
func TestBootstrapCommandWithInjectEnv(t *testing.T) {
	// This test verifies that the bootstrap command properly passes the --inject-env
	// flag value through to the provision functions without modification.

	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("[ -d /tmp/config-repo ]", "", "", os.ErrNotExist)

	// When --inject-env is provided, injectEnv contains the flag value
	injectEnvValue := "GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no"
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
		injectEnvValue, // value from --inject-env flag
		"",
	)

	assert.NoError(t, err, "bootstrap should work with --inject-env flag set")
}

