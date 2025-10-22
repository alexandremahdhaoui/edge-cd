#!/usr/bin/env bash
set -euo pipefail

# Source the log.sh for logging functions
. "$(dirname "${BASH_SOURCE[0]}")/../../../cmd/edge-cd/lib/log.sh"

# Define SRC_DIR relative to the test script
SRC_DIR="$(dirname "${BASH_SOURCE[0]}")/../../.."

# General Test Setup
# Create a temporary directory and copy project files
TMP_DIR=$(mktemp -d)
cp -R "${SRC_DIR}/cmd/edge-cd" "${TMP_DIR}/edge-cd"

# Set CONFIG_PATH to a temporary config file within the test environment
export CONFIG_PATH="${TMP_DIR}/edge-cd/config.yaml"

# Set SRC_DIR for sourced scripts to use the temporary directory
export SRC_DIR="${TMP_DIR}/edge-cd"

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/config.sh"
. "${TMP_DIR}/edge-cd/lib/svcmgr.sh"

# Helper function for assertions
assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"
    if [[ "$expected" == "$actual" ]]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: '${expected}'"
        logErr "  Actual:   '${actual}'"
        exit 1
    fi
}

assert_empty() {
    local actual="$1"
    local message="$2"
    if [[ -z "$actual" ]]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: (empty)"
        logErr "  Actual:   '$actual'"
        exit 1
    fi
}

assert_not_empty() {
    local actual="$1"
    local message="$2"
    if [[ -n "$actual" ]]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: (not empty)"
        logErr "  Actual:   (empty)"
        exit 1
    fi
}

assert_exit_code() {
    local expected_code="$1"
    local command="$2"
    local message="$3"
    set +e # Disable exit on error for this check
    ( eval "$command" ) &> /dev/null # Redirect stdout/stderr to null for this check
    local actual_code=$?
    set -e # Re-enable exit on error
    if [[ "$expected_code" -eq "$actual_code" ]]; then
        logInfo "PASS: ${message} (Exit code: $actual_code)"
    else
        logErr "FAIL: ${message} (Expected exit code: $expected_code, Actual: $actual_code)"
        exit 1
    fi
}

assert_stderr_contains() {
    local expected_substring="$1"
    local command="$2"
    local message="$3"
    set +e
    local stderr_output
    stderr_output=$( ( eval "$command" ) 2>&1 >/dev/null )
    local actual_code=$?
    set -e
    if [[ "$stderr_output" == *"$expected_substring"* ]]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected stderr to contain: '${expected_substring}'"
        logErr "  Actual stderr: '${stderr_output}'"
        exit 1
    fi
}

logInfo "Starting svcmgr.sh unit tests..."

# Task 22: Test __get_svc_mgr_name
logInfo "Running Task 22: Test __get_svc_mgr_name"

# Test Case 22.1: Retrieve service manager name (procd)
echo "serviceManager:
  name: procd" > "${CONFIG_PATH}"

ACTUAL_NAME=$(__get_svc_mgr_name)
assert_equals "procd" "${ACTUAL_NAME}" "__get_svc_mgr_name: procd name retrieved"

logInfo "All Task 22 tests completed."

# Task 23: Test __get_svc_mgr_path
logInfo "Running Task 23: Test __get_svc_mgr_path"

# Test Case 23.1: Retrieve service manager path (procd)
echo "serviceManager:
  name: procd" > "${CONFIG_PATH}"

ACTUAL_PATH=$(__get_svc_mgr_path)
EXPECTED_PATH="${TMP_DIR}/edge-cd/service-managers/procd"
assert_equals "${EXPECTED_PATH}" "${ACTUAL_PATH}" "__get_svc_mgr_path: procd path retrieved"

logInfo "All Task 23 tests completed."

# Task 24: Test __read_svc_mgr_config
logInfo "Running Task 24: Test __read_svc_mgr_config"

# Test Case 24.1: Read from procd config
echo "serviceManager:
  name: procd" > "${CONFIG_PATH}"
mkdir -p "${TMP_DIR}/edge-cd/service-managers/procd"
echo "commands:
  enable:
    - echo
    - enable
  restart:
    - echo
    - restart" > "${TMP_DIR}/edge-cd/service-managers/procd/config.yaml"

ACTUAL_CONFIG=$(__read_svc_mgr_config '.commands.enable')
EXPECTED_CONFIG="- echo
- enable"
assert_equals "${EXPECTED_CONFIG}" "${ACTUAL_CONFIG}" "__read_svc_mgr_config: enable command retrieved"

logInfo "All Task 24 tests completed."

# Task 25: Test restart_service
logInfo "Running Task 25: Test restart_service"

# Test Case 25.1: Restart a service using fake-svc
echo "serviceManager:
  name: fake-svc" > "${CONFIG_PATH}"
mkdir -p "${TMP_DIR}/edge-cd/service-managers/fake-svc"
echo "commands:
  restart:
    - echo
    - restart
    - __SERVICE_NAME__" > "${TMP_DIR}/edge-cd/service-managers/fake-svc/config.yaml"

output=$(restart_service "my-app" 2>/dev/null)
assert_equals "restart my-app" "${output}" "restart_service: command executed with service name"

logInfo "All Task 25 tests completed."

# Task 26: Test enable_service
logInfo "Running Task 26: Test enable_service"

# Test Case 26.1: Enable a service using fake-svc
echo "serviceManager:
  name: fake-svc" > "${CONFIG_PATH}"
mkdir -p "${TMP_DIR}/edge-cd/service-managers/fake-svc"
echo "commands:
  enable:
    - echo
    - enable
    - __SERVICE_NAME__" > "${TMP_DIR}/edge-cd/service-managers/fake-svc/config.yaml"

output=$(enable_service "my-app" 2>/dev/null)
assert_equals "enable my-app" "${output}" "enable_service: command executed with service name"

logInfo "All Task 26 tests completed."

logInfo "All svcmgr.sh unit tests completed successfully!"
