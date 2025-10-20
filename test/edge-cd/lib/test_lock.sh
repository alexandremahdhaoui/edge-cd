#!/usr/bin/env bash

# --- Preamble ---
# Test for cmd/edge-cd/lib/lock.sh

set -o errexit
set -o nounset
set -o pipefail

# --- Setup ---
SRC_DIR_OF_THIS_SCRIPT="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
LIB_DIR="${SRC_DIR_OF_THIS_SCRIPT}/../../../cmd/edge-cd/lib"
source "${LIB_DIR}/log.sh"
__LOADED_LIB_LOG=true
source "${LIB_DIR}/lock.sh"

export SCRIPT_NAME="edge-cd"

# --- Mocks ---
export LOCK_FILE_DIRNAME
LOCK_FILE_DIRNAME="$(mktemp -d)"


# Mock ps
ps() {
    # This mock will be redefined in each test case
    echo "ps mock not implemented for this test case" >&2
    exit 1
}
export -f ps

# --- Test Runner ---
TEST_COUNT=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    echo "--- Running test: ${test_name} ---"
    TEST_COUNT=$((TEST_COUNT + 1))
    ( # Run in a subshell to isolate environment changes and mocks
        set -o nounset # Ensure nounset is active in subshell
        # Setup for each test
        mkdir -p "${LOCK_FILE_DIRNAME}"
        # Run the test
        ${test_name}
        # Cleanup for each test
        rm -f "$(__get_lock_file_path)"
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
        rm -rf "${LOCK_FILE_DIRNAME}"
        exit 0
    else
        echo "Failed tests: ${#FAILED_TESTS[@]}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo "  - ${failed_test}"
        done
        rm -rf "${LOCK_FILE_DIRNAME}"
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

assert_file_exists() {
    local file_path="$1"
    local message="$2"
    if [[ ! -f "${file_path}" ]]; then
        echo "Assertion failed: ${message}" >&2
        echo "File does not exist: '${file_path}'" >&2
        exit 1
    fi
}

assert_file_not_exists() {
    local file_path="$1"
    local message="$2"
    if [[ -f "${file_path}" ]]; then
        echo "Assertion failed: ${message}" >&2
        echo "File should not exist: '${file_path}'" >&2
        exit 1
    fi
}

# --- Test Cases ---

test_lock_creates_file() {
    lock
    local lock_file
    lock_file="$(__get_lock_file_path)"
    assert_file_exists "${lock_file}" "Lock file should be created"
    local pid_in_file
    pid_in_file="$(cat "${lock_file}")"
    assert_equals "$$" "${pid_in_file}" "Lock file should contain correct PID"
}

test_lock_when_already_locked_by_self() {
    local lock_file
    lock_file="$(__get_lock_file_path)"
    echo "$$" > "${lock_file}"

    lock # Should return 0 and not change the file

    local pid_in_file
    pid_in_file="$(cat "${lock_file}")"
    assert_equals "$$" "${pid_in_file}" "Lock file should still contain own PID"
}

test_lock_fails_if_other_process_holds_lock() {
    local other_pid=12345
    local lock_file
    lock_file="$(__get_lock_file_path)"
    echo "${other_pid}" > "${lock_file}"

    ps() {
        echo "USER   PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND"
        echo "root ${other_pid}  0.0  0.0  12345  1234 ?        Ss   00:00   0:00 bash"
    }
    export -f ps

    local status=0
    (lock) || status=$?
    assert_equals "1" "${status}" "Lock should fail when another process holds the lock"
}

test_lock_takes_over_stale_lock() {
    local other_pid=12345
    local lock_file
    lock_file="$(__get_lock_file_path)"
    echo "${other_pid}" > "${lock_file}"

    ps() {
        echo "  PID TTY          TIME CMD"
        # Don't output the other_pid, simulating it's not running
    }
    export -f ps

    lock

    assert_file_exists "${lock_file}" "Lock file should exist"
    local pid_in_file
    pid_in_file="$(cat "${lock_file}")"
    assert_equals "$$" "${pid_in_file}" "Should take over stale lock and write own PID"
}

test_unlock_removes_file() {
    local lock_file
    lock_file="$(__get_lock_file_path)"
    touch "${lock_file}"

    unlock

    assert_file_not_exists "${lock_file}" "Unlock should remove the lock file"
}

# --- Main ---
run_test test_lock_creates_file
run_test test_lock_when_already_locked_by_self
run_test test_lock_fails_if_other_process_holds_lock
run_test test_lock_takes_over_stale_lock
run_test test_unlock_removes_file

report_results
