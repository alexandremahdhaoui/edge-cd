#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

__on_failure() {
  echo >&2 "[ERROR] Test failed"
}

__on_success() {
  rm /tmp/edge-cd.log
}

main() {
  # -- Determine which implementation to test (default: shell)
  EDGE_CD_IMPL="${EDGE_CD_IMPL:-shell}"

  # -- Run edge-cd in the background and redirect stderr to a log file
  touch /tmp/edge-cd.log
  trap __on_failure EXIT
  # In the Docker container, the config is directly in /opt/config
  # CONFIG_PATH is the relative directory path within the config repo (in this case, root)
  # CONFIG_REPO_DEST_PATH is /opt/config
  # CONFIG_SPEC_FILE is config.yaml
  # The full path to the config will be: /opt/config/./config.yaml

  if [ "${EDGE_CD_IMPL}" = "go" ]; then
    echo "[INFO] Testing edge-cd-go (Go implementation)"
    # Override edge-cd repo path for Go implementation to find config files at actual location
    CONFIG_PATH=. CONFIG_REPO_DEST_PATH=/opt/config CONFIG_SPEC_FILE=config.yaml EDGE_CD_REPO_DESTINATION_PATH=/opt/src/edge-cd /opt/src/edge-cd/bin/edge-cd-go 2>&1 | tee /tmp/edge-cd.log &
    EDGE_CD_PROCESS_PATTERN="edge-cd-go"
  else
    echo "[INFO] Testing edge-cd (shell implementation)"
    CONFIG_PATH=. CONFIG_REPO_DEST_PATH=/opt/config CONFIG_SPEC_FILE=config.yaml /opt/src/edge-cd/cmd/edge-cd/edge-cd 2>&1 | tee /tmp/edge-cd.log &
    EDGE_CD_PROCESS_PATTERN="^bash /opt/src/edge-cd/cmd/edge-cd/edge-cd$"
  fi

  # -- Wait for edge-cd to complete its first reconciliation loop
  # For shell: look for "Sleeping for"
  # For Go: look for "Sleeping" (from slog output)
  if ! timeout 60 bash -c "until ! pgrep -f '${EDGE_CD_PROCESS_PATTERN}' &>/dev/null || grep -q 'Sleeping' /tmp/edge-cd.log; do sleep 1; done"; then
    echo "[ERROR] Timeout waiting for 'Sleeping' in edge-cd.log."
    exit 1
  fi

  if ! pgrep -f "${EDGE_CD_PROCESS_PATTERN}" &>/dev/null; then
    echo "[ERROR] edge-cd is not running"
    exit 1
  fi

  # -- Verify package installation
  INSTALLED_PACKAGES=$(opkg list-installed)
  echo "${INSTALLED_PACKAGES}" | grep -q htop

  # -- Verify file creation from content
  grep -q "Hello from content" /etc/hello-world-content

  # -- Verify file synchronization
  grep -q "Hello from file" /etc/hello-from-file

  # -- Verify service is enabled and started
  [ -L /etc/rc.d/S95hello-world ]

  # -- Verify service is started (check for running file)
  [ -f /tmp/hello-world-running ]

  echo "[SUCCESS] All tests passed!"
  trap __on_success EXIT
}

main "$@"
