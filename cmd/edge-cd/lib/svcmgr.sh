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

declare -g __LOADED_LIB_SVCMGR=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")}"
LIB_DIR="${SRC_DIR}"
SVCMGR_DIR="${SRC_DIR}/../service-managers"

# ------------------------------------------------------------------#
# Imports
# ------------------------------------------------------------------#

[ -z "${__LOADED_LIB_CONFIG:-}" ] && . "${LIB_DIR}/config.sh"
[ -z "${__LOADED_LIB_LOG:-}" ] && . "${LIB_DIR}/log.sh"

# ------------------------------------------------------------------#
# Service Manager
# ------------------------------------------------------------------#

declare __SVCMGR_NAME
function __get_svc_mgr_name() {
	echo "${__SVCMGR_NAME:=$(read_config '.serviceManager.name')}"
}

declare __SVCMGR_PATH
function __get_svc_mgr_path() {
	echo "${__SVCMGR_PATH:=${SVCMGR_DIR}/$(__get_svc_mgr_name)}"
}

function __read_svc_mgr_config() {
	local yamlPath
	read_yaml_file "${yamlPath}" "$(__get_svc_mgr_path)"
}

function restart_service() {
	local serviceName="${1}"
	logInfo "Restarting service \"${serviceName}\""

	local -a cmd
	readarray -t cmd < <(__read_svc_mgr_config '.commands.restart' | sed "s/__SERVICE_NAME__/${serviceName}/g")
	"${cmd[@]}"
}

function enable_service() {
	local serviceName="${1}"
	logInfo "Enabling service \"${serviceName}\""

	local -a cmd
	readarray -t cmd < <(__read_svc_mgr_config '.commands.enable' | sed "s/__SERVICE_NAME__/${serviceName}/g")
	"${cmd[@]}"
}
