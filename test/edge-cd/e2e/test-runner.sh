#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

main() {

  # -- Run edge-cd in the background and redirect stderr to a log file
  touch /tmp/edge-cd.log
  CONFIG_PATH=/opt/config/config.yaml /opt/src/edge-cd/cmd/edge-cd/edge-cd 2> /tmp/edge-cd.log &

  # -- Wait for edge-cd to complete its first reconciliation loop
  timeout 60 bash -c 'until grep -q "Sleeping for" /tmp/edge-cd.log; do sleep 1; done'

  # Check if the timeout occurred
  if [ $? -ne 0 ]; then
    echo "Error: Timeout waiting for 'Sleeping for' in edge-cd.log. Log content:"
    cat /tmp/edge-cd.log
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

  echo "All tests passed!"
}

main "$@"
