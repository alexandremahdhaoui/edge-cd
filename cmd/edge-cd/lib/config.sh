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
# Extra Envs
# ------------------------------------------------------------------#

function set_extra_envs() {
	declare -a BACKUP_EXTRA_ENVS

	local key value
	while IFS='=' read -r key value; do
		[[ -z "$key" ]] && continue # skip malformed lines
		declare -x "$key=$value"

		BACKUP_EXTRA_ENVS+=("$key")
		backup_key="__BACKUP_${key}"
		declare "$backup_key=$value"
		declare -gx "$key=$value"
	done < <(yq -e '.extraEnvs[] | to_entries[] | "\(.key)=\(.value)"' "${CONFIG_PATH}")
}

function reset_extra_envs() {
	for var_name in "${BACKUP_EXTRA_ENVS[@]}"; do
		local backup_name="__BACKUP_${var_name}"

		# Check if the backup variable is set
		if [[ -v $backup_name ]]; then
			local backup_value
			backup_value=$(printf '%s' "${!backup_name}")
			declare -gx "$var_name=$backup_value"
		else
			logWarn "Backup variable $backup_name not found. Skipping reset for $var_name."
		fi
	done
}

# ------------------------------------------------------------------#
# Read value/config
# ------------------------------------------------------------------#

function read_yaml_stdin() {
	local yamlPath="${1}"
	local yamlContent="${2}"
	echo "${yamlContent}" | yq -e "${yamlPath}"
}

function read_yaml_file() {
	local yamlPath="${1}"
	local yamlFile="${2}"
	yq -e "${yamlPath}" "${yamlFile}"
}

function read_config() {
	local yamlPath="${1}"
	read_yaml_file "${CONFIG_PATH}"
}

# -- reads config in the following order of precedence:
#    1. Environment variables
#    2. Config
#    3. [not implemented yet] Default value
function read_value() {
	local env="${1}"
	local yamlPath="${2}"
	local defaultValue="${3:-""}"

	local valueFromConfig
	valueFromConfig="$(read_config "${yamlPath}")"
	echo "${!env:-${valueFromConfig:-${defaultValue}}}"
}
