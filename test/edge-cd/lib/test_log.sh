#!/usr/bin/env bash

# --- Preamble ---
# Test for cmd/edge-cd/lib/log.sh

set -o errexit
set -o nounset
set -o pipefail

# --- Setup ---
SRC_DIR_OF_THIS_SCRIPT="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
LIB_DIR="${SRC_DIR_OF_THIS_SCRIPT}/../../../cmd/edge-cd/lib"
source "${LIB_DIR}/log.sh"

# --- Test Runner ---
TEST_COUNT=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    echo "--- Running test: ${test_name} ---"
    TEST_COUNT=$((TEST_COUNT + 1))
    ( # Run in a subshell to isolate environment changes
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
        exit 0
    else
        echo "Failed tests: ${#FAILED_TESTS[@]}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo "  - ${failed_test}"
        done
        exit 1
    fi
}

# --- Assertions ---
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

test_logInfo_console() {
    export LOG_FORMAT="console"
    local output
    output="$(logInfo "hello world" 2>&1)"
    assert_equals "[INFO] hello world" "${output}" "logInfo console format"
}

test_logErr_console() {
    export LOG_FORMAT="console"
    local output
    output="$(logErr "hello error" 2>&1)"
    assert_equals "[ERROR] hello error" "${output}" "logErr console format"
}

test_logInfo_json() {
    export LOG_FORMAT="json"
    local output
    output="$(logInfo "hello world" 2>&1)"
    assert_equals "{\"level\":\"info\",\"message\":\"hello\\ world\"}" "${output}" "logInfo json format"
}

test_logErr_json() {
    export LOG_FORMAT="json"
    local output
    output="$(logErr "hello error" 2>&1)"
    assert_equals "{\"level\":\"error\",\"message\":\"hello\\ error\"}" "${output}" "logErr json format"
}

test_log_xtrace_restoration() {
    set -o xtrace
    logInfo "testing xtrace"
    if [[ "$- " != *x* ]]; then
      assert_equals "xtrace on" "xtrace off" "xtrace should be restored"
    fi
}

# --- Main ---
run_test test_logInfo_console
run_test test_logErr_console
run_test test_logInfo_json
run_test test_logErr_json
run_test test_log_xtrace_restoration

report_results
