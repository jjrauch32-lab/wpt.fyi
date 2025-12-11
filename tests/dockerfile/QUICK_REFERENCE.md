# Dockerfile Test Suite - Quick Reference

## 🚀 Quick Commands

```bash
# Run all tests
./tests/dockerfile/run_all_tests.sh

# Shell tests only
./tests/dockerfile/validate_dockerfile.sh

# Go tests only
go test -tags=small ./tests/dockerfile/...
```

## 📊 Current Test Status

- **Total Tests**: 21 (9 shell + 12 Go)
- **Passing**: 11 ✅
- **Failing**: 10 ❌ (Expected - detects real issues)
- **Status**: Tests working correctly

## 🔍 What the Tests Check

### OS Compatibility
- ✅ Detects Alpine vs Debian base images
- ✅ Validates package manager usage (apt vs apk)
- ✅ Checks user management commands
- ✅ Identifies OS-specific package names

### Security
- ✅ Non-root user creation
- ✅ Package installation optimization
- ✅ No hardcoded secrets

### Dependencies
- ✅ Required tools (curl, wget, git)
- ✅ Project-specific (java, nodejs, python, gcloud)
- ✅ Environment variables (CLOUD_SDK_VERSION)

### Best Practices
- ✅ Layer count optimization
- ✅ Semantic versioning
- ✅ Comment quality
- ✅ Multi-line continuation

## ⚠️ Current Issues Detected

| Issue | Severity | Impact |
|-------|----------|--------|
| apt-get on Alpine | 🔴 Critical | Build fails |
| dpkg on Alpine | 🔴 Critical | Build fails |
| useradd on Alpine | 🔴 Critical | Build fails |
| Debian package names | 🔴 Critical | Wrong packages |
| /etc/apt/ references | 🔴 Critical | Path doesn't exist |

## 🔧 Quick Fix

```dockerfile
# Change line 2 from:
FROM golang:1.25.5-alpine3.21

# To:
FROM golang:1.25.5-bookworm
```

## 📖 Documentation

- `README.md` - Full usage guide
- `TEST_SUMMARY.md` - Complete test catalog
- `TEST_RESULTS.md` - Detailed results
- `INTEGRATION.md` - CI/CD setup
- `QUICK_REFERENCE.md` - This file

## 💡 Tips

- Run tests before committing Dockerfile changes
- Tests are tagged `small` for fast execution
- Exit code 0 = all tests pass
- Exit code 1 = issues detected
- Verbose output: add `-v` flag to Go tests

## 🎯 Test Integration

### Makefile
```make
dockerfile_test:
	./tests/dockerfile/validate_dockerfile.sh
	go test -tags=small ./tests/dockerfile/...
```

### Pre-commit Hook
```bash
#!/bin/bash
if git diff --cached --name-only | grep -q '^Dockerfile$'; then
    ./tests/dockerfile/validate_dockerfile.sh || exit 1
fi
```

## 📞 Support

For questions or issues with the tests:
1. Check `README.md` for detailed documentation
2. Review `TEST_RESULTS.md` for failure explanations
3. See `INTEGRATION.md` for CI/CD setup

---

**Remember**: These tests detect real compatibility issues. 
If tests fail, fix the Dockerfile before committing!