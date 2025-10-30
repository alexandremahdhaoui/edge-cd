#!/bin/sh
#
# Copyright 2025 Alexandre Mahdhaoui
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -e
set -u

# --- Test Setup ---
TMP_DIR=$(mktemp -d)
# Copy the entire edge-cd directory to the temporary location
cp -R "$(cd "$(dirname "$0")/../../.." && pwd)/cmd/edge-cd" "${TMP_DIR}/edge-cd"

# Set up config path variables for testing
# CONFIG_PATH is the relative directory path within the config repo
export CONFIG_PATH="."
# CONFIG_REPO_DEST_PATH is where the config repo is located
export CONFIG_REPO_DEST_PATH="${TMP_DIR}/edge-cd"
# CONFIG_SPEC_FILE is the name of the spec file
export CONFIG_SPEC_FILE="config.yaml"
# The actual config file will be at: ${CONFIG_REPO_DEST_PATH}/${CONFIG_PATH}/${CONFIG_SPEC_FILE}
# Which resolves to: ${TMP_DIR}/edge-cd/./config.yaml

# Set default values that would normally come from edge-cd script
export __DEFAULT_CONFIG_REPO_DEST_PATH="/usr/local/src/edge-cd-config"
export __DEFAULT_CONFIG_SPEC_FILE="spec.yaml"

# Clean up on exit
trap 'rm -rf "${TMP_DIR}"' EXIT

# Source the library under test and its dependencies
. "${TMP_DIR}/edge-cd/lib/log.sh" # log.sh is a common dependency
. "${TMP_DIR}/edge-cd/lib/config.sh"

# --- Test Helpers ---
assert_equals() {
    expected="$1"
    actual="$2"
    message="$3"
    if [ "$expected" != "$actual" ]; then
        logErr "FAIL: ${message}"
        logErr "  Expected: '${expected}'"
        logErr "  Actual:   '${actual}'"
        return 1
    else
        logInfo "PASS: ${message}"
        return 0
    fi
}

assert_error() {
    command="$1"
    message="$2"
    set +e # Temporarily disable exit on error
    eval "$command"
    exit_code=$?
    set -e # Re-enable exit on error

    if [ "$exit_code" -ne 0 ]; then
        logInfo "PASS: ${message} - Command failed as expected with exit code ${exit_code}."
        return 0
    else
        logErr "FAIL: ${message} - Expected error, but command succeeded with exit code ${exit_code}."
        return 1
    fi
}

# --- Test Cases for read_yaml_stdin ---
logInfo "Running tests for read_yaml_stdin"

# Test 1.1: Read a simple string
YAML_CONTENT="key: value"
EXPECTED="value"
ACTUAL=$(read_yaml_stdin ".key" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin should read a simple string"

# Test 1.2: Read an integer
YAML_CONTENT="number: 123"
EXPECTED="123"
ACTUAL=$(read_yaml_stdin ".number" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin should read an integer"

# Test 1.3: Read a boolean
YAML_CONTENT="flag: true"
EXPECTED="true"
ACTUAL=$(read_yaml_stdin ".flag" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin should read a boolean"

# Test 1.4: Read a nested string
YAML_CONTENT=$(printf "parent:\n  child: nested_value")
EXPECTED="nested_value"
ACTUAL=$(read_yaml_stdin ".parent.child" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin should read a nested string"

# Test 1.5: Read a non-existent path (should fail)
YAML_CONTENT="key: value"
assert_error "read_yaml_stdin '.nonexistent' '${YAML_CONTENT}'" "read_yaml_stdin should fail for a non-existent path"

logInfo "All read_yaml_stdin tests completed."

# --- Test Cases for read_yaml_stdin_optional ---
logInfo "Running tests for read_yaml_stdin_optional"

# Test 2.1: Read an existing path
YAML_CONTENT="key: value"
EXPECTED="value"
ACTUAL=$(read_yaml_stdin_optional ".key" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin_optional should read an existing path"

# Test 2.2: Read a non-existent path (should return empty)
YAML_CONTENT="key: value"
EXPECTED=""
ACTUAL=$(read_yaml_stdin_optional ".nonexistent" "${YAML_CONTENT}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_stdin_optional should return empty for a non-existent path"

logInfo "All read_yaml_stdin_optional tests completed."

# --- Test Cases for read_yaml_file ---
logInfo "Running tests for read_yaml_file"

# Create a temporary YAML file for testing
TEST_YAML_FILE="${TMP_DIR}/test.yaml"
cat <<EOF > "${TEST_YAML_FILE}"
string_key: file_value
int_key: 456
bool_key: false
nested:
  file_child: nested_file_value
EOF

logInfo "Content of ${TEST_YAML_FILE}:"
cat "${TEST_YAML_FILE}"

# Test 3.1: Read a simple string from file
EXPECTED="file_value"
ACTUAL=$(read_yaml_file ".string_key" "${TEST_YAML_FILE}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_file should read a simple string from file"

# Test 3.2: Read an integer from file
EXPECTED="456"
ACTUAL=$(read_yaml_file ".int_key" "${TEST_YAML_FILE}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_file should read an integer from file"

# Test 3.3: Read a boolean from file (should fail)
assert_error "read_yaml_file \".bool_key\" \"${TEST_YAML_FILE}\"" "read_yaml_file should fail for a boolean key"

# Test 3.4: Read a nested string from file
EXPECTED="nested_file_value"
ACTUAL=$(read_yaml_file ".nested.file_child" "${TEST_YAML_FILE}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_file should read a nested string from file"

# Test 3.5: Read a non-existent path from file (should fail)
assert_error "read_yaml_file '.nonexistent' '${TEST_YAML_FILE}'" "read_yaml_file should fail for a non-existent path in file"

logInfo "All read_yaml_file tests completed."



# --- Test Cases for read_config ---

logInfo "Running tests for read_config"



# Create a temporary config.yaml for testing

cat <<EOF > "$(get_config_spec_abspath)"

app:

  name: edge-cd-app

  version: 1.0.0

EOF



# Test 4.1: Read a simple string from config

EXPECTED="edge-cd-app"

ACTUAL=$(read_config ".app.name")

assert_equals "${EXPECTED}" "${ACTUAL}" "read_config should read a simple string from config"



# Test 4.2: Read another string from config

EXPECTED="1.0.0"

ACTUAL=$(read_config ".app.version")

assert_equals "${EXPECTED}" "${ACTUAL}" "read_config should read another string from config"



# Test 4.3: Read a non-existent path from config (should fail)

assert_error "read_config '.nonexistent'" "read_config should fail for a non-existent path in config"



logInfo "All read_config tests completed."

# --- Test Cases for read_yaml_file_optional ---
logInfo "Running tests for read_yaml_file_optional"

# Create a temporary YAML file for testing
TEST_YAML_FILE_OPTIONAL="${TMP_DIR}/test_optional.yaml"
cat <<EOF > "${TEST_YAML_FILE_OPTIONAL}"
key_exists: value_optional
EOF

# Test 4.1.1: Read an existing path
EXPECTED="value_optional"
ACTUAL=$(read_yaml_file_optional ".key_exists" "${TEST_YAML_FILE_OPTIONAL}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_file_optional should read an existing path"

# Test 4.1.2: Read a non-existent path (should return "null")
EXPECTED="null"
ACTUAL=$(read_yaml_file_optional ".nonexistent" "${TEST_YAML_FILE_OPTIONAL}")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_yaml_file_optional should return \"null\" for a non-existent path"

logInfo "All read_yaml_file_optional tests completed."

# --- Test Cases for read_config_optional ---
logInfo "Running tests for read_config_optional"

# Create a temporary config.yaml for testing
cat <<EOF > "$(get_config_spec_abspath)"
optional_app:
  name: optional-edge-cd-app
EOF

# Test 4.2.1: Read an existing path
EXPECTED="optional-edge-cd-app"
ACTUAL=$(read_config_optional ".optional_app.name")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_config_optional should read an existing path"

# Test 4.2.2: Read a non-existent path (should return "null")
EXPECTED="null"
ACTUAL=$(read_config_optional ".nonexistent")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_config_optional should return \"null\" for a non-existent path"

logInfo "All read_config_optional tests completed."

# --- Test Cases for read_value ---
logInfo "Running tests for read_value"

# Test 5.1: Env Var set, Config set, Default provided.
export TEST_VAR_5_1="env_value_5_1"
cat <<EOF > "$(get_config_spec_abspath)"
key_5_1: config_value_5_1
EOF
EXPECTED="env_value_5_1"
ACTUAL=$(read_value TEST_VAR_5_1 ".key_5_1" "default_value_5_1")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_value should prioritize Env Var (5.1)"
unset TEST_VAR_5_1

# Test 5.2: Env Var unset, Config set, Default provided.
cat <<EOF > "$(get_config_spec_abspath)"
key_5_2: config_value_5_2
EOF
EXPECTED="config_value_5_2"
ACTUAL=$(read_value TEST_VAR_5_2 ".key_5_2" "default_value_5_2")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_value should prioritize Config (5.2)"

# Test 5.3: Env Var unset, Config unset, Default provided.
rm -f "$(get_config_spec_abspath)"
EXPECTED="default_value_5_3"
ACTUAL=$(read_value TEST_VAR_5_3 ".key_5_3" "default_value_5_3")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_value should use Default (5.3)"

# Test 5.4: Env Var unset, Config unset, Default unset (should exit 1).
rm -f "$(get_config_spec_abspath)"
# Run read_value in a subshell and capture its exit code
if (read_value TEST_VAR_5_4 ".key_5_4") 2>/dev/null; then
    logErr "FAIL: read_value should exit 1 if no value found (5.4) - Expected error, but command succeeded."
else
    exit_code=$?
    if [ "$exit_code" -ne 0 ]; then
        logInfo "PASS: read_value should exit 1 if no value found (5.4) - Command failed as expected with exit code ${exit_code}."
    else
        logErr "FAIL: read_value should exit 1 if no value found (5.4) - Expected error, but command succeeded with exit code ${exit_code}."
    fi
fi

# Test 5.5: Env Var "null", Config set, Default provided.
export TEST_VAR_5_5="null"
cat <<EOF > "$(get_config_spec_abspath)"
key_5_5: config_value_5_5
EOF
EXPECTED="config_value_5_5"
ACTUAL=$(read_value TEST_VAR_5_5 ".key_5_5" "default_value_5_5")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_value should treat Env Var \"null\" as unset (5.5)"
unset TEST_VAR_5_5

# Test 5.6: Env Var unset, Config "null", Default provided.
cat <<EOF > "$(get_config_spec_abspath)"
key_5_6: null
EOF
EXPECTED="default_value_5_6"
ACTUAL=$(read_value TEST_VAR_5_6 ".key_5_6" "default_value_5_6")
assert_equals "${EXPECTED}" "${ACTUAL}" "read_value should treat Config \"null\" as unset (5.6)"

logInfo "All read_value tests completed."

# Task 6: Test set_extra_envs and reset_extra_envs
logInfo "Running Task 6: Test set_extra_envs and reset_extra_envs"

# Test Case 6.1: Variable initially unset.
unset MY_VAR
cat <<EOF > "$(get_config_spec_abspath)"
extraEnvs:
  - MY_VAR: new_value
EOF
set_extra_envs
assert_equals "new_value" "${MY_VAR}" "set_extra_envs: unset variable becomes new_value"
reset_extra_envs
if [ -z "${MY_VAR+x}" ]; then
    logInfo "PASSED: reset_extra_envs: variable unset after reset"
else
    logErr "FAILED: reset_extra_envs: variable should be unset after reset"
    logErr "  Actual:   '${MY_VAR}'"
    exit 1
fi

# Test Case 6.2: Variable initially set.
export MY_VAR="original_value"
cat <<EOF > "$(get_config_spec_abspath)"
extraEnvs:
  - MY_VAR: new_value
EOF
set_extra_envs
assert_equals "new_value" "${MY_VAR}" "set_extra_envs: set variable becomes new_value"
reset_extra_envs
assert_equals "original_value" "${MY_VAR}" "reset_extra_envs: variable restored to original_value"

# Test Case 6.3: Multiple variables.
unset MY_VAR1
export MY_VAR2="original_value2"
cat <<EOF > "$(get_config_spec_abspath)"
extraEnvs:
  - MY_VAR1: new_value1
  - MY_VAR2: new_value2
EOF
set_extra_envs
assert_equals "new_value1" "${MY_VAR1}" "set_extra_envs: multiple variables - MY_VAR1"
assert_equals "new_value2" "${MY_VAR2}" "set_extra_envs: multiple variables - MY_VAR2"
reset_extra_envs
if [ -z "${MY_VAR1+x}" ]; then
    logInfo "PASSED: reset_extra_envs: multiple variables - MY_VAR1 unset after reset"
else
    logErr "FAILED: reset_extra_envs: multiple variables - MY_VAR1 should be unset after reset"
    logErr "  Actual:   '${MY_VAR1}'"
    exit 1
fi
# Direct check for MY_VAR2
if [ "${MY_VAR2}" = "original_value2" ]; then
    logInfo "PASSED: reset_extra_envs: multiple variables - MY_VAR2 restored"
else
    logErr "FAILED: reset_extra_envs: multiple variables - MY_VAR2 not restored"
    logErr "  Expected: 'original_value2'"
    logErr "  Actual:   '${MY_VAR2}'"
    exit 1
fi

logInfo "All Task 6 tests completed."

# --- Test Cases for log.sh ---

# Task 7: Test logInfo and logErr (Console Format)
logInfo "Running Task 7: Test logInfo and logErr (Console Format)"

# Set LOG_FORMAT to console
export LOG_FORMAT="console"

# Test Case 7.1: logInfo console output
EXPECTED="[INFO] Test info message"
ACTUAL=$( { logInfo "Test info message" 1>&2; } 2>&1 )
assert_equals "${EXPECTED}" "${ACTUAL}" "logInfo console output"

# Test Case 7.2: logErr console output
EXPECTED="[ERROR] Test error message"
ACTUAL=$( { logErr "Test error message" 1>&2; } 2>&1 )
assert_equals "${EXPECTED}" "${ACTUAL}" "logErr console output"

logInfo "All Task 7 tests completed."

# Task 8: Test logInfo and logErr (JSON Format)
logInfo "Running Task 8: Test logInfo and logErr (JSON Format)"

# Set LOG_FORMAT to json
export LOG_FORMAT="json"

# Test Case 8.1: logInfo JSON output
EXPECTED='{"level":"info","message":"Test info message"}'
ACTUAL=$( { logInfo "Test info message" 1>&2; } 2>&1 )
assert_equals "${EXPECTED}" "${ACTUAL}" "logInfo JSON output"

# Test Case 8.2: logErr JSON output
EXPECTED='{"level":"error","message":"Test error message"}'
ACTUAL=$( { logErr "Test error message" 1>&2; } 2>&1 )
assert_equals "${EXPECTED}" "${ACTUAL}" "logErr JSON output"

logInfo "All Task 8 tests completed."

logInfo "All config.sh unit tests completed successfully!"
