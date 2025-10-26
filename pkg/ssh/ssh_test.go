package ssh

import (
	"errors"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/execcontext"
)

func TestMockSSHClient(t *testing.T) {
	mockRunner := NewMockRunner()
	ctx := execcontext.New(make(map[string]string), []string{})

	// Test default behavior
	stdout, stderr, err := mockRunner.Run(ctx, "echo hello")
	if stdout != "" || stderr != "" || err != nil {
		t.Errorf("Expected empty output and nil error for default, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun("echo hello"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(1); err != nil {
		t.Error(err)
	}

	// Test with specific response
	mockRunner.SetResponse("ls -l", "file1\nfile2\n", "", nil)
	stdout, stderr, err = mockRunner.Run(ctx, "ls -l")
	if stdout != "file1\nfile2\n" || stderr != "" || err != nil {
		t.Errorf("Expected specific output, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun("ls -l"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(2); err != nil {
		t.Error(err)
	}

	// Test with error response
	mockErr := errors.New("permission denied")
	mockRunner.SetResponse("rm /root/file", "", "rm: permission denied\n", mockErr)
	stdout, stderr, err = mockRunner.Run(ctx, "rm /root/file")
	if stdout != "" || stderr != "rm: permission denied\n" || err != mockErr {
		t.Errorf("Expected specific error, got stdout: %q, stderr: %q, err: %v", stdout, stderr, err)
	}
	if err := mockRunner.AssertCommandRun("rm /root/file"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertNumberOfCommandsRun(3); err != nil {
		t.Error(err)
	}

	// Test command not run
	if err := mockRunner.AssertCommandRun("non-existent command"); err == nil {
		t.Error("Expected error for non-existent command, got nil")
	}
}
