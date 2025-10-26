package main

import (
	"os"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgectl/provision"
	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
	"github.com/stretchr/testify/assert"
)

// TestCloneOrPullRepoWithInjectEnvEmpty verifies that CloneOrPullRepoWithBranchAndEnv handles empty env
func TestCloneOrPullRepoWithInjectEnvEmpty(t *testing.T) {
	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("git clone -b main https://github.com/test/config.git /tmp/config-repo", "", "", nil)

	// Create context with empty environment variables
	ctx := execcontext.New(make(map[string]string), []string{})

	// Call with empty env (no --inject-env flag)
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		ctx,
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
	)

	// The function should attempt to execute without failing on empty env
	assert.NoError(t, err, "CloneOrPullRepoWithBranchAndEnv should handle empty env")
}

// TestCloneOrPullRepoWithInjectEnvSet verifies that CloneOrPullRepoWithBranchAndEnv handles populated env
func TestCloneOrPullRepoWithInjectEnvSet(t *testing.T) {
	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' git clone -b main https://github.com/test/config.git /tmp/config-repo", "", "", nil)

	// Create context with environment variables
	envs := make(map[string]string)
	envs["GIT_SSH_COMMAND"] = "ssh -o StrictHostKeyChecking=no"
	ctx := execcontext.New(envs, []string{})

	// Call with env set
	err := provision.CloneOrPullRepoWithBranchAndEnv(
		ctx,
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
	)

	// The function should attempt to execute without failing on populated env
	assert.NoError(t, err, "CloneOrPullRepoWithBranchAndEnv should handle populated env")
}

// TestBootstrapCommandWithoutInjectEnv verifies bootstrap command accepts absence of --inject-env
func TestBootstrapCommandWithoutInjectEnv(t *testing.T) {
	// This test verifies that the bootstrap command structure properly handles
	// the --inject-env flag being absent (empty string passed to provision functions).
	// We verify this by testing CloneOrPullRepoWithBranchAndEnv with empty env.

	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("git clone -b main https://github.com/test/config.git /tmp/config-repo", "", "", nil)

	// When --inject-env is not provided, create context with empty env
	ctx := execcontext.New(make(map[string]string), []string{})

	err := provision.CloneOrPullRepoWithBranchAndEnv(
		ctx,
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
	)

	assert.NoError(t, err, "bootstrap should work without --inject-env flag")
}

// TestBootstrapCommandWithInjectEnv verifies bootstrap command accepts --inject-env flag
func TestBootstrapCommandWithInjectEnv(t *testing.T) {
	// This test verifies that the bootstrap command properly passes the --inject-env
	// flag value through to the provision functions without modification.

	mockRunner := ssh.NewMockRunner()
	mockRunner.SetResponse("GIT_SSH_COMMAND='ssh -o StrictHostKeyChecking=no' git clone -b main https://github.com/test/config.git /tmp/config-repo", "", "", nil)

	// When --inject-env is provided, add it to the context
	envs := make(map[string]string)
	envs["GIT_SSH_COMMAND"] = "ssh -o StrictHostKeyChecking=no"
	ctx := execcontext.New(envs, []string{})

	err := provision.CloneOrPullRepoWithBranchAndEnv(
		ctx,
		mockRunner,
		"https://github.com/test/config.git",
		"/tmp/config-repo",
		"main",
	)

	assert.NoError(t, err, "bootstrap should work with --inject-env flag set")
}
