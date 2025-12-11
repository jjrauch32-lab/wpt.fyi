#!/bin/bash
# Dockerfile Validation Test Suite
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0
TEST_OUTPUT=""

report_test() {
    local test_name="$1"
    local result="$2"
    local message="${3:-}"
    
    if [[ "$result" == "PASS" ]]; then
        echo -e "${GREEN}✓${NC} $test_name"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗${NC} $test_name"
        if [[ -n "$message" ]]; then
            echo -e "  ${RED}Error: $message${NC}"
        fi
        ((TESTS_FAILED++))
        TEST_OUTPUT="${TEST_OUTPUT}\nFAILED: $test_name - $message"
    fi
}

test_dockerfile_exists() {
    if [[ -f "Dockerfile" && -r "Dockerfile" ]]; then
        report_test "Dockerfile exists and is readable" "PASS"
    else
        report_test "Dockerfile exists and is readable" "FAIL" "Dockerfile not found"
    fi
}

test_base_image_syntax() {
    local from_line=$(grep -E '^\s*FROM\s+' Dockerfile | head -1)
    if [[ -n "$from_line" ]]; then
        report_test "FROM instruction is present" "PASS"
    else
        report_test "FROM instruction is present" "FAIL" "No FROM instruction"
    fi
}

test_base_image_validity() {
    local base_image=$(grep -E '^\s*FROM\s+' Dockerfile | head -1 | awk '{print $2}')
    
    if [[ -z "$base_image" ]]; then
        report_test "Base image is specified" "FAIL" "Base image not found"
        return
    fi
    
    if echo "$base_image" | grep -q '^golang:'; then
        report_test "Base image is a golang image" "PASS"
    else
        report_test "Base image is a golang image" "FAIL" "Expected golang, got: $base_image"
    fi
    
    if echo "$base_image" | grep -qE '^golang:[0-9]+\.[0-9]+\.[0-9]+'; then
        report_test "Base image has valid version format" "PASS"
    else
        report_test "Base image has valid version format" "FAIL" "Invalid format: $base_image"
    fi
}

test_os_compatibility() {
    local base_image=$(grep -E '^\s*FROM\s+' Dockerfile | head -1 | awk '{print $2}')
    local is_alpine=false
    local is_debian=false
    
    if echo "$base_image" | grep -qi 'alpine'; then
        is_alpine=true
        report_test "Detected Alpine-based image" "PASS"
    elif echo "$base_image" | grep -qiE '(bookworm|bullseye|debian)'; then
        is_debian=true
        report_test "Detected Debian-based image" "PASS"
    else
        report_test "Base image OS type detection" "FAIL" "Cannot determine OS: $base_image"
        return
    fi
    
    local has_apt=$(grep -c 'apt-get\|apt ' Dockerfile || true)
    local has_apk=$(grep -c 'apk ' Dockerfile || true)
    local has_dpkg=$(grep -c 'dpkg' Dockerfile || true)
    
    if $is_alpine; then
        if [[ $has_apt -gt 0 ]] || [[ $has_dpkg -gt 0 ]]; then
            report_test "Alpine uses correct package manager" "FAIL" \
                "Alpine cannot use apt ($has_apt) or dpkg ($has_dpkg)"
        elif [[ $has_apk -gt 0 ]]; then
            report_test "Alpine uses correct package manager" "PASS"
        else
            report_test "Alpine package manager" "FAIL" "Should use apk"
        fi
        
        if grep -q 'useradd' Dockerfile; then
            report_test "Alpine user management" "FAIL" "Should use adduser not useradd"
        elif grep -q 'adduser' Dockerfile; then
            report_test "Alpine user management" "PASS"
        fi
    elif $is_debian; then
        if [[ $has_apt -eq 0 ]]; then
            report_test "Debian uses apt" "FAIL" "Should use apt-get"
        else
            report_test "Debian uses apt" "PASS"
        fi
    fi
}

test_command_compatibility() {
    local base_image=$(grep -E '^\s*FROM\s+' Dockerfile | head -1 | awk '{print $2}')
    
    if echo "$base_image" | grep -qi 'alpine'; then
        local incompatible=()
        grep -q 'apt-get' Dockerfile && incompatible+=("apt-get") || true
        grep -q 'dpkg' Dockerfile && incompatible+=("dpkg") || true
        grep -q 'useradd' Dockerfile && incompatible+=("useradd") || true
        
        if [[ ${#incompatible[@]} -gt 0 ]]; then
            report_test "No Debian commands in Alpine" "FAIL" "Found: ${incompatible[*]}"
        else
            report_test "No Debian commands in Alpine" "PASS"
        fi
    fi
}

test_required_dependencies() {
    local missing=()
    for dep in curl wget git; do
        grep -q "$dep" Dockerfile || missing+=("$dep")
    done
    
    if [[ ${#missing[@]} -eq 0 ]]; then
        report_test "Required dependencies present" "PASS"
    else
        report_test "Required dependencies present" "FAIL" "Missing: ${missing[*]}"
    fi
}

test_environment_variables() {
    if grep -q 'ENV.*CLOUD_SDK_VERSION' Dockerfile; then
        report_test "CLOUD_SDK_VERSION is set" "PASS"
    else
        report_test "CLOUD_SDK_VERSION is set" "FAIL" "Not found"
    fi
}

test_security_practices() {
    if grep -qE '(USER|adduser|useradd).*browser' Dockerfile; then
        report_test "Non-root user created" "PASS"
    else
        report_test "Non-root user created" "FAIL" "No browser user"
    fi
}

test_project_dependencies() {
    local missing=()
    for dep in java nodejs python gcloud; do
        grep -qi "$dep" Dockerfile || missing+=("$dep")
    done
    
    if [[ ${#missing[@]} -eq 0 ]]; then
        report_test "Project dependencies present" "PASS"
    else
        report_test "Project dependencies present" "FAIL" "Missing: ${missing[*]}"
    fi
}

main() {
    echo "======================================"
    echo "  Dockerfile Validation Test Suite"
    echo "======================================"
    echo ""
    
    if [[ ! -f "Dockerfile" ]]; then
        if [[ -f "../../Dockerfile" ]]; then
            cd ../..
        else
            echo -e "${RED}Error: Cannot find Dockerfile${NC}"
            exit 1
        fi
    fi
    
    test_dockerfile_exists
    test_base_image_syntax
    test_base_image_validity
    test_os_compatibility
    test_command_compatibility
    test_required_dependencies
    test_environment_variables
    test_security_practices
    test_project_dependencies
    
    echo ""
    echo "======================================"
    echo "  Test Results Summary"
    echo "======================================"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"
    echo ""
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        echo -e "${RED}Some tests failed. Review output above.${NC}"
        exit 1
    else
        echo -e "${GREEN}All tests passed!${NC}"
        exit 0
    fi
}

main "$@"