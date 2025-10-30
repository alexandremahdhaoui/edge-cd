package reconcile

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/config"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/files"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/git"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/pkgmgr"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/runtime"
	"github.com/alexandremahdhaoui/edge-cd/pkg/edgecd/svcmgr"
)

// Reconciler orchestrates the edge-cd reconciliation loop.
// It coordinates all operations: syncing repos, reconciling packages/files/services.
type Reconciler struct {
	config  *config.Config
	gitMgr  git.RepoManager
	pkgMgr  pkgmgr.PackageManager
	svcMgr  svcmgr.ServiceManager
	fileRec files.FileReconciler
}

// NewReconciler creates a new Reconciler with injected dependencies.
func NewReconciler(
	cfg *config.Config,
	gitMgr git.RepoManager,
	pkgMgr pkgmgr.PackageManager,
	svcMgr svcmgr.ServiceManager,
	fileRec files.FileReconciler,
) *Reconciler {
	return &Reconciler{
		config:  cfg,
		gitMgr:  gitMgr,
		pkgMgr:  pkgMgr,
		svcMgr:  svcMgr,
		fileRec: fileRec,
	}
}

// Run executes the reconciliation loop forever until context is cancelled.
func (r *Reconciler) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			slog.Info("Shutting down gracefully")
			return
		default:
			r.reconcile(ctx)
			r.sleep(ctx)
		}
	}
}

// reconcile performs a single reconciliation iteration.
func (r *Reconciler) reconcile(ctx context.Context) {
	state := runtime.NewRuntimeState()

	// 1. Sync edge-cd repo
	r.syncEdgeCDRepo()

	// 2. Sync config repo
	r.syncConfigRepo()

	// 3. Check if config changed
	configChanged := r.isConfigChanged()

	// 4. Reconcile packages (if changed)
	if configChanged {
		r.reconcilePackages()
	}

	// 5. Reconcile auto-upgrade
	r.reconcileAutoUpgrade()

	// 6. Reconcile edge-cd
	r.reconcileEdgeCD(state)

	// 7. Reconcile files
	r.reconcileFiles(state)

	// 8. Handle reboot
	if state.RequireReboot {
		r.reboot()
		return
	}

	// 9. Restart services
	r.restartServices(state)

	// 10. Commit changes
	r.commitLastChange()
}

// syncEdgeCDRepo clones or syncs the edge-cd repository.
func (r *Reconciler) syncEdgeCDRepo() {
	url := r.config.Spec.EdgeCD.Repo.URL
	branch := r.config.Spec.EdgeCD.Repo.Branch
	destPath := r.config.EdgeCDRepoPath

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		if err := r.gitMgr.CloneRepo(url, branch, destPath, []string{"cmd/edge-cd"}); err != nil {
			slog.Error("Failed to clone edge-cd repo", "error", err)
		}
	} else {
		if err := r.gitMgr.SyncRepo(destPath, branch, []string{"cmd/edge-cd"}); err != nil {
			slog.Error("Failed to sync edge-cd repo", "error", err)
		}
	}
}

// syncConfigRepo clones or syncs the configuration repository.
func (r *Reconciler) syncConfigRepo() {
	url := r.config.Spec.Config.Repo.URL
	branch := r.config.Spec.Config.Repo.Branch
	destPath := r.config.ConfigRepoPath
	configPath := r.config.Spec.Config.Path

	// Skip git operations for file:// URLs
	if strings.HasPrefix(url, "file://") {
		slog.Info("Using local file-based repository for config, skipping git clone")
		return
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		if err := r.gitMgr.CloneRepo(url, branch, destPath, []string{configPath}); err != nil {
			slog.Error("Failed to clone config repo", "error", err)
		}
	} else {
		if err := r.gitMgr.SyncRepo(destPath, branch, []string{configPath}); err != nil {
			slog.Error("Failed to sync config repo", "error", err)
		}
	}
}

// isConfigChanged checks if the config repository commit has changed.
func (r *Reconciler) isConfigChanged() bool {
	// Handle file:// URLs (skip commit tracking)
	if strings.HasPrefix(r.config.Spec.Config.Repo.URL, "file://") {
		slog.Info("Using local file-based repository, skipping commit synchronization")
		return false
	}

	// Read last commit from file
	lastCommitData, _ := os.ReadFile(r.config.ConfigCommitPath)
	lastCommit := strings.TrimSpace(string(lastCommitData))

	// Get current commit
	currentCommit, err := r.gitMgr.GetCurrentCommit(r.config.ConfigRepoPath)
	if err != nil {
		slog.Error("Failed to get current commit", "error", err)
		return false
	}

	// Compare
	if lastCommit == currentCommit {
		slog.Info("Config already in sync", "commit", currentCommit)
		return false
	}

	slog.Info("Starting configuration synchronization", "commit", currentCommit)
	return true
}

// reconcilePackages installs required packages.
func (r *Reconciler) reconcilePackages() {
	packages := r.config.Spec.PackageManager.RequiredPackages
	if len(packages) == 0 {
		return
	}

	slog.Info("Reconciling packages")
	if err := r.pkgMgr.Install(packages); err != nil {
		slog.Error("Failed to install packages", "error", err)
	}
}

// reconcileAutoUpgrade upgrades packages if auto-upgrade is enabled.
func (r *Reconciler) reconcileAutoUpgrade() {
	if !r.config.Spec.PackageManager.AutoUpgrade {
		return
	}

	packages := r.config.Spec.PackageManager.RequiredPackages
	if len(packages) == 0 {
		return
	}

	slog.Info("Auto-upgrading packages")
	if err := r.pkgMgr.Upgrade(packages); err != nil {
		slog.Error("Failed to upgrade packages", "error", err)
	}
}

// reconcileEdgeCD checks if edge-cd script has changed and marks service for restart.
func (r *Reconciler) reconcileEdgeCD(state *runtime.RuntimeState) {
	slog.Info("Reconciling EdgeCD")

	// Get last and current commits
	lastCommitData, _ := os.ReadFile(r.config.EdgeCDCommitPath)
	lastCommit := strings.TrimSpace(string(lastCommitData))

	currentCommit, err := r.gitMgr.GetCurrentCommit(r.config.EdgeCDRepoPath)
	if err != nil {
		slog.Error("Failed to get current commit", "error", err)
		return
	}

	// Check if edge-cd script changed between commits
	if lastCommit != "" && lastCommit != currentCommit {
		changedFiles, err := r.gitMgr.GetCommitDiff(r.config.EdgeCDRepoPath, lastCommit, currentCommit)
		if err != nil {
			slog.Error("Failed to get commit diff", "error", err)
		} else {
			for _, file := range changedFiles {
				if file == "cmd/edge-cd/edge-cd" || file == "cmd/edge-cd-go/main.go" {
					slog.Info("EdgeCD script has changed, marking service for restart")
					state.AddServiceRestart("edge-cd")
					break
				}
			}
		}
	}

	// Ensure edge-cd service is always enabled
	if err := r.svcMgr.Enable("edge-cd"); err != nil {
		slog.Error("Failed to enable edge-cd service", "error", err)
	}

	// Write current commit
	os.MkdirAll(filepath.Dir(r.config.EdgeCDCommitPath), 0755)
	os.WriteFile(r.config.EdgeCDCommitPath, []byte(currentCommit), 0644)
}

// reconcileFiles reconciles all files defined in the configuration.
func (r *Reconciler) reconcileFiles(state *runtime.RuntimeState) {
	if len(r.config.Spec.Files) == 0 {
		return
	}

	slog.Info("Reconciling files")

	result, err := r.fileRec.ReconcileFiles(
		r.config.ConfigRepoPath,
		r.config.Spec.Config.Path,
		r.config.Spec.Files,
	)

	if err != nil {
		slog.Error("Failed to reconcile files", "error", err)
		return
	}

	// Add services to restart
	for _, svc := range result.ServicesToRestart {
		state.AddServiceRestart(svc)
	}

	// Set reboot flag
	if result.RequiresReboot {
		state.RequireReboot = true
	}
}

// reboot reboots the system (placeholder implementation).
func (r *Reconciler) reboot() {
	slog.Info("Rebooting now")
	// In a real implementation, this would execute: exec.Command("reboot").Run()
	// For testing, we just log
	fmt.Println("REBOOT TRIGGERED") // Special marker for tests
}

// restartServices restarts all services that were marked for restart.
// Services are enabled before restarting to ensure they start on boot.
func (r *Reconciler) restartServices(state *runtime.RuntimeState) {
	services := state.GetServicesToRestart()
	if len(services) == 0 {
		return
	}

	slog.Info("Restarting services", "services", services)

	for _, svc := range services {
		// Enable service first to ensure it starts on boot
		if err := r.svcMgr.Enable(svc); err != nil {
			slog.Error("Failed to enable service", "service", svc, "error", err)
		}

		// Then restart the service
		if err := r.svcMgr.Restart(svc); err != nil {
			slog.Error("Failed to restart service", "service", svc, "error", err)
		}
	}
}

// commitLastChange writes the current config commit to file.
func (r *Reconciler) commitLastChange() {
	// Skip for file:// URLs
	if strings.HasPrefix(r.config.Spec.Config.Repo.URL, "file://") {
		return
	}

	currentCommit, err := r.gitMgr.GetCurrentCommit(r.config.ConfigRepoPath)
	if err != nil {
		slog.Error("Failed to get current commit", "error", err)
		return
	}

	os.MkdirAll(filepath.Dir(r.config.ConfigCommitPath), 0755)
	if err := os.WriteFile(r.config.ConfigCommitPath, []byte(currentCommit), 0644); err != nil {
		slog.Error("Failed to write commit file", "error", err)
		return
	}

	slog.Info("Synced commit successfully", "commit", currentCommit)
}

// sleep pauses for the configured polling interval or until context is cancelled.
func (r *Reconciler) sleep(ctx context.Context) {
	interval := r.config.Spec.PollingInterval
	if interval <= 0 {
		interval = 60 // default
	}

	slog.Info("Sleeping", "seconds", interval)

	timer := time.NewTimer(time.Duration(interval) * time.Second)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		return
	}
}
