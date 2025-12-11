# Dockerfile Test Suite - CI/CD Integration Guide

## Integration with Existing Build System

### Makefile Integration

Add the following target to the project's `Makefile`:

```makefile
# Dockerfile validation tests
dockerfile_test:
	@echo "Running Dockerfile validation tests..."
	./tests/dockerfile/validate_dockerfile.sh
	go test -tags=small ./tests/dockerfile/...

# Add to existing test target
test: go_test python_test dockerfile_test
```

### GitHub Actions Integration

The tests are already integrated via the existing CI workflow which rebuilds 
the Docker image on Dockerfile changes. Add explicit validation:

```yaml
# In .github/workflows/ci.yml
dockerfile_validation:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Validate Dockerfile syntax and compatibility
      run: |
        ./tests/dockerfile/validate_dockerfile.sh
    - name: Run Dockerfile Go tests
      run: |
        go test -tags=small ./tests/dockerfile/...
```

### Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# Run Dockerfile tests before committing changes to Dockerfile

if git diff --cached --name-only | grep -q '^Dockerfile$'; then
    echo "Dockerfile changed, running validation tests..."
    ./tests/dockerfile/validate_dockerfile.sh
    if [ $? -ne 0 ]; then
        echo "❌ Dockerfile validation failed. Commit aborted."
        exit 1
    fi
fi
```

## Quick Commands

```bash
# Run all tests
./tests/dockerfile/run_all_tests.sh

# Run only shell tests
./tests/dockerfile/validate_dockerfile.sh

# Run only Go tests
go test -tags=small ./tests/dockerfile/...

# Run with verbose output
go test -tags=small -v ./tests/dockerfile/...
```

## Expected Behavior

### On Success
- All tests pass
- Exit code 0
- Green checkmarks in output

### On Failure
- Tests identify specific issues
- Exit code 1
- Red X marks and detailed error messages
- Guidance on fixing issues

## Current Status

⚠️ **The tests currently FAIL** because the Dockerfile has incompatible 
configuration (Alpine base with Debian commands). This is expected and 
demonstrates the tests are working correctly.

To fix, either:
1. Change base image back to Debian: `golang:1.25.5-bookworm`
2. Update all commands for Alpine compatibility

## Test Maintenance

- Update tests when changing base image OS
- Add tests for new dependencies
- Review security checks periodically
- Keep package name mappings current