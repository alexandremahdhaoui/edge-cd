#!/usr/bin/env bash

# --- Preamble ---
# Test for cmd/edge-cd/lib/pkgmgr.sh

set -o errexit
set -o nounset
set -o pipefail

# --- Setup ---
SRC_DIR_OF_THIS_SCRIPT="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
LIB_DIR="${SRC_DIR_OF_THIS_SCRIPT}/../../../cmd/edge-cd/lib"

# Mock dependencies before sourcing the script
logInfo() { :; }
logErr() { :; }
read_config() { :; }
read_yaml_stdin() { :; }
yq() { :; }

# Source the script to be tested
source "${LIB_DIR}/pkgmgr.sh"

# --- Mocks ---
export CONFIG_PATH
CONFIG_PATH="$(mktemp)"
export PKGMGR_DIR
PKGMGR_DIR="$(mktemp -d)"

# --- Test Runner & Assertions (omitted for brevity, similar to other test files) ---
TEST_COUNT=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    echo "--- Running test: ${test_name} ---"
    TEST_COUNT=$((TEST_COUNT + 1))
    ( 
        set -o nounset
        ${test_name}
    )
    if [[ $? -ne 0 ]]; then
        FAILED_TESTS+=("${test_name}")
        echo "--- FAILED: ${test_name} ---"
    else
        echo "--- PASSED: ${test_name} ---"
    fi
}

report_results() {
    echo
    echo "--- Test Results ---"
    echo "Total tests: ${TEST_COUNT}"
    if [[ ${#FAILED_TESTS[@]} -eq 0 ]]; then
        echo "All tests passed!"
        rm -f "${CONFIG_PATH}"
        rm -rf "${PKGMGR_DIR}"
        exit 0
    else
        echo "Failed tests: ${#FAILED_TESTS[@]}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo "  - ${failed_test}"
        done
        rm -f "${CONFIG_PATH}"
        rm -rf "${PKGMGR_DIR}"
        exit 1
    fi
}

assert_equals() {
    local expected="$1"
    local actual="$2"
    local message="$3"

    if [[ "${expected}" != "${actual}" ]]; then
        echo "Assertion failed: ${message}" >&2
        echo "Expected: '${expected}'" >&2
        echo "Actual:   '${actual}'" >&2
        exit 1
    fi
}

# --- Test Cases ---

test_reconcile_package_auto_upgrade_false() {
    yq() { echo "false"; }
    export -f yq
    # This should do nothing and return 0
    reconcile_package_auto_upgrade
}

test_reconcile_package_auto_upgrade_true() {
    yq() { echo "true"; }
    
    __get_package_manager_config() { echo "fake_config"; }

    mock_update() { echo "update_called"; }
    mock_upgrade() { echo "upgrade_called"; }
    export -f mock_update mock_upgrade

    read_yaml_stdin() {
        if [[ "$1" == ".update[]" ]]; then echo "mock_update"; fi
        if [[ "$1" == ".upgrade[]" ]]; then echo "mock_upgrade"; fi
    }
    export -f __get_package_manager_config read_yaml_stdin

    local output
    output="$(reconcile_package_auto_upgrade)"
    assert_equals $'update_called\nupgrade_called' "${output}" "Upgrade commands should be called"
}

test_reconcile_packages_no_packages() {
    yq() { echo ""; }
    logInfo() { echo "$*"; }
    export -f yq logInfo

    local output
    output="$(reconcile_packages)"
    assert_equals "No package to install" "${output}" "Should log no packages to install"
}

test_reconcile_packages_with_packages() {
    yq() { echo $'pkg1\npkg2'; }

    __get_package_manager_config() { echo "fake_config"; }

    mock_update() { echo "update_called"; }
    mock_install() { echo "install_called $*" ; }
    export -f mock_update mock_install

    read_yaml_stdin() {
        if [[ "$1" == ".update[]" ]]; then echo "mock_update"; fi
        if [[ "$1" == ".install[]" ]]; then echo "mock_install"; fi
    }
    export -f yq __get_package_manager_config read_yaml_stdin

    local output
    output="$(reconcile_packages)"
    assert_equals $'update_called\ninstall_called pkg1 pkg2' "${output}" "Install commands should be called with packages"
}


# --- Main ---
run_test test_reconcile_package_auto_upgrade_false
run_test test_reconcile_package_auto_upgrade_true
run_test test_reconcile_packages_no_packages
run_test test_reconcile_packages_with_packages

report_results
