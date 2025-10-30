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

__LOADED_LIB_FILES=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
LIB_DIR="${SRC_DIR}/lib"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"
[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"
[ -z "${__LOADED_LIB_RUNTIME:-}" ] && . "${LIB_DIR}/runtime.sh"

# ------------------------------------------------------------------#
# File Reconciliation
# ------------------------------------------------------------------#

reconcile_file() {
	srcPath="$1"
	destPath="$2"
	fileMod="$3"
	restartServices="$4"
	requireReboot="$5"

	# -- validate inputs
	if [ -z "${srcPath}" ] || [ "${srcPath}" = "null" ]; then
		logErr "srcPath must be specified"
		exit 1
	fi

	if [ -z "${destPath}" ] || [ "${destPath}" = "null" ]; then
		logErr "destPath must be specified"
		exit 1
	elif ! echo "${destPath}" | grep -q '^/'; then
		logErr "destPath must be an absolute path"
		exit 1
	fi

	# -- set defaults
	if [ -z "${fileMod}" ] || [ "${fileMod}" = "null" ]; then fileMod=644; fi
	if [ "${restartServices}" = "null" ]; then restartServices=""; fi

	# -- compare src and dest
	if cmp "${srcPath}" "${destPath}" 2>/dev/null; then
		return 0
	fi

	logInfo "Drift detected: updating file \"${destPath}\""
	mkdir -p "$(dirname "${destPath}")"

	# mark associated services for restart
	services_to_restart=$(echo "${restartServices}" | yq -e -r '.[]' 2>/dev/null || true)
	if [ -n "${services_to_restart}" ]; then
		while IFS= read -r service || [ -n "$service" ]; do
			[ -z "${service}" ] && continue
			add_service_for_restart "${service}"
		done <<EOF
${services_to_restart}
EOF
	fi

	# -- set require reboot flag
	if [ "${requireReboot}" = "true" ]; then
		RTV_REQUIRE_REBOOT=true
	fi

	cp -f "${srcPath}" "${destPath}"
	chmod "${fileMod}" "${destPath}"
}

reconcile_file_spec() {
	fileSpec="$1"

	type="$(read_yaml_stdin '.type' "${fileSpec}")"

	case "${type}" in
		file | directory)
			srcPath="$(read_yaml_stdin '.srcPath' "${fileSpec}")"
			;;
		content)
			content="$(read_yaml_stdin '.content' "${fileSpec}")"
			;;
		*)
			logErr "Unknown type=\"${type}\" for destPath=\"${destPath}\""
			exit 1
			;;
	esac

	destPath="$(read_yaml_stdin '.destPath' "${fileSpec}")"
	fileMod="$(read_yaml_stdin_optional '.fileMod' "${fileSpec}")"
	restartServices="$(read_yaml_stdin_optional '.syncBehavior.restartServices' "${fileSpec}")"
	requireReboot="$(read_yaml_stdin_optional '.syncBehavior.reboot' "${fileSpec}")"

	# TODO: Add a "git" type? E.g.: clone any repo and copy file from repo to the dest
	case "${type}" in
		file)
			srcPath="${CONFIG_REPO_DEST_PATH}/${CONFIG_PATH}/${srcPath}"
			logInfo "File type: srcPath=${srcPath}, destPath=${destPath}"
			;;

		# -- directory does not support recursive
		directory)
			srcPath="${CONFIG_REPO_DEST_PATH}/${CONFIG_PATH}/${srcPath}"
			dir_files=$(find "${srcPath}" -type f ! -name '\.*' -exec readlink -f {} \;)
			if [ -n "${dir_files}" ]; then
				while IFS= read -r srcPath || [ -n "$srcPath" ]; do
					[ -z "${srcPath}" ] && continue
					reconcile_file "${srcPath}" "${destPath}" "${fileMod}" "${restartServices}" "${requireReboot}"
				done <<EOF
${dir_files}
EOF
			fi
			# -- directory returns early
			return
			;;

		content)
			# -- type content
			srcPath="$(mktemp)"
			printf '%s\n' "${content}" >"${srcPath}"
			;;

		*)
			logErr "Unknown type=\"${type}\" for destPath=\"${destPath}\""
			exit 1
			;;
	esac

	# -- sync file
	reconcile_file "${srcPath}" "${destPath}" "${fileMod}" "${restartServices}" "${requireReboot}"
}

reconcile_files() {
	logInfo "Reconciling files"

	config_spec_path="$(get_config_spec_abspath)"
	logInfo "Reading files from config: ${config_spec_path}"

	files_json=$(yq e -o=j -I=0 '.files[]' "${config_spec_path}" 2>/dev/null || true)
	logInfo "Found files count: $(echo \"${files_json}\" | grep -c . || echo 0)"

	if [ -n "${files_json}" ]; then
		while IFS= read -r fileSpec || [ -n "$fileSpec" ]; do
			[ -z "${fileSpec}" ] && continue
			logInfo "Processing file spec: ${fileSpec}"
			reconcile_file_spec "${fileSpec}"
		done <<EOF
${files_json}
EOF
	fi
}
