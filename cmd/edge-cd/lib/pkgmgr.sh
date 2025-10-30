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

__LOADED_LIB_PKGMGR=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
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

__get_package_manager_config() {
	pkgManager="$(read_config '.packageManager.name')"

	if [ "${pkgManager}" = "CUSTOM" ]; then
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

reconcile_package_auto_upgrade() {
	autoUpgrade="$(yq -e '.packageManager.autoUpgrade' "$(get_config_spec_abspath)" 2>/dev/null || echo "false")"
	if [ "${autoUpgrade}" != "true" ]; then
		return 0
	fi

	logInfo "Upgrading packages"

	config="$(__get_package_manager_config)"

	# Build update command from YAML array
	update_cmd=$(echo "${config}" | yq -e -r '.update[]' | tr '\n' ' ')
	eval "${update_cmd}"

	# Get packages as space-separated list
	packages="$(yq -e '.packageManager.requiredPackages[]' "$(get_config_spec_abspath)" 2>/dev/null | tr '\n' ' ' || echo "")"

	# Build upgrade command from YAML array
	upgrade_cmd=$(echo "${config}" | yq -e -r '.upgrade[]' | tr '\n' ' ')

	eval "${upgrade_cmd} ${packages}"
}

reconcile_packages() {
	packages="$(yq -e '.packageManager.requiredPackages[]' "$(get_config_spec_abspath)" 2>/dev/null | tr '\n' ' ' || echo "")"
	if [ -z "${packages}" ] || [ "${packages}" = " " ]; then
		logInfo "No package to install"
		return
	fi

	logInfo "Installing packages '${packages}'"

	config="$(__get_package_manager_config)"

	# Build install command from YAML array
	install_cmd=$(echo "${config}" | yq -e -r '.install[]' | tr '\n' ' ')

	# Build update command from YAML array
	update_cmd=$(echo "${config}" | yq -e -r '.update[]' | tr '\n' ' ')

	eval "${update_cmd}"
	eval "${install_cmd} ${packages}"
}

