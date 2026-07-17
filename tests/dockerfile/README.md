# Dockerfile Test Suite

Comprehensive tests for validating the Dockerfile in the wpt.fyi project.

## Test Files

### `validate_dockerfile.sh`
Shell script that performs static analysis:
- Base image validity and versioning
- OS compatibility (Alpine vs Debian)
- Package manager compatibility
- User management commands
- Required dependencies
- Environment variables
- Security best practices

**Usage:**
```bash
./tests/dockerfile/validate_dockerfile.sh
```

### `dockerfile_test.go`
Go unit tests using testify:
- Tagged with `//go:build small`
- Run with: `go test -tags=small ./tests/dockerfile/...`

## Critical Issue Detected

**⚠️ IMPORTANT:** The Dockerfile change from `golang:1.25.3-bookworm` to 
`golang:1.25.5-alpine3.21` introduces compatibility issues:

Alpine base image but Debian commands:
- `apt-get` (should be `apk`)
- `dpkg` (not available in Alpine)
- `useradd` (should be `adduser`)
- `/etc/apt/` references (don't exist in Alpine)

**This will cause build failure.**

### Fix Options:
1. Revert to Debian: `golang:1.25.5-bookworm`
2. Update all commands for Alpine compatibility

## Test Coverage

- ✅ Dockerfile syntax and structure
- ✅ Base image validity
- ✅ OS compatibility detection
- ✅ Package manager compatibility
- ✅ User management commands
- ✅ Required dependencies
- ✅ Environment variables
- ✅ Security practices
- ✅ Project-specific dependencies

## Running Tests

```bash
# Shell tests
./tests/dockerfile/validate_dockerfile.sh

# Go tests
go test -tags=small ./tests/dockerfile/...

# Verbose
go test -tags=small -v ./tests/dockerfile/...
```

## Integration

Add to Makefile:
```make
dockerfile_test:
	./tests/dockerfile/validate_dockerfile.sh
```