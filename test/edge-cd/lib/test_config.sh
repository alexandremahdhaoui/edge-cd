#!/usr/bin/env bash

# --- Preamble ---
# Test for cmd/edge-cd/lib/config.sh

set -o errexit
set -o nounset
set -o pipefail

# --- Setup ---
SRC_DIR_OF_THIS_SCRIPT="$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")"
LIB_DIR="${SRC_DIR_OF_THIS_SCRIPT}/../../../cmd/edge-cd/lib"
source "${LIB_DIR}/config.sh"

# --- Mocks ---
export CONFIG_PATH
CONFIG_PATH="$(mktemp)"
# yq is not mocked to test real yq command.

# --- Test Runner ---
TEST_COUNT=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    echo "--- Running test: ${test_name} ---"
    TEST_COUNT=$((TEST_COUNT + 1))
    ( # Run in a subshell to isolate environment changes and mocks
        set -o nounset # Ensure nounset is active in subshell
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
        # Cleanup
        rm -f "${CONFIG_PATH}"
        exit 0
    else
        echo "Failed tests: ${#FAILED_TESTS[@]}"
        for failed_test in "${FAILED_TESTS[@]}"; do
            echo "  - ${failed_test}"
        done
        # Cleanup
        rm -f "${CONFIG_PATH}"
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

test_read_value_from_env() {
    export MY_VAR="env_value"
    local result
    result="$(read_value "MY_VAR" ".some.path" "default")"
    assert_equals "env_value" "${result}" "Should read value from environment variable"
    unset MY_VAR
}

test_read_value_from_config() {
    echo "some: {path: config_value}" > "${CONFIG_PATH}"
    export MY_VAR_EMPTY=""

    local result
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default")"
    assert_equals "config_value" "${result}" "Should read value from config file"
}

test_read_value_from_default() {
    echo "some: {path: null}" > "${CONFIG_PATH}"
    export MY_VAR_EMPTY=""

    local result
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default_value")"
    assert_equals "default_value" "${result}" "Should use default value when config value is null"
}

test_read_value_precedence_env_over_config() {
    export MY_VAR="env_value"
    echo "some: {path: config_value}" > "${CONFIG_PATH}"

    local result
    result="$(read_value "MY_VAR" ".some.path" "default")"
    assert_equals "env_value" "${result}" "Env var should have precedence over config"
    unset MY_VAR
}

test_read_value_precedence_config_over_default() {
    echo "some: {path: config_value}" > "${CONFIG_PATH}"
    export MY_VAR_EMPTY=""

    local result
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default_value")"
    assert_equals "config_value" "${result}" "Config should have precedence over default"
}

test_read_value_no_value_fails() {
    >"${CONFIG_PATH}" # Create an empty file
    export MY_VAR_EMPTY=""

    local status=0
    # Expect the function to fail (exit 1)
    (read_value "MY_VAR_EMPTY" ".non.existent.path" "" &>/dev/null) || status=$?
    assert_equals "1" "${status}" "Should exit with status 1 when no value is found"
}

test_set_and_reset_extra_envs() {
    # Setup
    cat >"${CONFIG_PATH}" <<EOF
extraEnvs:
  - VAR1: "value1"
  - VAR2: "value2"
EOF

    # Test set_extra_envs
    set_extra_envs
    assert_equals "value1" "${VAR1}" "VAR1 should be set"
    assert_equals "value2" "${VAR2}" "VAR2 should be set"
    assert_equals "VAR1 VAR2" "$(echo "${BACKUP_EXTRA_ENVS[*]}")" "BACKUP_EXTRA_ENVS should contain var names"

    # Test reset_extra_envs
    reset_extra_envs
    assert_equals "" "${VAR1:-}" "VAR1 should be unset"
    assert_equals "" "${VAR2:-}" "VAR2 should be unset"
}

test_set_and_reset_extra_envs_with_backup() {
    # Setup
    export EXISTING_VAR="original_value"
    cat >"${CONFIG_PATH}" <<EOF
extraEnvs:
  - EXISTING_VAR: "new_value"
  - NEW_VAR: "a_value"
EOF

    # Test set_extra_envs
    set_extra_envs
    assert_equals "new_value" "${EXISTING_VAR}" "EXISTING_VAR should be updated"
    assert_equals "a_value" "${NEW_VAR}" "NEW_VAR should be set"

    # Test reset_extra_envs
    reset_extra_envs
    assert_equals "original_value" "${EXISTING_VAR}" "EXISTING_VAR should be restored"
    assert_equals "" "${NEW_VAR:-}" "NEW_VAR should be unset"
    unset EXISTING_VAR
}

test_read_yaml_stdin() {
    local result
    result="$(read_yaml_stdin ".a.b" "a:\n  b: value")"
    assert_equals "value" "${result}" "read_yaml_stdin should get correct value from yq"
}

test_read_yaml_file() {
    echo "a: {b: value}" >"${CONFIG_PATH}"

    local result
    result="$(read_yaml_file ".a.b" "${CONFIG_PATH}")"
    assert_equals "value" "${result}" "read_yaml_file should call yq with correct arguments"
}

test_set_extra_envs_no_envs_in_config() {
    # Setup
    cat >"${CONFIG_PATH}" <<EOF
some_other_key: value
EOF
# yq is not mocked, so no need to unset.

    # Test set_extra_envs
    set_extra_envs
    assert_equals "" "$(echo "${BACKUP_EXTRA_ENVS[*]}")" "BACKUP_EXTRA_ENVS should be empty"
}

# --- Main ---
run_test test_read_value_from_env
run_test test_read_value_from_config
run_test test_read_value_from_default
run_test test_read_value_precedence_env_over_config
run_test test_read_value_precedence_config_over_default
run_test test_read_value_no_value_fails
run_test test_set_and_reset_extra_envs
run_test test_set_and_reset_extra_envs_with_backup
run_test test_read_yaml_stdin
run_test test_read_yaml_file
run_test test_set_extra_envs_no_envs_in_config

report_results