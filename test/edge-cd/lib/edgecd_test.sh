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

# Create a test config directory and file
TEST_CONFIG_DIR="${TMP_DIR}/test-config"
mkdir -p "${TEST_CONFIG_DIR}"

# Create a minimal test config.yaml
cat > "${TEST_CONFIG_DIR}/config.yaml" <<'EOF'
config:
  spec: "config.yaml"
  repo:
    branch: "test-branch"
    destPath: "/tmp/test-config"
    url: "https://github.com/test/config.git"
  commitPath: "/tmp/test-config-commit.txt"

edgeCD:
  repo:
    branch: "test-edge-branch"
    destinationPath: "/tmp/test-edge-cd"
    url: "https://github.com/test/edge-cd.git"
  commitPath: "/tmp/test-edge-cd-commit.txt"

packageManager:
  name: "opkg"

pollingIntervalSecond: 60
EOF

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source dependencies
. "${TMP_DIR}/edge-cd/lib/config.sh"

# Source the library under test
. "${TMP_DIR}/edge-cd/lib/edgecd.sh"

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

assert_not_empty() {
    actual="$1"
    message="$2"
    if [ -n "$actual" ]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: (not empty)"
        logErr "  Actual:   (empty)"
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

logInfo "Starting edgecd.sh unit tests..."

# Test 1: Test declare_edgecd_config
logInfo "Running Test 1: Test declare_edgecd_config"

# Test Case 1.1: declare_edgecd_config runs without error
declare_edgecd_config
logInfo "PASS: declare_edgecd_config: runs without error"

logInfo "All Test 1 tests completed."

# Test 2: Test init_edgecd_config
logInfo "Running Test 2: Test init_edgecd_config"

# Test Case 2.1: init_edgecd_config fails when CONFIG_PATH not set
# Note: We need to unset in subshell to avoid set -u issues
assert_exit_code 1 "( unset CONFIG_PATH; unset CONFIG_SPEC_FILE; unset CONFIG_REPO_DEST_PATH; set +u; init_edgecd_config )" "init_edgecd_config: fails when CONFIG_PATH not set"

# Test Case 2.2: init_edgecd_config succeeds with CONFIG_PATH set
export CONFIG_PATH="."
export CONFIG_REPO_DEST_PATH="${TEST_CONFIG_DIR}"
export CONFIG_SPEC_FILE="config.yaml"

init_edgecd_config

# Verify backup variables are set
assert_not_empty "${__BACKUP_CONFIG_PATH}" "init_edgecd_config: sets __BACKUP_CONFIG_PATH"
assert_not_empty "${__BACKUP_CONFIG_SPEC_FILE}" "init_edgecd_config: sets __BACKUP_CONFIG_SPEC_FILE"
assert_not_empty "${__BACKUP_EDGE_CD_REPO_URL}" "init_edgecd_config: sets __BACKUP_EDGE_CD_REPO_URL"

logInfo "All Test 2 tests completed."

# Test 3: Test reset_edgecd_config
logInfo "Running Test 3: Test reset_edgecd_config"

# Test Case 3.1: reset_edgecd_config restores backup values
# First, modify some config variables
CONFIG_PATH="modified"
CONFIG_SPEC_FILE="modified.yaml"
CONFIG_REPO_BRANCH="modified-branch"

# Reset to backup
reset_edgecd_config

# Verify values are restored
assert_equals "." "${CONFIG_PATH}" "reset_edgecd_config: restores CONFIG_PATH"
assert_equals "config.yaml" "${CONFIG_SPEC_FILE}" "reset_edgecd_config: restores CONFIG_SPEC_FILE"
assert_equals "test-branch" "${CONFIG_REPO_BRANCH}" "reset_edgecd_config: restores CONFIG_REPO_BRANCH"

logInfo "All Test 3 tests completed."

# Test 4: Test configuration precedence (env > yaml > default)
logInfo "Running Test 4: Test configuration precedence"

# Test Case 4.1: Environment variable takes precedence
unset CONFIG_SPEC_FILE || true
export CONFIG_PATH="."
export CONFIG_REPO_DEST_PATH="${TEST_CONFIG_DIR}"
export CONFIG_SPEC_FILE="config.yaml"
export CONFIG_REPO_BRANCH="env-override-branch"

init_edgecd_config

assert_equals "env-override-branch" "${CONFIG_REPO_BRANCH}" "config precedence: environment variable overrides YAML"

# Test Case 4.2: YAML value is used when provided in config
assert_equals "/tmp/test-edge-cd-commit.txt" "${EDGE_CD_COMMIT_PATH}" "config precedence: YAML value used when provided"

# Test Case 4.3: Backup values are persisted correctly
# After init, backup values should match current values
assert_equals "${CONFIG_PATH}" "${__BACKUP_CONFIG_PATH}" "backup values: CONFIG_PATH backed up correctly"

logInfo "All Test 4 tests completed."

# Test 5: Verify library functions exist
logInfo "Running Test 5: Verify library functions exist"

# Test Case 5.1: Verify declare_edgecd_config function exists
if command -v declare_edgecd_config >/dev/null 2>&1; then
    logInfo "PASS: declare_edgecd_config function exists"
else
    logErr "FAIL: declare_edgecd_config function does not exist"
    exit 1
fi

# Test Case 5.2: Verify init_edgecd_config function exists
if command -v init_edgecd_config >/dev/null 2>&1; then
    logInfo "PASS: init_edgecd_config function exists"
else
    logErr "FAIL: init_edgecd_config function does not exist"
    exit 1
fi

# Test Case 5.3: Verify reset_edgecd_config function exists
if command -v reset_edgecd_config >/dev/null 2>&1; then
    logInfo "PASS: reset_edgecd_config function exists"
else
    logErr "FAIL: reset_edgecd_config function does not exist"
    exit 1
fi

logInfo "All Test 5 tests completed."

logInfo "All edgecd.sh unit tests completed successfully!"
