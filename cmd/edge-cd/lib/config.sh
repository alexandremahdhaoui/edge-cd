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

declare -g __LOADED_LIB_CONFIG=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")}"

# ------------------------------------------------------------------#
# Extra Envs
# ------------------------------------------------------------------#

declare -ga BACKUP_EXTRA_ENVS

function set_extra_envs() {
	# Clear the list of variables to reset from any previous run
	BACKUP_EXTRA_ENVS=()

	local key value
	while IFS='=' read -r key value; do
		[[ -z "$key" ]] && continue # skip malformed lines

		BACKUP_EXTRA_ENVS+=("$key")
		local backup_key="__BACKUP_${key}"
		if [[ -v "$key" ]]; then
			# Backup the current value.
			declare -g "$backup_key=${!key}"
		else
			# Use a special marker to indicate the variable was originally unset.
			declare -g "$backup_key=__WAS_UNSET__"
		fi

		declare -gx "$key=$value"
	done < <(yq -e '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' "${CONFIG_PATH}")
}

function reset_extra_envs() {
	for var_name in "${BACKUP_EXTRA_ENVS[@]}"; do
		local backup_name="__BACKUP_${var_name}"

		if [[ -v "$backup_name" ]]; then
			local backup_value="${!backup_name}"
			if [[ "$backup_value" == "__WAS_UNSET__" ]]; then
				unset "$var_name"
			else
				declare -gx "$var_name=$backup_value"
			fi
		else
			echo "[WARN] Backup variable $backup_name not found: Skipping reset for $var_name" >&2
		fi
	done
}

# ------------------------------------------------------------------#
# Read value/config
# ------------------------------------------------------------------#

function read_yaml_stdin() {
	local yamlPath="${1}"
	local yamlContent="${2}"
	echo -e "${yamlContent}" | yq -e "${yamlPath}"
}

function read_yaml_stdin_optional() {
	local yamlPath="${1}"
	local yamlContent="${2}"
	echo -e "${yamlContent}" | yq "${yamlPath}"
}

function read_yaml_file() {
	local yamlPath="${1}"
	local yamlFile="${2}"
	yq -e -r "${yamlPath}" "${yamlFile}"
}

function read_config() {
	local yamlPath="${1}"
	read_yaml_file "${yamlPath}" "${CONFIG_PATH}"
}

# -- reads config in the following order of precedence:
#    1. Environment variables
#    2. Config
#    3. [Optional] Default value
# The function fails if there is no value var in envs
# or in the yamlPath and if specfied in the default value
# A value equal to "null" is treated like an unset or empty value,
# unless it is the default value.
function read_value() {
	local env_var_name="${1}"
	local yamlPath="${2}"
	local defaultValue="${3:-""}"

	# 1. Check Environment Variable
	if [[ -n "${env_var_name}" ]]; then
		if [[ -n "${!env_var_name:-}" ]] && [[ "${!env_var_name:-}" != "null" ]]; then
			printf '%s' "${!env_var_name}"
			return
		fi
	fi

	# 2. Check Configuration File
	local valueFromConfig
	valueFromConfig="$(read_config "${yamlPath}" 2>/dev/null || true)"

	if [[ -n "${valueFromConfig}" && "${valueFromConfig}" != "null" ]]; then
		printf '%s' "${valueFromConfig}"
		return
	fi

	# 3. Use Default Value
	if [[ -n "${defaultValue}" ]]; then
		printf '%s' "${defaultValue}"
		return # Return successfully if default value is used
	fi

	echo 1>&2 "[ERROR] cannot read value from config; args: env_var_name=${env_var_name}, !env_var_name=${!env_var_name:-}, yamlPath=${yamlPath}, defaultValue=${defaultValue}"
	exit 1
}
