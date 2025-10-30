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

__LOADED_LIB_CONFIG=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"

# ------------------------------------------------------------------#
# Extra Envs
# ------------------------------------------------------------------#

# Store backup env var names as newline-separated string
BACKUP_EXTRA_ENVS=""

get_config_spec_abspath() {
	# Use defaults if variables are not set
	repo_path="${CONFIG_REPO_DEST_PATH:-${__DEFAULT_CONFIG_REPO_DEST_PATH}}"
	config_path="${CONFIG_PATH}"  # Required - no default
	spec_file="${CONFIG_SPEC_FILE:-${__DEFAULT_CONFIG_SPEC_FILE}}"

	# CONFIG_PATH is required
	if [ -z "${config_path}" ]; then
		echo "[ERROR] CONFIG_PATH must be set" >&2
		exit 1
	fi

	echo "${repo_path}/${config_path}/${spec_file}"
}

set_extra_envs() {
	# Clear the list of variables to reset from any previous run
	BACKUP_EXTRA_ENVS=""

	yq_output=$(yq '(.extraEnvs // []) | .[] | to_entries | .[] | .key + "=" + .value' "$(get_config_spec_abspath)" 2>/dev/null || true)

	if [ -n "${yq_output}" ]; then # Only process if yq returned something
		# Use heredoc to avoid subshell issue with pipe
		while IFS='=' read -r key value || [ -n "$key" ]; do
			[ -z "$key" ] && continue # skip malformed lines

			# Add to backup list
			if [ -z "${BACKUP_EXTRA_ENVS}" ]; then
				BACKUP_EXTRA_ENVS="$key"
			else
				BACKUP_EXTRA_ENVS="${BACKUP_EXTRA_ENVS}
$key"
			fi

			backup_key="__BACKUP_${key}"
			# Check if variable is set using eval
			if eval "[ -n \"\${${key}+x}\" ]"; then
				# Backup the current value using eval
				eval "export ${backup_key}=\"\${${key}}\""
			else
				# Use a special marker to indicate the variable was originally unset
				eval "export ${backup_key}=__WAS_UNSET__"
			fi

			eval "export ${key}=\"${value}\""
		done <<EOF
${yq_output}
EOF
	fi
}

reset_extra_envs() {
	if [ -n "${BACKUP_EXTRA_ENVS}" ]; then
		# Use heredoc to avoid subshell issue with pipe
		while IFS= read -r var_name || [ -n "$var_name" ]; do
			[ -z "${var_name}" ] && continue

			backup_name="__BACKUP_${var_name}"

			# Check if backup variable exists
			if eval "[ -n \"\${${backup_name}+x}\" ]"; then
				backup_value=$(eval "echo \"\${${backup_name}}\"")
				if [ "$backup_value" = "__WAS_UNSET__" ]; then
					eval "unset ${var_name}"
				else
					eval "export ${var_name}=\"${backup_value}\""
				fi
			else
				echo "[WARN] Backup variable $backup_name not found: Skipping reset for $var_name" >&2
			fi
		done <<EOF
${BACKUP_EXTRA_ENVS}
EOF
	fi
}

# ------------------------------------------------------------------#
# Read value/config
# ------------------------------------------------------------------#

read_yaml_stdin() {
	yamlPath="$1"
	yamlContent="$2"
	printf '%s\n' "${yamlContent}" | yq -e "${yamlPath}"
}

read_yaml_stdin_optional() {
	yamlPath="$1"
	yamlContent="$2"
	result=$(printf '%s\n' "${yamlContent}" | yq "${yamlPath}")
	if [ "${result}" = "null" ]; then
		echo ""
	else
		echo "${result}"
	fi
}

read_yaml_file() {
	yamlPath="$1"
	yamlFile="$2"
	yq -e "${yamlPath}" "${yamlFile}"
}

read_yaml_file_optional() {
	yamlPath="$1"
	yamlFile="$2"
	# Check if file exists before calling yq
	if [ ! -f "${yamlFile}" ]; then
		echo "null"
		return 0
	fi
	yq "${yamlPath}" "${yamlFile}"
}

read_config() {
	yamlPath="$1"
	read_yaml_file "${yamlPath}" "$(get_config_spec_abspath)"
}

read_config_optional() {
	yamlPath="$1"
	read_yaml_file_optional "${yamlPath}" "$(get_config_spec_abspath)"
}

# -- reads config in the following order of precedence:
#    1. Environment variables
#    2. Config
#    3. [Optional] Default value
# The function fails if there is no value var in envs
# or in the yamlPath and if specfied in the default value
# A value equal to "null" is treated like an unset or empty value,
# unless it is the default value.
read_value() {
	env_var_name="$1"
	yamlPath="$2"
	defaultValue="${3:-}"

	# 1. Check Environment Variable using eval for indirect expansion
	if [ -n "${env_var_name}" ]; then
		env_value=$(eval "echo \"\${${env_var_name}:-}\"")
		if [ -n "${env_value}" ] && [ "${env_value}" != "null" ]; then
			printf '%s' "${env_value}"
			return
		fi
	fi

	# 2. Check Configuration File
	valueFromConfig="$(read_config_optional "${yamlPath}" 2>/dev/null)"
	if [ -n "${valueFromConfig}" ] && [ "${valueFromConfig}" != "null" ]; then
		printf '%s' "${valueFromConfig}"
		return
	fi

	# 3. Use Default Value
	if [ -n "${defaultValue}" ]; then
		printf '%s' "${defaultValue}"
		return
	fi

	env_value=$(eval "echo \"\${${env_var_name}:-}\"")
	echo 1>&2 "[ERROR] cannot read value from config; args: env_var_name=${env_var_name}, !env_var_name=${env_value}, yamlPath=${yamlPath}, defaultValue=${defaultValue}"
	exit 1
}
