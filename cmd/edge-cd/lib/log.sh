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

__LOADED_LIB_LOG=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"

# ------------------------------------------------------------------#
# Logger
# ------------------------------------------------------------------#

LOG_FORMAT="${LOG_FORMAT-console}"

__log() {
	xtrace=0
	case "$-" in
		*x*) xtrace=1; set +x ;;
	esac

	logLevel="$1"
	shift

	if [ "${LOG_FORMAT}" = "json" ]; then
		# Convert to lowercase using tr
		logLevel_lower=$(printf '%s' "${logLevel}" | tr '[:upper:]' '[:lower:]' || echo "${logLevel}")
		# Escape the message for JSON (simplified - just escape quotes and backslashes)
		logMessage=$(printf '%s' "$*" | sed 's/\\/\\\\/g; s/"/\\"/g' || printf '%s' "$*")

		printf '{"level":"%s","message":"%s"}\n' "${logLevel_lower}" "${logMessage}" 1>&2
	else
		# Convert to uppercase using tr
		logLevel_upper=$(printf '%s' "${logLevel}" | tr '[:lower:]' '[:upper:]' || echo "${logLevel}")
		printf "[%s] %s\n" "${logLevel_upper}" "$*" 1>&2
	fi

	# Re-enable tracing if it was on before
	if [ "${xtrace}" = "1" ]; then
		set -x
	fi
	return 0
}

logInfo() {
	__log INFO "$@"
}

logErr() {
	__log ERROR "$@"
}
