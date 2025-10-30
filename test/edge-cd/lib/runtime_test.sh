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
. "${TMP_DIR}/edge-cd/lib/runtime.sh"

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

assert_empty() {
    actual="$1"
    message="$2"
    if [ -z "$actual" ]; then
        logInfo "PASS: ${message}"
    else
        logErr "FAIL: ${message}"
        logErr "  Expected: (empty)"
        logErr "  Actual:   '$actual'"
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

logInfo "Starting runtime.sh unit tests..."

# Test 1: Test declare_runtime_variables
logInfo "Running Test 1: Test declare_runtime_variables"

# Test Case 1.1: declare_runtime_variables runs without error
declare_runtime_variables
logInfo "PASS: declare_runtime_variables: runs without error"

logInfo "All Test 1 tests completed."

# Test 2: Test reset_runtime_variables
logInfo "Running Test 2: Test reset_runtime_variables"

# Test Case 2.1: reset_runtime_variables clears all state
RTV_REQUIRE_SERVICE_RESTART="service1
service2
"
RTV_REQUIRE_SELF_RESTART=true
RTV_REQUIRE_REBOOT=true

reset_runtime_variables

assert_empty "${RTV_REQUIRE_SERVICE_RESTART}" "reset_runtime_variables: RTV_REQUIRE_SERVICE_RESTART is empty"
assert_equals "false" "${RTV_REQUIRE_SELF_RESTART}" "reset_runtime_variables: RTV_REQUIRE_SELF_RESTART is false"
assert_equals "false" "${RTV_REQUIRE_REBOOT}" "reset_runtime_variables: RTV_REQUIRE_REBOOT is false"

logInfo "All Test 2 tests completed."

# Test 3: Test add_service_for_restart
logInfo "Running Test 3: Test add_service_for_restart"

# Test Case 3.1: add_service_for_restart adds service correctly
reset_runtime_variables
add_service_for_restart "nginx"

assert_contains "${RTV_REQUIRE_SERVICE_RESTART}" "nginx" "add_service_for_restart: service 'nginx' added"

# Test Case 3.2: add_service_for_restart adds multiple services
add_service_for_restart "apache2"

assert_contains "${RTV_REQUIRE_SERVICE_RESTART}" "nginx" "add_service_for_restart: service 'nginx' still present"
assert_contains "${RTV_REQUIRE_SERVICE_RESTART}" "apache2" "add_service_for_restart: service 'apache2' added"

# Test Case 3.3: add_service_for_restart returns 1 for empty service name
assert_exit_code 1 "add_service_for_restart ''" "add_service_for_restart: returns 1 for empty service name"

logInfo "All Test 3 tests completed."

# Test 4: Test add_service_for_restart handles duplicates
logInfo "Running Test 4: Test add_service_for_restart handles duplicates"

# Test Case 4.1: add_service_for_restart allows duplicates (they will be filtered by get_services_to_restart)
reset_runtime_variables
add_service_for_restart "nginx"
add_service_for_restart "nginx"

# Both duplicates should be in the raw list
NGINX_COUNT=$(echo "${RTV_REQUIRE_SERVICE_RESTART}" | grep -c "nginx" || true)
assert_equals "2" "${NGINX_COUNT}" "add_service_for_restart: duplicates are added to raw list"

logInfo "All Test 4 tests completed."

# Test 5: Test get_services_to_restart
logInfo "Running Test 5: Test get_services_to_restart"

# Test Case 5.1: get_services_to_restart returns sorted unique list
reset_runtime_variables
add_service_for_restart "nginx"
add_service_for_restart "apache2"
add_service_for_restart "nginx"  # duplicate
add_service_for_restart "mysql"

SERVICES=$(get_services_to_restart)

# Check that all services are present
assert_contains "${SERVICES}" "nginx" "get_services_to_restart: contains 'nginx'"
assert_contains "${SERVICES}" "apache2" "get_services_to_restart: contains 'apache2'"
assert_contains "${SERVICES}" "mysql" "get_services_to_restart: contains 'mysql'"

# Check that services are unique (only 3 lines, not 4)
LINE_COUNT=$(echo "${SERVICES}" | wc -l)
assert_equals "3" "${LINE_COUNT}" "get_services_to_restart: returns unique services (3 services, not 4)"

# Check that services are sorted
EXPECTED_ORDER="apache2
mysql
nginx"
assert_equals "${EXPECTED_ORDER}" "${SERVICES}" "get_services_to_restart: returns sorted services"

# Test Case 5.2: get_services_to_restart returns empty for no services
reset_runtime_variables
SERVICES=$(get_services_to_restart)
assert_empty "${SERVICES}" "get_services_to_restart: returns empty when no services"

logInfo "All Test 5 tests completed."

# Test 6: Test RTV_REQUIRE_REBOOT flag
logInfo "Running Test 6: Test RTV_REQUIRE_REBOOT flag"

# Test Case 6.1: RTV_REQUIRE_REBOOT can be set and reset
reset_runtime_variables
assert_equals "false" "${RTV_REQUIRE_REBOOT}" "RTV_REQUIRE_REBOOT: initially false"

RTV_REQUIRE_REBOOT=true
assert_equals "true" "${RTV_REQUIRE_REBOOT}" "RTV_REQUIRE_REBOOT: can be set to true"

reset_runtime_variables
assert_equals "false" "${RTV_REQUIRE_REBOOT}" "RTV_REQUIRE_REBOOT: reset to false"

logInfo "All Test 6 tests completed."

logInfo "All runtime.sh unit tests completed successfully!"
