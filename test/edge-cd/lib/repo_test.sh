#!/bin/sh
set -e
set -u

# Source the log.sh for logging functions
. "$(dirname "$0")/../../../cmd/edge-cd/lib/log.sh"

# Define SRC_DIR relative to the test script
SRC_DIR="$(dirname "$0")/../../.."

# General Test Setup
# Create a temporary directory and copy project files
TMP_DIR=$(mktemp -d)
cp -R "${SRC_DIR}/cmd/edge-cd" "${TMP_DIR}/edge-cd"

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/repo.sh"

# Helper function for assertions
assert_equals() {
    expected="$1"
    actual="$2"
    message="$3"
    if [ "$expected" = "$actual" ]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: '${expected}'"
        logErr "  Actual:   '${actual}'"
        exit 1
    fi
}

assert_exit_code() {
    expected_code="$1"
    command="$2"
    message="$3"
    set +e # Disable exit on error for this check
    ( eval "$command" )
    actual_code=$?
    set -e # Re-enable exit on error
    if [ "$expected_code" -eq "$actual_code" ]; then
        logInfo "PASS: ${message} (Exit code: $actual_code)"
    else
        logErr "FAIL: ${message} (Expected exit code: $expected_code, Actual: $actual_code)"
        exit 1
    fi
}

logInfo "Starting repo.sh unit tests..."

# Test 1: Test is_file_url with file:// URLs
logInfo "Running Test 1: Test is_file_url with file:// URLs"

# Test Case 1.1: is_file_url returns 0 for file:// URL
assert_exit_code 0 "is_file_url 'file:///opt/config'" "is_file_url: returns 0 for file:///opt/config"

# Test Case 1.2: is_file_url returns 0 for file:// URL without trailing slash
assert_exit_code 0 "is_file_url 'file://localhost/path/to/repo'" "is_file_url: returns 0 for file://localhost/path/to/repo"

logInfo "All Test 1 tests completed."

# Test 2: Test is_file_url with non-file URLs
logInfo "Running Test 2: Test is_file_url with non-file URLs"

# Test Case 2.1: is_file_url returns 1 for https:// URL
assert_exit_code 1 "is_file_url 'https://github.com/user/repo.git'" "is_file_url: returns 1 for https:// URL"

# Test Case 2.2: is_file_url returns 1 for http:// URL
assert_exit_code 1 "is_file_url 'http://example.com/repo.git'" "is_file_url: returns 1 for http:// URL"

# Test Case 2.3: is_file_url returns 1 for git:// URL
assert_exit_code 1 "is_file_url 'git://github.com/user/repo.git'" "is_file_url: returns 1 for git:// URL"

# Test Case 2.4: is_file_url returns 1 for ssh:// URL
assert_exit_code 1 "is_file_url 'ssh://git@github.com/user/repo.git'" "is_file_url: returns 1 for ssh:// URL"

# Test Case 2.5: is_file_url returns 1 for relative path
assert_exit_code 1 "is_file_url '/absolute/path/to/repo'" "is_file_url: returns 1 for absolute path"

logInfo "All Test 2 tests completed."

# Test 3: Verify library functions exist
logInfo "Running Test 3: Verify library functions exist"

# Test Case 3.1: Verify clone_repo_edge_cd function exists
if command -v clone_repo_edge_cd >/dev/null 2>&1; then
    logInfo "PASS: clone_repo_edge_cd function exists"
else
    logErr "FAIL: clone_repo_edge_cd function does not exist"
    exit 1
fi

# Test Case 3.2: Verify sync_repo_edge_cd function exists
if command -v sync_repo_edge_cd >/dev/null 2>&1; then
    logInfo "PASS: sync_repo_edge_cd function exists"
else
    logErr "FAIL: sync_repo_edge_cd function does not exist"
    exit 1
fi

# Test Case 3.3: Verify clone_repo_config function exists
if command -v clone_repo_config >/dev/null 2>&1; then
    logInfo "PASS: clone_repo_config function exists"
else
    logErr "FAIL: clone_repo_config function does not exist"
    exit 1
fi

# Test Case 3.4: Verify sync_repo_config function exists
if command -v sync_repo_config >/dev/null 2>&1; then
    logInfo "PASS: sync_repo_config function exists"
else
    logErr "FAIL: sync_repo_config function does not exist"
    exit 1
fi

logInfo "All Test 3 tests completed."

logInfo "All repo.sh unit tests completed successfully!"
logInfo "Note: Full git operations are tested in E2E tests"
