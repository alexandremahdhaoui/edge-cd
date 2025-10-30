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

# Source dependencies
. "${TMP_DIR}/edge-cd/lib/runtime.sh"
. "${TMP_DIR}/edge-cd/lib/config.sh"

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/files.sh"

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

assert_file_exists() {
    file_path="$1"
    message="$2"
    if [ -f "$file_path" ]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  File does not exist: $file_path"
        exit 1
    fi
}

assert_file_content() {
    file_path="$1"
    expected_content="$2"
    message="$3"
    if [ -f "$file_path" ]; then
        actual_content=$(cat "$file_path")
        if [ "$expected_content" = "$actual_content" ]; then
            logInfo "PASS: ${message}"
        else
            logErr "FAIL: ${message}"
            logErr "  Expected content: '$expected_content'"
            logErr "  Actual content:   '$actual_content'"
            exit 1
        fi
    else
        logErr "FAIL: ${message} (file does not exist)"
        exit 1
    fi
}

assert_file_perms() {
    file_path="$1"
    expected_perms="$2"
    message="$3"
    if [ -f "$file_path" ]; then
        # Get file permissions in octal format (e.g., 644)
        actual_perms=$(stat -c '%a' "$file_path" 2>/dev/null || stat -f '%Lp' "$file_path" 2>/dev/null)
        if [ "$expected_perms" = "$actual_perms" ]; then
            logInfo "PASS: ${message}"
        else
            logErr "FAIL: ${message}"
            logErr "  Expected perms: '$expected_perms'"
            logErr "  Actual perms:   '$actual_perms'"
            exit 1
        fi
    else
        logErr "FAIL: ${message} (file does not exist)"
        exit 1
    fi
}

assert_contains() {
    haystack="$1"
    needle="$2"
    message="$3"
    if echo "$haystack" | grep -q "$needle"; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Haystack: '$haystack'"
        logErr "  Needle:   '$needle'"
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

logInfo "Starting files.sh unit tests..."

# Initialize runtime variables for testing
declare_runtime_variables
reset_runtime_variables

# Test 1: Test reconcile_file creates new files
logInfo "Running Test 1: Test reconcile_file creates new files"

# Test Case 1.1: reconcile_file creates a new file
SRC_FILE="${TMP_DIR}/source.txt"
DEST_FILE="${TMP_DIR}/dest.txt"
echo "test content" > "${SRC_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" ""

assert_file_exists "${DEST_FILE}" "reconcile_file: creates new file"
assert_file_content "${DEST_FILE}" "test content" "reconcile_file: file has correct content"

logInfo "All Test 1 tests completed."

# Test 2: Test reconcile_file detects no drift for identical files
logInfo "Running Test 2: Test reconcile_file detects no drift"

# Test Case 2.1: reconcile_file returns 0 for identical files
SRC_FILE="${TMP_DIR}/source2.txt"
DEST_FILE="${TMP_DIR}/dest2.txt"
echo "identical content" > "${SRC_FILE}"
echo "identical content" > "${DEST_FILE}"

# Store mtime before reconcile
MTIME_BEFORE=$(stat -c '%Y' "${DEST_FILE}" 2>/dev/null || stat -f '%m' "${DEST_FILE}" 2>/dev/null)
sleep 1

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" ""

# Check that file was not modified (mtime unchanged)
MTIME_AFTER=$(stat -c '%Y' "${DEST_FILE}" 2>/dev/null || stat -f '%m' "${DEST_FILE}" 2>/dev/null)
assert_equals "${MTIME_BEFORE}" "${MTIME_AFTER}" "reconcile_file: does not modify identical files"

logInfo "All Test 2 tests completed."

# Test 3: Test reconcile_file updates files with drift
logInfo "Running Test 3: Test reconcile_file updates files with drift"

# Test Case 3.1: reconcile_file updates files with different content
SRC_FILE="${TMP_DIR}/source3.txt"
DEST_FILE="${TMP_DIR}/dest3.txt"
echo "new content" > "${SRC_FILE}"
echo "old content" > "${DEST_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" ""

assert_file_content "${DEST_FILE}" "new content" "reconcile_file: updates file with drift"

logInfo "All Test 3 tests completed."

# Test 4: Test reconcile_file sets correct file permissions
logInfo "Running Test 4: Test reconcile_file sets correct file permissions"

# Test Case 4.1: reconcile_file sets 644 permissions
SRC_FILE="${TMP_DIR}/source4.txt"
DEST_FILE="${TMP_DIR}/dest4.txt"
echo "content" > "${SRC_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" ""

assert_file_perms "${DEST_FILE}" "644" "reconcile_file: sets 644 permissions"

# Test Case 4.2: reconcile_file sets 755 permissions
SRC_FILE="${TMP_DIR}/source5.txt"
DEST_FILE="${TMP_DIR}/dest5.txt"
echo "content" > "${SRC_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "755" "" ""

assert_file_perms "${DEST_FILE}" "755" "reconcile_file: sets 755 permissions"

logInfo "All Test 4 tests completed."

# Test 5: Test reconcile_file marks services for restart
logInfo "Running Test 5: Test reconcile_file marks services for restart"

# Test Case 5.1: reconcile_file marks service for restart when file changes
reset_runtime_variables
SRC_FILE="${TMP_DIR}/source6.txt"
DEST_FILE="${TMP_DIR}/dest6.txt"
echo "new content" > "${SRC_FILE}"
echo "old content" > "${DEST_FILE}"

# Mock yq to return a service name
# Note: In real scenario, restartServices would be a YAML array
# For testing, we'll pass a simple string that yq would process
RESTART_SERVICES_JSON='["nginx"]'

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "${RESTART_SERVICES_JSON}" ""

SERVICES=$(get_services_to_restart)
assert_contains "${SERVICES}" "nginx" "reconcile_file: marks service 'nginx' for restart"

logInfo "All Test 5 tests completed."

# Test 6: Test reconcile_file sets reboot flag
logInfo "Running Test 6: Test reconcile_file sets reboot flag"

# Test Case 6.1: reconcile_file sets reboot flag when requireReboot is true
reset_runtime_variables
SRC_FILE="${TMP_DIR}/source7.txt"
DEST_FILE="${TMP_DIR}/dest7.txt"
echo "new content" > "${SRC_FILE}"
echo "old content" > "${DEST_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" "true"

assert_equals "true" "${RTV_REQUIRE_REBOOT}" "reconcile_file: sets reboot flag to true"

# Test Case 6.2: reconcile_file does not set reboot flag when requireReboot is false
reset_runtime_variables
SRC_FILE="${TMP_DIR}/source8.txt"
DEST_FILE="${TMP_DIR}/dest8.txt"
echo "new content" > "${SRC_FILE}"
echo "old content" > "${DEST_FILE}"

reconcile_file "${SRC_FILE}" "${DEST_FILE}" "644" "" "false"

assert_equals "false" "${RTV_REQUIRE_REBOOT}" "reconcile_file: does not set reboot flag when false"

logInfo "All Test 6 tests completed."

# Test 7: Test reconcile_file validates inputs
logInfo "Running Test 7: Test reconcile_file validates inputs"

# Test Case 7.1: reconcile_file exits with error for empty srcPath
assert_exit_code 1 "reconcile_file '' '/tmp/dest' '644' '' ''" "reconcile_file: exits with error for empty srcPath"

# Test Case 7.2: reconcile_file exits with error for null srcPath
assert_exit_code 1 "reconcile_file 'null' '/tmp/dest' '644' '' ''" "reconcile_file: exits with error for null srcPath"

# Test Case 7.3: reconcile_file exits with error for empty destPath
assert_exit_code 1 "reconcile_file '/tmp/src' '' '644' '' ''" "reconcile_file: exits with error for empty destPath"

# Test Case 7.4: reconcile_file exits with error for relative destPath
assert_exit_code 1 "reconcile_file '/tmp/src' 'relative/path' '644' '' ''" "reconcile_file: exits with error for relative destPath"

logInfo "All Test 7 tests completed."

logInfo "All files.sh unit tests completed successfully!"
logInfo "Note: Full YAML-based file reconciliation is tested in E2E tests"
