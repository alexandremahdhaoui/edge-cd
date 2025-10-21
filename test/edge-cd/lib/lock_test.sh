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

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/lock.sh"

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
    ( eval "$command" )
    local actual_code=$?
    set -e # Re-enable exit on error
    if [[ "$expected_code" -eq "$actual_code" ]]; then
        logInfo "PASS: ${message} (Exit code: $actual_code)"
    else
        logErr "FAIL: ${message} (Expected exit code: $expected_code, Actual: $actual_code)"
        exit 1
    fi
}

logInfo "Starting lock.sh unit tests..."

# Task 9: Test __get_lock_file_path
logInfo "Running Task 9: Test __get_lock_file_path"

# Test Case 9.1: LOCK_FILE_DIRNAME unset (default value)
unset LOCK_FILE_DIRNAME
EXPECTED="/tmp/edge-cd/edge-cd.lock"
ACTUAL=$(__get_lock_file_path)
assert_equals "${EXPECTED}" "${ACTUAL}" "__get_lock_file_path: default value"

# Test Case 9.2: LOCK_FILE_DIRNAME set to a custom value
export LOCK_FILE_DIRNAME="/var/run"
EXPECTED="/var/run/edge-cd.lock"
ACTUAL=$(__get_lock_file_path)
assert_equals "${EXPECTED}" "${ACTUAL}" "__get_lock_file_path: custom value"

logInfo "All Task 9 tests completed."

unset LOCK_FILE_DIRNAME # Reset LOCK_FILE_DIRNAME to default writable path

# Task 10: Test lock (Acquire Lock)
logInfo "Running Task 10: Test lock (Acquire Lock)"

# Test Case 10.1: Acquire lock when no other process holds it.
LOCK_FILE_PATH=$(__get_lock_file_path)
rm -f "${LOCK_FILE_PATH}" # Ensure no lock file exists

lock

# Assert that the lock file is created
if [[ -f "${LOCK_FILE_PATH}" ]]; then
    logInfo "PASS: lock: lock file created"
else
    logErr "FAIL: lock: lock file not created"
    exit 1
fi

# Assert that the content of the lock file is the current process's PID
EXPECTED="$$"
ACTUAL=$(cat "${LOCK_FILE_PATH}")
assert_equals "${EXPECTED}" "${ACTUAL}" "lock: lock file content is current PID"

logInfo "All Task 10 tests completed."

# Task 11: Test lock (Lock Already Held by Current Process)
logInfo "Running Task 11: Test lock (Lock Already Held by Current Process)"

# Test Case 11.1: Lock already held by current process.
LOCK_FILE_PATH=$(__get_lock_file_path)

# Manually create a lock file with the current PID
mkdir -p "$(dirname "${LOCK_FILE_PATH}")"
echo "$$" > "${LOCK_FILE_PATH}"

# Call lock again
lock

# Assert that the lock file content remains unchanged
EXPECTED="$$"
ACTUAL=$(cat "${LOCK_FILE_PATH}")
assert_equals "${EXPECTED}" "${ACTUAL}" "lock: lock file content unchanged when already held"

# Assert that the function returns successfully (exit code 0)
# This is implicitly checked by set -e, if it didn't exit, it passed.
logInfo "PASS: lock: function returned successfully when already held"

logInfo "All Task 11 tests completed."

# Task 12: Test lock (Lock Held by Another Process - Running)
logInfo "Running Task 12: Test lock (Lock Held by Another Process - Running)"

# Test Case 12.1: Lock held by another running process.
LOCK_FILE_PATH=$(__get_lock_file_path)
rm -f "${LOCK_FILE_PATH}" # Ensure no lock file exists

# Start a dummy background process
sleep 60 & # This process will hold the lock
DUMMY_PID=$!

# Create a lock file with the dummy process's PID
mkdir -p "$(dirname "${LOCK_FILE_PATH}")"
echo "${DUMMY_PID}" > "${LOCK_FILE_PATH}"

# Call lock in a subshell and assert that it exits with status 1
assert_exit_code 1 "lock" "lock: exits with error when held by another running process"

# Kill the dummy process
kill "${DUMMY_PID}"

logInfo "All Task 12 tests completed."

# Task 13: Test lock (Lock Held by Another Process - Not Running)
logInfo "Running Task 13: Test lock (Lock Held by Another Process - Not Running)"

# Test Case 13.1: Lock held by a non-existent process.
LOCK_FILE_PATH=$(__get_lock_file_path)
rm -f "${LOCK_FILE_PATH}" # Ensure no lock file exists

# Create a lock file with a fake, non-existent PID
mkdir -p "$(dirname "${LOCK_FILE_PATH}")"
echo "99999" > "${LOCK_FILE_PATH}"

# Call lock
lock

# Assert that the lock file content is updated to the current process's PID
EXPECTED="$$"
ACTUAL=$(cat "${LOCK_FILE_PATH}")
assert_equals "${EXPECTED}" "${ACTUAL}" "lock: lock file content updated to current PID (stale lock)"

logInfo "All Task 13 tests completed."

