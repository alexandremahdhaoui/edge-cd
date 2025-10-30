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

__LOADED_LIB_SVCMGR=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"
LIB_DIR="${SRC_DIR}/lib"
SVCMGR_DIR="${SRC_DIR}/service-managers"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"
[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"

# ------------------------------------------------------------------#
# Service Manager
# ------------------------------------------------------------------#

__SVCMGR_NAME=""
__get_svc_mgr_name() {
	if [ -z "${__SVCMGR_NAME}" ]; then
		__SVCMGR_NAME=$(read_config '.serviceManager.name')
	fi
	echo "${__SVCMGR_NAME}"
}

__SVCMGR_PATH=""
__get_svc_mgr_path() {
	if [ -z "${__SVCMGR_PATH}" ]; then
		__SVCMGR_PATH="${SVCMGR_DIR}/$(__get_svc_mgr_name)"
	fi
	echo "${__SVCMGR_PATH}"
}

__read_svc_mgr_config() {
	yamlPath="$1"
	read_yaml_file "${yamlPath}" "$(__get_svc_mgr_path)/config.yaml"
}

restart_service() {
	serviceName="$1"
	logInfo "Restarting service \"${serviceName}\""

	# Build command from YAML array, replacing __SERVICE_NAME__
	cmd=$(__read_svc_mgr_config '.commands.restart' | yq -e -r '.[]' | sed -e "s/__SERVICE_NAME__/${serviceName}/g" | tr '\n' ' ')

	# Execute the built command
	eval "${cmd}"
}

enable_service() {
	serviceName="$1"
	logInfo "Enabling service \"${serviceName}\""

	# Build command from YAML array, replacing __SERVICE_NAME__
	cmd=$(__read_svc_mgr_config '.commands.enable' | yq -e -r '.[]' | sed -e "s/__SERVICE_NAME__/${serviceName}/g" | tr '\n' ' ')

	# Execute the built command
	eval "${cmd}"
}
