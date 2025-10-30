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

__LOADED_LIB_EDGECD=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
LIB_DIR="${SRC_DIR}/lib"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"
[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"

# ------------------------------------------------------------------#
# Default Configuration Values
# ------------------------------------------------------------------#

__DEFAULT_EDGE_CD_REPO_BRANCH=main
__DEFAULT_EDGE_CD_REPO_DESTINATION_PATH=/usr/local/src/edge-cd
__DEFAULT_EDGE_CD_REPO_URL=https://github.com/alexandremahdhaoui/edge-cd.git

__DEFAULT_EDGE_CD_COMMIT_PATH=/tmp/edge-cd/edge-cd-last-synchronized-commit.txt
__DEFAULT_CONFIG_COMMIT_PATH=/tmp/edge-cd/config-last-synchronized-commit.txt

__DEFAULT_CONFIG_REPO_BRANCH=main
__DEFAULT_CONFIG_REPO_DEST_PATH=/usr/local/src/edge-cd-config
__DEFAULT_CONFIG_SPEC_FILE=spec.yaml

# ------------------------------------------------------------------#
# EdgeCD Configuration Management
# ------------------------------------------------------------------#

declare_edgecd_config() {
	# Variables are implicitly declared in POSIX shell
	# This function serves as documentation of expected variables:
	# - CONFIG_PATH: Path within config repo to configuration directory
	# - CONFIG_SPEC_FILE: Name of the spec YAML file
	# - CONFIG_REPO_BRANCH: Git branch for config repository
	# - CONFIG_REPO_DEST_PATH: Local path where config repo is cloned
	# - CONFIG_REPO_URL: URL of config repository
	# - EDGE_CD_REPO_BRANCH: Git branch for EdgeCD repository
	# - EDGE_CD_REPO_DESTINATION_PATH: Local path where EdgeCD repo is cloned
	# - EDGE_CD_REPO_URL: URL of EdgeCD repository
	# - EDGE_CD_COMMIT_PATH: Path to file tracking last synced EdgeCD commit
	# - CONFIG_COMMIT_PATH: Path to file tracking last synced config commit
	:
}

init_edgecd_config() {
	# First, initialize the critical variables needed for get_config_spec_abspath()
	# These need to be set before set_extra_envs() is called
	CONFIG_SPEC_FILE="${CONFIG_SPEC_FILE:-${__DEFAULT_CONFIG_SPEC_FILE}}"
	CONFIG_REPO_DEST_PATH="${CONFIG_REPO_DEST_PATH:-${__DEFAULT_CONFIG_REPO_DEST_PATH}}"

	# CONFIG_PATH must be provided - no default
	if [ -z "${CONFIG_PATH}" ]; then
		logErr "CONFIG_PATH environment variable must be set"
		exit 1
	fi

	# Now we can safely call set_extra_envs since the path variables are initialized
	set_extra_envs

	# -- user config
	__BACKUP_CONFIG_PATH="$(read_value CONFIG_PATH '.config.path')"
	__BACKUP_CONFIG_SPEC_FILE="$(read_value CONFIG_SPEC_FILE '.config.spec' "${__DEFAULT_CONFIG_SPEC_FILE}")"
	__BACKUP_CONFIG_REPO_BRANCH="$(read_value CONFIG_REPO_BRANCH '.config.repo.branch' "${__DEFAULT_CONFIG_REPO_BRANCH}")"
	__BACKUP_CONFIG_REPO_DEST_PATH="$(read_value CONFIG_REPO_DEST_PATH '.config.repo.destPath' "${__DEFAULT_CONFIG_REPO_DEST_PATH}")"
	__BACKUP_CONFIG_REPO_URL="$(read_value CONFIG_REPO_URL '.config.repo.url')"

	# -- edge config
	__BACKUP_EDGE_CD_REPO_BRANCH="$(read_value EDGE_CD_REPO_BRANCH '.edgeCD.repo.branch' "${__DEFAULT_EDGE_CD_REPO_BRANCH}")"
	__BACKUP_EDGE_CD_REPO_DESTINATION_PATH="$(read_value EDGE_CD_REPO_DESTINATION_PATH '.edgeCD.repo.destinationPath' "${__DEFAULT_EDGE_CD_REPO_DESTINATION_PATH}")"
	__BACKUP_EDGE_CD_REPO_URL="$(read_value EDGE_CD_REPO_URL '.edgeCD.repo.url' "${__DEFAULT_EDGE_CD_REPO_URL}")"

	__BACKUP_EDGE_CD_COMMIT_PATH="$(read_value EDGE_CD_COMMIT_PATH '.edgeCD.commitPath' "${__DEFAULT_EDGE_CD_COMMIT_PATH}")"
	__BACKUP_CONFIG_COMMIT_PATH="$(read_value CONFIG_COMMIT_PATH '.config.commitPath' "${__DEFAULT_CONFIG_COMMIT_PATH}")"
}

reset_edgecd_config() {
	reset_extra_envs

	CONFIG_PATH="${__BACKUP_CONFIG_PATH}"
	CONFIG_SPEC_FILE="${__BACKUP_CONFIG_SPEC_FILE}"

	CONFIG_REPO_BRANCH="${__BACKUP_CONFIG_REPO_BRANCH}"
	CONFIG_REPO_DEST_PATH="${__BACKUP_CONFIG_REPO_DEST_PATH}"
	CONFIG_REPO_URL="${__BACKUP_CONFIG_REPO_URL}"

	EDGE_CD_REPO_BRANCH="${__BACKUP_EDGE_CD_REPO_BRANCH}"
	EDGE_CD_REPO_DESTINATION_PATH="${__BACKUP_EDGE_CD_REPO_DESTINATION_PATH}"
	EDGE_CD_REPO_URL="${__BACKUP_EDGE_CD_REPO_URL}"

	EDGE_CD_COMMIT_PATH="${__BACKUP_EDGE_CD_COMMIT_PATH}"
	CONFIG_COMMIT_PATH="${__BACKUP_CONFIG_COMMIT_PATH}"
}
