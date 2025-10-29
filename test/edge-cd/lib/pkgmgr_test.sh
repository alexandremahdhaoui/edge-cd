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

# Set up config path variables for testing
# CONFIG_PATH is the relative directory path within the config repo
export CONFIG_PATH="."
# CONFIG_REPO_DEST_PATH is where the config repo is located
export CONFIG_REPO_DEST_PATH="${TMP_DIR}/edge-cd"
# CONFIG_SPEC_FILE is the name of the spec file
export CONFIG_SPEC_FILE="config.yaml"
# Set default values that would normally come from edge-cd script
export __DEFAULT_CONFIG_REPO_DEST_PATH="/usr/local/src/edge-cd-config"
export __DEFAULT_CONFIG_SPEC_FILE="spec.yaml"

# Set SRC_DIR for sourced scripts to use the temporary directory
export SRC_DIR="${TMP_DIR}/edge-cd"

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/config.sh"
. "${TMP_DIR}/edge-cd/lib/pkgmgr.sh"

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

logInfo "Starting pkgmgr.sh unit tests..."

# Task 15: Test __get_package_manager_config (Valid Manager)
logInfo "Running Task 15: Test __get_package_manager_config (Valid Manager)"

# Test Case 15.1: Valid package manager (opkg)
echo "packageManager:
  name: opkg" > "$(get_config_spec_abspath)"

mkdir -p "${TMP_DIR}/edge-cd/package-managers"
echo "update: ["echo", "opkg", "update"]
upgrade: ["echo", "opkg", "upgrade"]
install: ["echo", "opkg", "install"]" > "${TMP_DIR}/edge-cd/package-managers/opkg.yaml"



EXPECTED_CONFIG="update: [echo, opkg, update]
upgrade: [echo, opkg, upgrade]
install: [echo, opkg, install]"
ACTUAL_CONFIG=$(__get_package_manager_config)
assert_equals "${EXPECTED_CONFIG}" "${ACTUAL_CONFIG}" "__get_package_manager_config: opkg config retrieved"

logInfo "All Task 15 tests completed."

# Task 16: Test __get_package_manager_config (Invalid Manager)
logInfo "Running Task 16: Test __get_package_manager_config (Invalid Manager)"

# Test Case 16.1: Non-existent package manager
echo "packageManager:
  name: nonexistent-pkg" > "$(get_config_spec_abspath)"

assert_stderr_contains "[ERROR] Package manager \"nonexistent-pkg\" cannot be found" "__get_package_manager_config" "exits with error for non-existent manager"
assert_exit_code 1 "__get_package_manager_config" "exits with status 1 for non-existent manager"

logInfo "All Task 16 tests completed."

# Task 17: Test __get_package_manager_config (CUSTOM Manager)
logInfo "Running Task 17: Test __get_package_manager_config (CUSTOM Manager)"

# Test Case 17.1: CUSTOM package manager
echo "packageManager:
  name: CUSTOM" > "$(get_config_spec_abspath)"

assert_stderr_contains "Not implemented yet" "__get_package_manager_config" "exits with error for CUSTOM manager"
assert_exit_code 1 "__get_package_manager_config" "exits with status 1 for CUSTOM manager"

logInfo "All Task 17 tests completed."

# Task 18: Test reconcile_package_auto_upgrade (Auto-upgrade Enabled)
logInfo "Running Task 18: Test reconcile_package_auto_upgrade (Auto-upgrade Enabled)"

# Test Case 18.1: Auto-upgrade enabled, packages specified
echo "packageManager:
  name: fake-pkg
  autoUpgrade: true
  requiredPackages: [\"pkg1\", \"pkg2\"]" > "$(get_config_spec_abspath)"

# Create a fake package manager config that echoes its commands
mkdir -p "${TMP_DIR}/edge-cd/package-managers"
echo "update:
  - echo
  - FAKE_UPDATE
upgrade:
  - echo
  - FAKE_UPGRADE" > "${TMP_DIR}/edge-cd/package-managers/fake-pkg.yaml"
output=$(reconcile_package_auto_upgrade 2>/dev/null)

assert_equals "FAKE_UPDATE" "$(echo "${output}" | head -n 1)" "reconcile_package_auto_upgrade: update command executed"
assert_equals "FAKE_UPGRADE pkg1 pkg2" "$(echo "${output}" | tail -n 1)" "reconcile_package_auto_upgrade: upgrade command executed with packages"

logInfo "All Task 18 tests completed."

# Task 19: Test reconcile_package_auto_upgrade (Auto-upgrade Disabled)
logInfo "Running Task 19: Test reconcile_package_auto_upgrade (Auto-upgrade Disabled)"

# Test Case 19.1: Auto-upgrade disabled
echo "packageManager:
  name: fake-pkg
  autoUpgrade: false
  requiredPackages: [\"pkg1\", \"pkg2\"]" > "$(get_config_spec_abspath)"

# Capture stdout/stderr
output=$(reconcile_package_auto_upgrade 2>&1 >/dev/null)

assert_empty "${output}" "reconcile_package_auto_upgrade: no commands executed when disabled"

logInfo "All Task 19 tests completed."

# Task 20: Test reconcile_packages (Packages to Install)
logInfo "Running Task 20: Test reconcile_packages (Packages to Install)"

# Test Case 20.1: Packages to install
echo "packageManager:
  name: fake-pkg
  requiredPackages: [\"pkgA\", \"pkgB\"]" > "$(get_config_spec_abspath)"

# Create a fake package manager config that echoes its commands
echo "update:
  - echo
  - FAKE_UPDATE
install:
  - echo
  - FAKE_INSTALL" > "${TMP_DIR}/edge-cd/package-managers/fake-pkg.yaml"

# Capture stdout/stderr
output=$(reconcile_packages 2>/dev/null)

assert_equals "FAKE_UPDATE" "$(echo "${output}" | head -n 1)" "reconcile_packages: update command executed"
assert_equals "FAKE_INSTALL pkgA pkgB" "$(echo "${output}" | tail -n 1)" "reconcile_packages: install command executed with packages"

logInfo "All Task 20 tests completed."

# Task 21: Test reconcile_packages (No Packages to Install)
logInfo "Running Task 21: Test reconcile_packages (No Packages to Install)"

# Test Case 21.1: No packages to install (empty list)
echo "packageManager:
  name: fake-pkg
  requiredPackages: []" > "$(get_config_spec_abspath)"

# Capture stdout/stderr
output=$(reconcile_packages 2>&1)

assert_stderr_contains "[INFO] No package to install" "reconcile_packages" "logs 'No package to install'"
assert_empty "$(echo "${output}" | grep -v "No package to install")" "reconcile_packages: no commands executed when no packages"

# Test Case 21.2: No packages to install (missing section)
echo "packageManager:
  name: fake-pkg" > "$(get_config_spec_abspath)"

output=$(reconcile_packages 2>&1)

assert_stderr_contains "[INFO] No package to install" "reconcile_packages" "logs 'No package to install' when section missing"
assert_empty "$(echo "${output}" | grep -v "No package to install")" "reconcile_packages: no commands executed when section missing"

logInfo "All Task 21 tests completed."

logInfo "All pkgmgr.sh unit tests completed successfully!"
