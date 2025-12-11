# Dockerfile Test Results

## Test Execution Summary
- **Date**: 2025-12-11 06:05:34 UTC
- **Repository**: web-platform-tests/wpt.fyi
- **Branch**: Current branch
- **Changed File**: Dockerfile
- **Change**: `golang:1.25.3-bookworm` → `golang:1.25.5-alpine3.21`

## Test Suite Composition

### Shell Tests (validate_dockerfile.sh)
- **Total Tests**: 9
- **Test Type**: Static analysis and validation
- **Language**: Bash

### Go Tests (dockerfile_test.go)
- **Total Tests**: 12
- **Test Type**: Unit tests with assertions
- **Language**: Go
- **Framework**: testify/assert
- **Build Tag**: `small`

## Test Results Summary

### ✅ Passing Tests (11/21)

1. **TestDockerfileExists** - File exists and is readable
2. **TestBaseImageFormat** - FROM instruction is valid
3. **TestBaseImageVersion** - Semantic versioning is correct (1.25.5)
4. **TestOSTypeDetection** - Alpine OS detected correctly
5. **TestNonRootUserCreation** - Browser user is created
6. **TestRequiredDependencies** - curl, wget, git present
7. **TestProjectSpecificDependencies** - java, nodejs, python, gcloud referenced
8. **TestCloudSDKVersion** - ENV CLOUD_SDK_VERSION=527.0.0 set correctly
9. **TestLayerCount** - Reasonable number of layers (4 RUN commands)
10. **test_dockerfile_exists** (shell) - File validation passed
11. **test_base_image_syntax** (shell) - FROM instruction validated

### ❌ Failing Tests (10/21)

#### 1. TestPackageManagerCompatibility
**Issue**: Alpine image uses Debian package manager commands

**Failures**:
- ❌ Found `apt-get` commands (should use `apk`)
- ❌ Found `dpkg` commands (not available on Alpine)
- ❌ No `apk` commands found

**Impact**: Build will fail with "apt-get: command not found"

**Lines Affected**: 9-17, 23-34, 48-53

**Fix**: Replace all `apt-get update && apt-get install` with `apk add --no-cache`

#### 2. TestUserManagementCommands
**Issue**: Alpine image uses Debian user management command

**Failure**:
- ❌ Found `useradd` (should use `adduser` on Alpine)

**Impact**: Build will fail with "useradd: command not found"

**Line Affected**: 6

**Fix**: Replace `useradd --uid 9999 --user-group --create-home browser` with `adduser -D -u 9999 browser`

#### 3. TestDebianPackagesOnAlpine
**Issue**: Debian-specific package names referenced

**Failures**:
- ❌ `python3-crcmod` (Debian package name)
- ❌ `firefox-esr` (Debian package name)
- ❌ `tox` (Different packaging on Alpine)

**Impact**: Packages won't be found or will install incorrectly

**Lines Affected**: 23-34

**Fix**: Use Alpine equivalents: `py3-crcmod`, `firefox`, `py3-tox`

#### 4-10. Additional Shell Test Failures
Multiple shell validation tests failed for the same compatibility reasons.

## Detailed Failure Analysis

### Critical Issue #1: Package Manager Incompatibility

**Current (Lines 23-34):**
```dockerfile
RUN apt-get update -qqy && apt-get install -qqy --no-install-suggests \
        curl \
        firefox-esr \
        java-11-amazon-corretto-jdk \
        nodejs \
        python3.11 \
        python3-crcmod \
        sudo \
        tox \
        wget \
        xvfb && \
    rm /usr/bin/firefox
```

**Issue**: `apt-get` doesn't exist on Alpine Linux

**Should be (for Alpine):**
```dockerfile
RUN apk add --no-cache \
        curl \
        firefox \
        openjdk11 \
        nodejs npm \
        python3 \
        py3-pip \
        sudo \
        wget \
        xvfb && \
    pip3 install crcmod tox
```

### Critical Issue #2: User Creation Command

**Current (Line 6):**
```dockerfile
RUN chmod a+rx $HOME && useradd --uid 9999 --user-group --create-home browser
```

**Issue**: `useradd` doesn't exist on Alpine Linux (uses BusyBox utilities)

**Should be:**
```dockerfile
RUN chmod a+rx $HOME && adduser -D -u 9999 -h /home/browser browser
```

### Critical Issue #3: APT Repository Management

**Current (Lines 9-17):**
```dockerfile
RUN export DISTRO_CODENAME=$(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) && \
    echo "deb [signed-by=/usr/share/keyrings/corretto.gpg] https://apt.corretto.aws stable main" > /etc/apt/sources.list.d/corretto.list && \
    ...
```

**Issue**: 
- `/etc/apt/sources.list.d/` doesn't exist on Alpine
- `.deb` repositories not compatible with Alpine's `apk`
- `dpkg` command not available
- `VERSION_CODENAME` may not exist in Alpine's `/etc/os-release`

**Should be**: Complete redesign needed for Alpine package sources

## Package Name Mappings

| Debian Package | Alpine Equivalent | Notes |
|----------------|-------------------|-------|
| python3-crcmod | py3-crcmod or pip install | May need py3-pip |
| firefox-esr | firefox | ESR is default on Alpine |
| tox | py3-tox or pip install | Available in community repo |
| java-11-amazon-corretto-jdk | openjdk11 | Alpine uses OpenJDK |
| nodejs (from NodeSource) | nodejs npm | Available in main repo |
| python3.11 | python3 | Alpine provides latest stable |

## Risk Assessment

### Severity: 🔴 CRITICAL - BUILD BREAKING

**Current State**: The Dockerfile **WILL NOT BUILD** with Alpine base image.

**Expected Build Failures**:
1. **Line 6**: `useradd: not found`
2. **Lines 9-17**: `/etc/apt/sources.list.d/: No such file or directory`
3. **Line 13**: `dpkg: not found`
4. **Line 23**: `apt-get: not found`
5. **Line 48**: `apt-get: not found`

### Impact Assessment:
- ❌ Docker image build fails immediately
- ❌ CI/CD pipeline completely broken
- ❌ Local development environment unusable
- ❌ All dependent services unable to start
- ❌ Staging/production deployments blocked

### Timeline to Failure:
- **Immediate**: First `docker build` command will fail
- **CI Impact**: Next commit/PR will show red build status
- **User Impact**: Developers cannot start local environment

## Recommendations

### Option 1: Revert Base Image ⭐ RECOMMENDED

**Change:**
```dockerfile
FROM golang:1.25.5-bookworm
```

**Advantages**:
- ✅ Maintains all existing functionality
- ✅ Zero additional code changes needed
- ✅ Immediate fix
- ✅ No testing required beyond existing tests
- ✅ No risk

**Time**: 1 minute to implement

### Option 2: Full Alpine Migration

**Requirements**:
1. Replace all `apt-get` with `apk add`
2. Replace `useradd` with `adduser`
3. Update ~15 package names
4. Remove Debian repository configuration (~20 lines)
5. Restructure package installation strategy
6. Test all dependencies work on Alpine
7. Update documentation

**Advantages**:
- ✅ Smaller image size (~100MB savings potential)
- ✅ Faster builds (apk is faster than apt)
- ✅ More secure (smaller attack surface)

**Disadvantages**:
- ❌ Significant effort required (2-4 hours)
- ❌ Extensive testing needed
- ❌ Potential compatibility issues with dependencies
- ❌ musl libc vs glibc differences may cause issues

**Time**: 2-4 hours to implement + testing time

### Option 3: Staged Migration

Create a separate `Dockerfile.alpine` for testing while maintaining the Debian version.

**Time**: 4-8 hours total

## Action Items

### Immediate (Next Commit):
1. ✅ Revert Dockerfile to `golang:1.25.5-bookworm`
2. ✅ Run these tests to verify: `./tests/dockerfile/run_all_tests.sh`
3. ✅ Commit with message: "Revert Dockerfile to Debian base for compatibility"

### Short-term (This Sprint):
1. Document Alpine migration requirements
2. Create ticket for Alpine migration (if desired)
3. Integrate Dockerfile tests into CI pipeline

### Long-term (Future):
1. Evaluate if Alpine migration provides sufficient benefit
2. If yes, create feature branch for Alpine migration
3. Implement and test thoroughly
4. Update all documentation

## Test Artifacts Location

All test code and documentation:
- `tests/dockerfile/validate_dockerfile.sh` - Shell validation (215 lines)
- `tests/dockerfile/dockerfile_test.go` - Go unit tests (196 lines)
- `tests/dockerfile/README.md` - Usage documentation
- `tests/dockerfile/TEST_SUMMARY.md` - Complete test catalog
- `tests/dockerfile/INTEGRATION.md` - CI/CD integration guide
- `tests/dockerfile/run_all_tests.sh` - Convenience script

## Running the Tests

```bash
# Run all tests (recommended)
./tests/dockerfile/run_all_tests.sh

# Run only shell validation
./tests/dockerfile/validate_dockerfile.sh

# Run only Go tests
go test -tags=small ./tests/dockerfile/...

# Run with verbose output
go test -tags=small -v ./tests/dockerfile/...
```

## Conclusion

The test suite successfully identified critical, build-breaking incompatibilities
between the Alpine Linux base image and the Debian-based commands in the Dockerfile.

**Test Suite Status**: ✅ Functioning correctly and providing actionable feedback

**Dockerfile Status**: ❌ Contains critical errors that prevent building

**Recommended Action**: Revert to Debian-based image immediately

---

*These tests should be run on every Dockerfile change to catch similar issues early.*