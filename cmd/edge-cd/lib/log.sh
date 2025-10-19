#!/usr/bin/env bash

# ------------------------------------------------------------------#
# Logger
# ------------------------------------------------------------------#

LOG_FORMAT="${LOG_FORMAT-console}"

# must be located at "${HOME}/.local/lib/bootstrap-cluster/lock.sh"

function __log() {
	local xtrace=0
	if [[ "$-" == *x* ]]; then
		xtrace=1
		set +o xtrace
	fi

	local logLevel="${1}"
	if [[ "${LOG_FORMAT}" == "json" ]]; then
		local logMessage
		logMessage="$(printf "%q" "${2}")"

		cat <<-EOF 1>&2
			{"level":"${logLevel,,}","message":"${logMessage}"}
		EOF
	else
		echo "[${logLevel^^}]" "${@:2}" 1>&2
	fi

	if [[ "${xtrace}" == "1" ]]; then
		set -o xtrace
	fi
}

function logInfo() {
	__log INFO "${*}"
}

function logErr() {
	__log ERROR "${*}"
}
