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

declare -g __LOADED_LIB_PKGMGR=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")}"
LIB_DIR="${SRC_DIR}/lib"
PKGMGR_DIR="${SRC_DIR}/package-managers"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"
[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"

# ------------------------------------------------------------------#
# Preambule
# ------------------------------------------------------------------#

function __get_package_manager_config() {
	local pkgManager configPath
	pkgManager="$(read_config '.packageManager.name')"

	if [[ "${pkgManager}" == "CUSTOM" ]]; then
		logErr "Not implemented yet"
		exit 1
	fi

	configPath="${PKGMGR_DIR}/${pkgManager}.yaml"
	if [ ! -f "${configPath}" ]; then
		logErr "Package manager \"${pkgManager}\" cannot be found"
		exit 1
	fi

	yq -e -r '.' "${configPath}"
}

function reconcile_package_auto_upgrade() {
	local autoUpgrade
	autoUpgrade="$(yq -e '.packageManager.autoUpgrade' "$(get_config_spec_abspath)" 2>/dev/null || echo "false")"
	if [[ "${autoUpgrade}" != "true" ]]; then
		return 0
	fi

	logInfo "Upgrading packages"

	local config
	config="$(__get_package_manager_config)"

	local -a update
	local -a upgrade
	readarray -t update < <(echo "${config}" | yq -e -r '.update[]')
	readarray -t upgrade < <(echo "${config}" | yq -e -r '.upgrade[]')
	"${update[@]}"

	local packages
	packages="$(yq -e '.packageManager.requiredPackages[]' "$(get_config_spec_abspath)" || echo "")"
	local -a packageArray
	readarray -t packageArray <<<"${packages}"

	"${upgrade[@]}" "${packageArray[@]}"
}

function reconcile_packages() {
	local packages
	packages="$(yq -e '.packageManager.requiredPackages[]' "$(get_config_spec_abspath)" 2>/dev/null || echo "")"
	[ "${packages}" == "" ] \
		&& logInfo "No package to install" \
		&& return

	logInfo "Installing packages '${packages}'"

	local -a packageArray
	readarray -t packageArray <<<"${packages}"

	local config
	config="$(__get_package_manager_config)"

	local -a install
	local -a update
	readarray -t install < <(echo "${config}" | yq -e -r '.install[]')
	readarray -t update < <(echo "${config}" | yq -e -r '.update[]')

	"${update[@]}"
	"${install[@]}" "${packageArray[@]}"
}

