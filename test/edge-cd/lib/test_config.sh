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
# Mock yq will be defined per-test
yq() {
    echo "[ERROR] yq mock not defined for this test" >&2
    return 1
}
export -f yq

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
    yq() {
        assert_equals "-e" "$1" "yq arg 1"
        assert_equals ".some.path" "$2" "yq arg 2"
        assert_equals "${CONFIG_PATH}" "$3" "yq arg 3"
        echo "config_value"
    }
    export -f yq
    export MY_VAR_EMPTY=""

    local result
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default")"
    assert_equals "config_value" "${result}" "Should read value from config file"
}

test_read_value_from_default_with_bug() {
    yq() { echo "null"; }
    export -f yq
    export MY_VAR_EMPTY=""

    local result
    # The function has a bug: it prints the default value but then exits with 1.
    # This test verifies the current buggy behavior.
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default_value" 2>/dev/null)" || {
        local status=$?
        assert_equals "default_value" "${result}" "Should print default value to stdout"
        assert_equals "1" "${status}" "Should exit 1 due to bug"
        return
    }
    echo "Test failed: read_value did not exit with an error as expected" >&2
    exit 1
}

test_read_value_precedence_env_over_config() {
    export MY_VAR="env_value"
    yq() { echo "config_value"; }
    export -f yq

    local result
    result="$(read_value "MY_VAR" ".some.path" "default")"
    assert_equals "env_value" "${result}" "Env var should have precedence over config"
    unset MY_VAR
}

test_read_value_precedence_config_over_default() {
    yq() { echo "config_value"; }
    export -f yq
    export MY_VAR_EMPTY=""

    local result
    result="$(read_value "MY_VAR_EMPTY" ".some.path" "default_value")"
    assert_equals "config_value" "${result}" "Config should have precedence over default"
}

test_read_value_no_value_fails() {
    yq() { return 1; }
    export -f yq
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
    yq() {
        # Simulate yq extracting envs from the config file
        echo 'VAR1=value1'
        echo 'VAR2=value2'
    }
    export -f yq

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
    yq() {
        echo 'EXISTING_VAR=new_value'
        echo 'NEW_VAR=a_value'
    }
    export -f yq

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
    yq() {
        assert_equals "-e" "$1" "yq arg 1"
        assert_equals ".a.b" "$2" "yq arg 2"
        # Simulate yq reading from stdin and processing
        local content
        content="$(cat)"
        if [[ "${content}" == "$(echo -e "a:\n  b: value")" ]]; then
            echo "value"
        else
            echo "unexpected stdin for yq mock" >&2
            return 1
        fi
    }
    export -f yq
    local result
    result="$(read_yaml_stdin ".a.b" "a:\n  b: value")"
    assert_equals "value" "${result}" "read_yaml_stdin should get correct value from yq"
}

test_read_yaml_file() {
    echo "a: {b: value}" >"${CONFIG_PATH}"
    yq() {
        assert_equals "-e" "$1" "yq arg 1"
        assert_equals ".a.b" "$2" "yq arg 2"
        assert_equals "${CONFIG_PATH}" "$3" "yq arg 3"
        echo "value"
    }
    export -f yq

    local result
    result="$(read_yaml_file ".a.b" "${CONFIG_PATH}")"
    assert_equals "value" "${result}" "read_yaml_file should call yq with correct arguments"
}

# --- Main ---
run_test test_read_value_from_env
run_test test_read_value_from_config
run_test test_read_value_from_default_with_bug
run_test test_read_value_precedence_env_over_config
run_test test_read_value_precedence_config_over_default
run_test test_read_value_no_value_fails
run_test test_set_and_reset_extra_envs
run_test test_set_and_reset_extra_envs_with_backup
run_test test_read_yaml_stdin
run_test test_read_yaml_file

report_results