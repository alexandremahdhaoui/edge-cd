package reconcile

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/config"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/files"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/git"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/pkgmgr"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/runtime"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/svcmgr"
	"github.com/alexandremahdhaoui/edge-cd/pkg/userconfig"
)

func TestNewReconciler(t *testing.T) {
	cfg := &config.Config{}
	gitMgr := &git.MockRepoManager{}
	pkgMgr := &pkgmgr.MockPackageManager{}
	svcMgr := &svcmgr.MockServiceManager{}
	fileRec := &files.MockFileReconciler{}

	r := NewReconciler(cfg, gitMgr, pkgMgr, svcMgr, fileRec)

	if r == nil {
		t.Fatal("NewReconciler returned nil")
	}

	if r.config != cfg {
		t.Error("Config not set correctly")
	}

	if r.gitMgr != gitMgr {
		t.Error("GitManager not set correctly")
	}

	if r.pkgMgr != pkgMgr {
		t.Error("PackageManager not set correctly")
	}

	if r.svcMgr != svcMgr {
		t.Error("ServiceManager not set correctly")
	}

	if r.fileRec != fileRec {
		t.Error("FileReconciler not set correctly")
	}
}

func TestSyncEdgeCDRepo_CloneOnFirstRun(t *testing.T) {
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "edge-cd")

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			EdgeCD: userconfig.EdgeCDSection{
				Repo: userconfig.RepoConfig{
					URL:             "https://github.com/test/edge-cd.git",
					Branch:          "main",
					DestinationPath: destPath,
				},
			},
		},
		EdgeCDRepoPath: destPath,
	}

	gitMgr := &git.MockRepoManager{
		CloneRepoFunc: func(url, branch, destPath string, sparseCheckoutPaths []string) error {
			// Verify correct parameters
			if url != "https://github.com/test/edge-cd.git" {
				t.Errorf("CloneRepo url = %v, want https://github.com/test/edge-cd.git", url)
			}
			if branch != "main" {
				t.Errorf("CloneRepo branch = %v, want main", branch)
			}
			if len(sparseCheckoutPaths) != 1 || sparseCheckoutPaths[0] != "cmd/edge-cd" {
				t.Errorf("CloneRepo sparseCheckoutPaths = %v, want [cmd/edge-cd]", sparseCheckoutPaths)
			}
			return nil
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, nil, nil)
	r.syncEdgeCDRepo()

	// Verify CloneRepo was called
	if gitMgr.CloneRepoFunc == nil {
		t.Error("CloneRepo was not called")
	}
}

func TestSyncEdgeCDRepo_SyncOnSubsequentRun(t *testing.T) {
	tempDir := t.TempDir()
	destPath := filepath.Join(tempDir, "edge-cd")

	// Create directory to simulate existing repo
	os.MkdirAll(destPath, 0755)

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			EdgeCD: userconfig.EdgeCDSection{
				Repo: userconfig.RepoConfig{
					URL:             "https://github.com/test/edge-cd.git",
					Branch:          "main",
					DestinationPath: destPath,
				},
			},
		},
		EdgeCDRepoPath: destPath,
	}

	syncCalled := false
	gitMgr := &git.MockRepoManager{
		SyncRepoFunc: func(repoPath, branch string, sparseCheckoutPaths []string) error {
			syncCalled = true
			if repoPath != destPath {
				t.Errorf("SyncRepo repoPath = %v, want %v", repoPath, destPath)
			}
			return nil
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, nil, nil)
	r.syncEdgeCDRepo()

	if !syncCalled {
		t.Error("SyncRepo was not called")
	}
}

func TestSyncConfigRepo_SkipsFileURL(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			Config: userconfig.ConfigSection{
				Path: "devices/test",
				Repo: userconfig.ConfigRepo{
					URL:      "file:///opt/config",
					Branch:   "main",
					DestPath: "/opt/config",
				},
			},
		},
		ConfigRepoPath: "/opt/config",
	}

	cloneCalled := false
	gitMgr := &git.MockRepoManager{
		CloneRepoFunc: func(url, branch, destPath string, sparseCheckoutPaths []string) error {
			cloneCalled = true
			return nil
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, nil, nil)
	r.syncConfigRepo()

	// Should NOT call CloneRepo for file:// URLs
	if cloneCalled {
		t.Error("CloneRepo was called for file:// URL (should be skipped)")
	}
}

func TestIsConfigChanged_DetectsChange(t *testing.T) {
	tempDir := t.TempDir()
	commitPath := filepath.Join(tempDir, "last-commit.txt")

	// Write old commit
	os.WriteFile(commitPath, []byte("abc123"), 0644)

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			Config: userconfig.ConfigSection{
				Repo: userconfig.ConfigRepo{
					URL: "https://github.com/test/config.git",
				},
			},
		},
		ConfigRepoPath:   "/opt/config",
		ConfigCommitPath: commitPath,
	}

	gitMgr := &git.MockRepoManager{
		GetCurrentCommitFunc: func(repoPath string) (string, error) {
			return "def456", nil // Different commit
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, nil, nil)
	changed := r.isConfigChanged()

	if !changed {
		t.Error("isConfigChanged() = false, want true (commit changed)")
	}
}

func TestIsConfigChanged_NoChange(t *testing.T) {
	tempDir := t.TempDir()
	commitPath := filepath.Join(tempDir, "last-commit.txt")

	// Write current commit
	os.WriteFile(commitPath, []byte("abc123"), 0644)

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			Config: userconfig.ConfigSection{
				Repo: userconfig.ConfigRepo{
					URL: "https://github.com/test/config.git",
				},
			},
		},
		ConfigRepoPath:   "/opt/config",
		ConfigCommitPath: commitPath,
	}

	gitMgr := &git.MockRepoManager{
		GetCurrentCommitFunc: func(repoPath string) (string, error) {
			return "abc123", nil // Same commit
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, nil, nil)
	changed := r.isConfigChanged()

	if changed {
		t.Error("isConfigChanged() = true, want false (commit unchanged)")
	}
}

func TestIsConfigChanged_SkipsFileURL(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			Config: userconfig.ConfigSection{
				Repo: userconfig.ConfigRepo{
					URL: "file:///opt/config",
				},
			},
		},
	}

	r := NewReconciler(cfg, nil, nil, nil, nil)
	changed := r.isConfigChanged()

	if changed {
		t.Error("isConfigChanged() = true for file:// URL (should always return false)")
	}
}

func TestReconcilePackages(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PackageManager: userconfig.PackageManagerSection{
				RequiredPackages: []string{"git", "curl", "htop"},
			},
		},
	}

	installCalled := false
	var installedPkgs []string

	pkgMgr := &pkgmgr.MockPackageManager{
		InstallFunc: func(packages []string) error {
			installCalled = true
			installedPkgs = packages
			return nil
		},
	}

	r := NewReconciler(cfg, nil, pkgMgr, nil, nil)
	r.reconcilePackages()

	if !installCalled {
		t.Error("Install was not called")
	}

	if len(installedPkgs) != 3 {
		t.Errorf("Installed %d packages, want 3", len(installedPkgs))
	}
}

func TestReconcileAutoUpgrade_WhenEnabled(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PackageManager: userconfig.PackageManagerSection{
				AutoUpgrade:      true,
				RequiredPackages: []string{"git", "curl"},
			},
		},
	}

	upgradeCalled := false
	pkgMgr := &pkgmgr.MockPackageManager{
		UpgradeFunc: func(packages []string) error {
			upgradeCalled = true
			return nil
		},
	}

	r := NewReconciler(cfg, nil, pkgMgr, nil, nil)
	r.reconcileAutoUpgrade()

	if !upgradeCalled {
		t.Error("Upgrade was not called when autoUpgrade=true")
	}
}

func TestReconcileAutoUpgrade_WhenDisabled(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PackageManager: userconfig.PackageManagerSection{
				AutoUpgrade:      false,
				RequiredPackages: []string{"git", "curl"},
			},
		},
	}

	upgradeCalled := false
	pkgMgr := &pkgmgr.MockPackageManager{
		UpgradeFunc: func(packages []string) error {
			upgradeCalled = true
			return nil
		},
	}

	r := NewReconciler(cfg, nil, pkgMgr, nil, nil)
	r.reconcileAutoUpgrade()

	if upgradeCalled {
		t.Error("Upgrade was called when autoUpgrade=false")
	}
}

func TestReconcileEdgeCD_MarksServiceForRestart(t *testing.T) {
	tempDir := t.TempDir()
	commitPath := filepath.Join(tempDir, "edge-cd-commit.txt")

	// Write old commit
	os.WriteFile(commitPath, []byte("old123"), 0644)

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			EdgeCD: userconfig.EdgeCDSection{
				Repo: userconfig.RepoConfig{},
			},
		},
		EdgeCDRepoPath:   "/opt/edge-cd",
		EdgeCDCommitPath: commitPath,
	}

	gitMgr := &git.MockRepoManager{
		GetCurrentCommitFunc: func(repoPath string) (string, error) {
			return "new456", nil
		},
		GetCommitDiffFunc: func(repoPath, oldCommit, newCommit string) ([]string, error) {
			// Script changed
			return []string{"cmd/edge-cd/edge-cd", "README.md"}, nil
		},
	}

	enableCalled := false
	svcMgr := &svcmgr.MockServiceManager{
		EnableFunc: func(serviceName string) error {
			enableCalled = true
			if serviceName != "edge-cd" {
				t.Errorf("Enable called with %v, want edge-cd", serviceName)
			}
			return nil
		},
	}

	r := NewReconciler(cfg, gitMgr, nil, svcMgr, nil)
	state := &runtime.RuntimeState{
		ServicesToRestart: make(map[string]bool),
	}

	r.reconcileEdgeCD(state)

	// Verify service marked for restart
	services := state.GetServicesToRestart()
	if len(services) != 1 || services[0] != "edge-cd" {
		t.Errorf("Services to restart = %v, want [edge-cd]", services)
	}

	// Verify Enable was called
	if !enableCalled {
		t.Error("Enable was not called")
	}
}

func TestReconcileFiles(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			Config: userconfig.ConfigSection{
				Path: "devices/test",
			},
			Files: []userconfig.FileSpec{
				{Type: "content", DestPath: "/etc/test", Content: "test"},
			},
		},
		ConfigRepoPath: "/opt/config",
	}

	fileRecCalled := false
	fileRec := &files.MockFileReconciler{
		ReconcileFilesFunc: func(configRepoPath, configPath string, fileSpecs []userconfig.FileSpec) (*files.ReconcileResult, error) {
			fileRecCalled = true
			return &files.ReconcileResult{
				ServicesToRestart: []string{"nginx", "redis"},
				RequiresReboot:    true,
			}, nil
		},
	}

	r := NewReconciler(cfg, nil, nil, nil, fileRec)
	state := &runtime.RuntimeState{
		ServicesToRestart: make(map[string]bool),
	}

	r.reconcileFiles(state)

	if !fileRecCalled {
		t.Error("FileReconciler.ReconcileFiles was not called")
	}

	// Verify services collected
	services := state.GetServicesToRestart()
	if len(services) != 2 {
		t.Errorf("Got %d services, want 2", len(services))
	}

	// Verify reboot flag
	if !state.RequireReboot {
		t.Error("RequireReboot not set")
	}
}

func TestRestartServices(t *testing.T) {
	cfg := &config.Config{}

	restartCalls := []string{}
	svcMgr := &svcmgr.MockServiceManager{
		RestartFunc: func(serviceName string) error {
			restartCalls = append(restartCalls, serviceName)
			return nil
		},
	}

	r := NewReconciler(cfg, nil, nil, svcMgr, nil)

	state := &runtime.RuntimeState{
		ServicesToRestart: map[string]bool{
			"nginx":  true,
			"redis":  true,
			"edge-cd": true,
		},
	}

	r.restartServices(state)

	if len(restartCalls) != 3 {
		t.Errorf("Restart called %d times, want 3", len(restartCalls))
	}
}

func TestSleep_RespectsInterval(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PollingInterval: 1, // 1 second
		},
	}

	r := NewReconciler(cfg, nil, nil, nil, nil)

	ctx := context.Background()
	start := time.Now()
	r.sleep(ctx)
	elapsed := time.Since(start)

	// Should sleep for approximately 1 second
	if elapsed < 900*time.Millisecond || elapsed > 1200*time.Millisecond {
		t.Errorf("Sleep duration = %v, want ~1s", elapsed)
	}
}

func TestSleep_RespectsContextCancellation(t *testing.T) {
	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PollingInterval: 60, // 60 seconds
		},
	}

	r := NewReconciler(cfg, nil, nil, nil, nil)

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	r.sleep(ctx)
	elapsed := time.Since(start)

	// Should return immediately after context cancel (~100ms)
	if elapsed > 500*time.Millisecond {
		t.Errorf("Sleep duration = %v, should return quickly after context cancel", elapsed)
	}
}

func TestRun_ExitsOnContextCancel(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		Spec: &userconfig.Spec{
			PollingInterval: 1,
			EdgeCD: userconfig.EdgeCDSection{
				Repo: userconfig.RepoConfig{},
			},
			Config: userconfig.ConfigSection{
				Repo: userconfig.ConfigRepo{
					URL: "file:///opt/config",
				},
			},
			PackageManager: userconfig.PackageManagerSection{},
		},
		EdgeCDRepoPath:   tempDir,
		EdgeCDCommitPath: filepath.Join(tempDir, "edge-cd-commit.txt"),
		ConfigRepoPath:   tempDir,
		ConfigCommitPath: filepath.Join(tempDir, "config-commit.txt"),
	}

	gitMgr := &git.MockRepoManager{
		GetCurrentCommitFunc: func(repoPath string) (string, error) {
			return "abc123", nil
		},
	}
	pkgMgr := &pkgmgr.MockPackageManager{}
	svcMgr := &svcmgr.MockServiceManager{
		EnableFunc: func(serviceName string) error {
			return nil
		},
	}
	fileRec := &files.MockFileReconciler{}

	r := NewReconciler(cfg, gitMgr, pkgMgr, svcMgr, fileRec)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// Should exit after 500ms
	start := time.Now()
	r.Run(ctx)
	elapsed := time.Since(start)

	if elapsed > 1*time.Second {
		t.Errorf("Run duration = %v, should exit quickly after context timeout", elapsed)
	}
}
