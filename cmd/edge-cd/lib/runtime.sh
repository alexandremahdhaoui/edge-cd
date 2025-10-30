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

__LOADED_LIB_RUNTIME=true
SRC_DIR="${SRC_DIR:-$(dirname "$(readlink -f "$0")")}"

# ------------------------------------------------------------------#
# Runtime Variables
# ------------------------------------------------------------------#

declare_runtime_variables() {
	# Variables are implicitly declared in POSIX shell
	# This function serves as documentation of expected variables:
	# - RTV_REQUIRE_SERVICE_RESTART: newline-separated list of services to restart
	# - RTV_REQUIRE_SELF_RESTART: boolean flag for edge-cd self-restart
	# - RTV_REQUIRE_REBOOT: boolean flag for system reboot
	:
}

reset_runtime_variables() {
	RTV_REQUIRE_SERVICE_RESTART=""
	RTV_REQUIRE_SELF_RESTART=false
	RTV_REQUIRE_REBOOT=false
}

add_service_for_restart() {
	service_name="$1"

	if [ -z "${service_name}" ]; then
		return 1
	fi

	# Add service to the list with newline separator
	RTV_REQUIRE_SERVICE_RESTART="${RTV_REQUIRE_SERVICE_RESTART}${service_name}
"
}

get_services_to_restart() {
	# Return unique sorted list of services to restart
	# Filter out empty lines
	services_list=$(printf '%s\n' "${RTV_REQUIRE_SERVICE_RESTART}" | grep -v '^\s*$' | sort | uniq)
	echo "${services_list}"
}

should_restart() {
	if [ "${RTV_REQUIRE_REBOOT}" != "true" ]; then
		return 1
	fi
	return 0
}
