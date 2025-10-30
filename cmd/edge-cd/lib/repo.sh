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

__LOADED_LIB_REPO=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
LIB_DIR="${SRC_DIR}/lib"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"

# ------------------------------------------------------------------#
# Helper Functions
# ------------------------------------------------------------------#

is_file_url() {
	url="$1"
	case "${url}" in
		file://*)
			return 0
			;;
		*)
			return 1
			;;
	esac
}

# ------------------------------------------------------------------#
# EdgeCD Repository Operations
# ------------------------------------------------------------------#

clone_repo_edge_cd() {
	logInfo "Cloning EdgeCD repo"
	git clone --filter=blob:none --no-checkout "${EDGE_CD_REPO_URL}" "${EDGE_CD_REPO_DESTINATION_PATH}"

	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" sparse-checkout init
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" sparse-checkout set "cmd/edge-cd"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" checkout "${EDGE_CD_REPO_BRANCH}"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" fetch origin "${EDGE_CD_REPO_BRANCH}"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" pull
}

sync_repo_edge_cd() {
	logInfo "Pulling EdgeCD repo"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" sparse-checkout set "cmd/edge-cd"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" fetch origin "${EDGE_CD_REPO_BRANCH}"
	git -C "${EDGE_CD_REPO_DESTINATION_PATH}" reset --hard FETCH_HEAD
}

# ------------------------------------------------------------------#
# Config Repository Operations
# ------------------------------------------------------------------#

clone_repo_config() {
	if is_file_url "${CONFIG_REPO_URL}"; then
		logInfo "Using local file-based repository for config. This is a non-production option. Skipping git clone."
		return 0
	fi

	logInfo "Cloning config repo"
	git clone --filter=blob:none --no-checkout "${CONFIG_REPO_URL}" "${CONFIG_REPO_DEST_PATH}"

	git -C "${CONFIG_REPO_DEST_PATH}" sparse-checkout init
	# Set sparse-checkout to include the config path directory
	git -C "${CONFIG_REPO_DEST_PATH}" sparse-checkout set "${CONFIG_PATH}"
	git -C "${CONFIG_REPO_DEST_PATH}" checkout "${CONFIG_REPO_BRANCH}"
	git -C "${CONFIG_REPO_DEST_PATH}" fetch origin "${CONFIG_REPO_BRANCH}"
	git -C "${CONFIG_REPO_DEST_PATH}" pull
}

sync_repo_config() {
	if is_file_url "${CONFIG_REPO_URL}"; then
		logInfo "Using local file-based repository for config. This is a non-production configuration option. Skipping git pull."
		return 0
	fi

	logInfo "Pulling config repo"
	# Set sparse-checkout to include the config path directory
	git -C "${CONFIG_REPO_DEST_PATH}" sparse-checkout set "${CONFIG_PATH}"
	git -C "${CONFIG_REPO_DEST_PATH}" fetch origin "${CONFIG_REPO_BRANCH}"
	git -C "${CONFIG_REPO_DEST_PATH}" reset --hard FETCH_HEAD
}
