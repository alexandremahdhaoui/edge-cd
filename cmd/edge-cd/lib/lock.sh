#!/usr/bin/env bash
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

declare -g __LOADED_LIB_LOCK=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")}"

# ------------------------------------------------------------------#
# Lock
# ------------------------------------------------------------------#

__DEFAULT_LOCK_FILE_DIRNAME="/tmp/edge-cd"
LOCK_FILE_DIRNAME="${LOCK_FILE_DIRNAME:-${__DEFAULT_LOCK_FILE_DIRNAME}}"

function __get_lock_file_path() {
	echo "${LOCK_FILE_DIRNAME}/edge-cd.lock"
}

function lock() {
	local lockFilePath
	lockFilePath="$(__get_lock_file_path)"

	logInfo "Locking '${lockFilePath}'"
	if [ -f "${lockFilePath}" ]; then
		local lockPID
		lockPID="$(cat "${lockFilePath}")"

		if [ "${$}" == "$(cat "${lockFilePath}")" ]; then
			return 0 # already locked by this process
		fi

		if ps aux | awk '{print $2}' | grep -q "^${lockPID}$"; then
			logErr "Cannot take lock at '${lockFilePath}'"
			logErr "Another \"edge-cd\" process is already running"
			exit 1
		fi

		# previous process is not running, we can take over lock
	fi

	mkdir -p "$(dirname "${lockFilePath}")"
	echo "${$}" > "${lockFilePath}"
}

function unlock() {
	logInfo "Unlocking 'EdgeCD'"
	rm -f "$(__get_lock_file_path)"
}
