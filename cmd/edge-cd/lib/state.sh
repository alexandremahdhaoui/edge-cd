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
# State
# ------------------------------------------------------------------#

__DEFAULT_STATE_BASE_PATH="${HOME}/.local/state/bootstrap-cluster/${CLUSTER_NAME}"
STATE_BASE_PATH="${STATE_BASE_PATH:-${__DEFAULT_STATE_BASE_PATH}}"

function execFunc() {
	local funcName="${1}"
	local stateFile="${STATE_BASE_PATH}/${SCRIPT_NAME}.state"

	# -- check if function is done
	if grep -q "^${funcName}$" "${stateFile}"; then
		logInfo "Skipping '${funcName}': already executed"
		return 0
	fi

	logInfo "Executing '${funcName}'"
	"${funcName}" "${@:2}"
	logInfo "Successfully executed '${funcName}'"

	# -- mark function as done
	mkdir -p "${STATE_BASE_PATH}"
	echo "${funcName}" | tee 1>/dev/null "${stateFile}"
}
