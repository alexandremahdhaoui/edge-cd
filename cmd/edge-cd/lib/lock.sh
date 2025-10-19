#!/usr/bin/env bash

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

	echo "${$}" "${lockFilePath}"
}

function unlock() {
	logInfo "Unlocking '${SCRIPT_NAME}'"
	rm -f "$(__get_lock_file_path)"
}
