package pkgmgr_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/alexandremahdhaoui/edge-cd/pkg/pkgmgr"
	"github.com/alexandremahdhaoui/edge-cd/pkg/ssh"
)

func TestInstallPackages(t *testing.T) {
	mockRunner := ssh.NewMockRunner()
	packages := []string{"git", "yq"}

	// Test new install
	mockRunner.SetResponse("opkg update", "", "", nil)
	mockRunner.SetResponse(
		"opkg list-installed | grep -q '^git '",
		"",
		"",
		errors.New("exit status 1"),
	) // Not installed
	mockRunner.SetResponse("opkg install git", "", "", nil)
	mockRunner.SetResponse(
		"opkg list-installed | grep -q '^yq '",
		"",
		"",
		errors.New("exit status 1"),
	) // Not installed
	mockRunner.SetResponse("opkg install yq", "", "", nil)

	err := pkgmgr.InstallPackages(mockRunner, packages)
	if err != nil {
		t.Errorf("Expected no error on new install, got %v", err)
	}
	if err := mockRunner.AssertCommandRun("opkg update"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg list-installed | grep -q '^git '"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg install git"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg list-installed | grep -q '^yq '"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg install yq"); err != nil {
		t.Error(err)
	}

	// Test existing install (idempotency)
	mockRunner = ssh.NewMockRunner() // Reset mock
	mockRunner.SetResponse("opkg update", "", "", nil)
	mockRunner.SetResponse(
		"opkg list-installed | grep -q '^git '",
		"git - 2.34.1-1\n",
		"",
		nil,
	) // Already installed
	mockRunner.SetResponse(
		"opkg list-installed | grep -q '^yq '",
		"",
		"",
		errors.New("exit status 1"),
	) // Not installed
	mockRunner.SetResponse("opkg install yq", "", "", nil)

	err = pkgmgr.InstallPackages(mockRunner, packages)
	if err != nil {
		t.Errorf("Expected no error on existing install, got %v", err)
	}
	if err := mockRunner.AssertCommandRun("opkg update"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg list-installed | grep -q '^git '"); err != nil {
		t.Error(err)
	}
	// The following command should NOT be run, so we assert it was NOT run.
	if err := mockRunner.AssertCommandRun("opkg install git"); err == nil {
		t.Error("Expected 'opkg install git' not to be run for existing package")
	}
	if err := mockRunner.AssertCommandRun("opkg list-installed | grep -q '^yq '"); err != nil {
		t.Error(err)
	}
	if err := mockRunner.AssertCommandRun("opkg install yq"); err != nil {
		t.Error(err)
	}

	// Test installation failure
	mockRunner = ssh.NewMockRunner() // Reset mock
	mockRunner.SetResponse("opkg update", "", "", nil)
	mockRunner.SetResponse(
		"opkg list-installed | grep -q '^git '",
		"",
		"",
		errors.New("exit status 1"),
	) // Not installed
	mockRunner.SetResponse(
		"opkg install git",
		"",
		"Error installing git\n",
		errors.New("exit status 1"),
	)

	err = pkgmgr.InstallPackages(mockRunner, packages[:1]) // Only test git
	if err == nil {
		t.Error("Expected error on installation failure, got nil")
	}
	if err != nil && !strings.Contains(err.Error(), "failed to install package git") {
		t.Errorf("Expected specific error message, got %v", err)
	}
}
