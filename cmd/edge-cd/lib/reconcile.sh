#!/bin/sh
#
# Copyright 2025 Alexandre Mahdhaoui
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# ------------------------------------------------------------------#
# Preambule
# ------------------------------------------------------------------#

__LOADED_LIB_RECONCILE=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
LIB_DIR="${SRC_DIR}/lib"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"
[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"
[ -z "${__LOADED_LIB_RUNTIME:-}" ] && . "${LIB_DIR}/runtime.sh"
[ -z "${__LOADED_LIB_REPO:-}" ] && . "${LIB_DIR}/repo.sh"
[ -z "${__LOADED_LIB_FILES:-}" ] && . "${LIB_DIR}/files.sh"
[ -z "${__LOADED_LIB_PKGMGR:-}" ] && . "${LIB_DIR}/pkgmgr.sh"
[ -z "${__LOADED_LIB_SVCMGR:-}" ] && . "${LIB_DIR}/svcmgr.sh"

# ------------------------------------------------------------------#
# Commit Tracking
# ------------------------------------------------------------------#

is_commit_in_sync() {
	if is_file_url "${CONFIG_REPO_URL}"; then
		logInfo "Using local file-based repository for config. Skipping commit synchronization."
		return 1 # Return non-zero status to signal skip
	fi

	lastCommit=$(cat "${CONFIG_COMMIT_PATH}" 2>/dev/null || echo "")
	currentCommit="$(git -C "${CONFIG_REPO_DEST_PATH}" rev-parse HEAD)"

	if [ "${lastCommit}" = "${currentCommit}" ]; then
		logInfo "Skipping synchronization: user config is already in sync with commit ${currentCommit}"
		return 1 # Return non-zero status to signal skip
	fi

	logInfo "Starting configuration synchronization for commit ${currentCommit}"
}

# ------------------------------------------------------------------#
# EdgeCD Self-Update
# ------------------------------------------------------------------#

reconcile_edge_cd() {
	logInfo "Reconciling EdgeCD"

	name=edge-cd

	# -- Get last synced commit and current commit
	lastCommit=$(cat "${EDGE_CD_COMMIT_PATH}" 2>/dev/null || echo "")
	currentCommit="$(git -C "${EDGE_CD_REPO_DESTINATION_PATH}" rev-parse HEAD)"

	# -- Check if edge-cd script has changed between commits
	if [ -n "${lastCommit}" ] && [ "${lastCommit}" != "${currentCommit}" ]; then
		if git -C "${EDGE_CD_REPO_DESTINATION_PATH}" diff --name-only "${lastCommit}" "${currentCommit}" | grep -q "^cmd/edge-cd/edge-cd$"; then
			logInfo "EdgeCD script has changed, marking service for restart"
			add_service_for_restart "${name}"
		fi
	fi

	# -- Ensure edge-cd service is always enabled (if service file exists)
	if [ -f "/etc/init.d/${name}" ]; then
		enable_service "${name}"
	else
		logInfo "Service file /etc/init.d/${name} does not exist, skipping service enable"
	fi

	# -- Commit this change
	mkdir -p "$(dirname "${EDGE_CD_COMMIT_PATH}")"
	echo "${currentCommit}" >"${EDGE_CD_COMMIT_PATH}"
}

# ------------------------------------------------------------------#
# Polling Backoff
# ------------------------------------------------------------------#

polling_backoff() {
	sleepTime="$(read_config '.pollingIntervalSecond' || echo "")"
	if [ -z "${sleepTime}" ]; then
		logWarn "Failed to read pollingIntervalSecond from config: Sleeping 60s" >&2
		sleepTime=60
	fi

	logInfo "Sleeping for ${sleepTime} seconds"
	sleep "${sleepTime}"
}

# ------------------------------------------------------------------#
# Main Reconciliation
# ------------------------------------------------------------------#

reconcile() {
	[ -d "${EDGE_CD_REPO_DESTINATION_PATH}" ] || clone_repo_edge_cd
	sync_repo_edge_cd

	[ -d "${CONFIG_REPO_DEST_PATH}" ] || clone_repo_config
	sync_repo_config

	# Check if the repository has been updated since the last
	# successful sync.
	if ! is_commit_in_sync; then
		# only re-install packages on new commits as drift detection
		# is not yet implemented
		reconcile_packages
	fi

	reconcile_package_auto_upgrade
	reconcile_edge_cd
	reconcile_files

	# -- Reboot if flag set
	if should_restart; then
		logInfo "Rebooting now"
		reboot
	fi

	# -- restart services
	services_list=$(get_services_to_restart)
	if [ -n "${services_list}" ]; then
		while IFS= read -r service || [ -n "$service" ]; do
			[ -z "${service}" ] && continue
			enable_service "${service}"
			restart_service "${service}"
		done <<EOF
${services_list}
EOF
	fi
}
