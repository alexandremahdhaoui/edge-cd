#!/usr/bin/env bash

# shellcheck source=../lib/config.sh
# shellcheck source=../lib/log.sh

. "${LIB_DIR}/config.sh"
. "${LIB_DIR}/log.sh"

function __get_package_manager_config() {
	local pkgManager configPaht
	pkgManager="$(read_config '.packageManager.name')"

	if [[ "${pkgManager}" == "CUSTOM" ]]; then
		logErr "Not implemented yet"
		exit 1
	fi

	configPaht="${PKG_MGR_DIR}/${pkgManager}"
	if [ -f "${configPaht}" ]; then
		logErr "Package manager \"${pkgManager}\" cannot be found"
		exit 1
	fi

	read_yaml '.packageManager' "${configPaht}"
}

function auto_upgrade_packages() {
	local autoUpgrade
	autoUpgrade="$(yq -e '.packageManager.autoUpgrade' "${CONFIG_PATH}" || echo "false")"
	if [[ "${autoUpgrade}" != "true" ]]; then
		return 0
	fi

	logInfo "Upgrading packages"
	local config
	config="$(__get_package_manager_config)"
	local update upgrade
	update="$(read_yaml_stdin '.update' "${config}")"
	upgrade="$(read_yaml_stdin '.upgrade' "${config}")"

	${update}
	${upgrade}
}

function sync_packages() {
	local packages
	packages="$(yq -e '.packageManager.requiredPackages[]' "${CONFIG_PATH}" || echo "")"
	[ "${packages}" == "" ] \
		&& logInfo "No package to install" \
		&& return

	local config
	config="$(__get_package_manager_config)"

	local install update upgrade autoUpgrade
	install="$(read_yaml_stdin '.install' "${config}")"
	update="$(read_yaml_stdin '.update' "${config}")"
	upgrade="$(read_yaml_stdin '.upgrade' "${config}")"

	logInfo "Installing packages '${packages}'"
	IFS=$'\n' read -r -a package_array <<<"${packages}"
	${update}
	${install} "${package_array[@]}"
}
