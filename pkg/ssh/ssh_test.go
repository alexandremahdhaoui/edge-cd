package ssh

import (
	"errors"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
)

func TestMockSSHClient(t *testing.T) {
	mockRunner := NewMockRunner()
	ctx := execcontext.New(make(map[string]string), []string{})

	// Test default behavior - commands are now formatted with FormatCmd
	echoCmd := execcontext.FormatCmd(ctx, "echo", "hello")
	stdout, stderr, err := mockRunner.Run(ctx, "echo", "hello")
	if stdout != "" || stderr != "" || err != nil {
		t.Errorf("Expected empty output and nil error for default, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun(echoCmd); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(1); err != nil {
		t.Error(err)
	}

	// Test with specific response
	lsCmd := execcontext.FormatCmd(ctx, "ls", "-l")
	mockRunner.SetResponse(lsCmd, "file1\nfile2\n", "", nil)
	stdout, stderr, err = mockRunner.Run(ctx, "ls", "-l")
	if stdout != "file1\nfile2\n" || stderr != "" || err != nil {
		t.Errorf("Expected specific output, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun(lsCmd); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(2); err != nil {
		t.Error(err)
	}

	// Test with error response
	rmCmd := execcontext.FormatCmd(ctx, "rm", "/root/file")
	mockErr := errors.New("permission denied")
	mockRunner.SetResponse(rmCmd, "", "rm: permission denied\n", mockErr)
	stdout, stderr, err = mockRunner.Run(ctx, "rm", "/root/file")
	if stdout != "" || stderr != "rm: permission denied\n" || err != mockErr {
		t.Errorf("Expected specific error, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun(rmCmd); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(3); err != nil {
		t.Error(err)
	}

	// Test command not run
	nonExistentCmd := execcontext.FormatCmd(ctx, "non-existent", "command")
	if err := mockRunner.AssertCommandRun(nonExistentCmd); err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}
