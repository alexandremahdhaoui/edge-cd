#!/usr/bin/env bash

# --- Preamble ---
# Test for cmd/edge-cd/lib/svcmgr.sh

set -o errexit
set -o nounset
set -o pipefail

# --- Setup ---
SRC_DIR_OF_THIS_SCRIPT="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
LIB_DIR="${SRC_DIR_OF_THIS_SCRIPT}/../../../cmd/edge-cd/lib"

# Source the script to be tested
# Dependencies are not mocked to test the actual functionality.
source "${LIB_DIR}/svcmgr.sh"

# --- Mocks ---
export CONFIG_PATH
CONFIG_PATH="$(mktemp)"
export SVCMGR_DIR
SVCMGR_DIR="$(mktemp -d)"

# --- Test Runner & Assertions (omitted for brevity) ---
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
        rm -rf "${SVCMGR_DIR}"
        exit 0
    else
        echo "Failed tests: ${#FAILED_TESTS[@]}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo "  - ${failed_test}"
        done
        rm -f "${CONFIG_PATH}"
        rm -rf "${SVCMGR_DIR}"
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

mock_restart() { echo "restart_called $*"; }
export -f mock_restart

test_restart_service() {
    echo "serviceManager: { name: procd }" > "${CONFIG_PATH}"
    mkdir -p "${SVCMGR_DIR}"
    cat > "${SVCMGR_DIR}/procd" <<EOF
commands:
  restart:
    - mock_restart
    - __SERVICE_NAME__
EOF

    local output
    output="$(restart_service "myservice")"
    assert_equals "restart_called myservice" "${output}" "Restart command should be called with service name"
}

mock_enable() { echo "enable_called $*"; }
export -f mock_enable

test_enable_service() {
    echo "serviceManager: { name: procd }" > "${CONFIG_PATH}"
    mkdir -p "${SVCMGR_DIR}"
    cat > "${SVCMGR_DIR}/procd" <<EOF
commands:
  enable:
    - mock_enable
    - __SERVICE_NAME__
EOF

    local output
    output="$(enable_service "myservice")"
    assert_equals "enable_called myservice" "${output}" "Enable command should be called with service name"
}

# --- Main ---
run_test test_restart_service
run_test test_enable_service

report_results
