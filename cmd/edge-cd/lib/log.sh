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

declare -g __LOADED_LIB_LOG=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")}"

# ------------------------------------------------------------------#
# Logger
# ------------------------------------------------------------------#

LOG_FORMAT="${LOG_FORMAT-console}"

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
