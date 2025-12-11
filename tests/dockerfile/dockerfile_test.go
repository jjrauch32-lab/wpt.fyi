//go:build small

// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package dockerfile_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDockerfileExists(t *testing.T) {
	dockerfilePath := getDockerfilePath(t)
	_, err := os.Stat(dockerfilePath)
	assert.NoError(t, err, "Dockerfile should exist and be readable")
}

func TestBaseImageFormat(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Dockerfile should contain FROM instruction")
	baseImage := matches[1]
	assert.True(t, strings.HasPrefix(baseImage, "golang:"),
		"Base image should be golang, got: %s", baseImage)
}

func TestBaseImageVersion(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+golang:([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should specify golang version")
	version := matches[1]
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	assert.True(t, versionRegex.MatchString(version),
		"Version should follow semver: %s", version)
}

func TestOSTypeDetection(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should contain FROM instruction")
	
	baseImage := strings.ToLower(matches[1])
	isAlpine := strings.Contains(baseImage, "alpine")
	isDebian := strings.Contains(baseImage, "bookworm") ||
		strings.Contains(baseImage, "bullseye") ||
		strings.Contains(baseImage, "debian")
	
	assert.True(t, isAlpine || isDebian,
		"Should be Alpine or Debian: %s", baseImage)
}

func TestPackageManagerCompatibility(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should contain FROM instruction")
	
	baseImage := strings.ToLower(matches[1])
	isAlpine := strings.Contains(baseImage, "alpine")
	
	hasApt := strings.Contains(content, "apt-get") || strings.Contains(content, "apt ")
	hasDpkg := strings.Contains(content, "dpkg")
	hasApk := strings.Contains(content, "apk ")
	
	if isAlpine {
		assert.False(t, hasApt, "Alpine should not use apt-get")
		assert.False(t, hasDpkg, "Alpine should not use dpkg")
		assert.True(t, hasApk || !hasApt, "Alpine should use apk")
	} else {
		assert.True(t, hasApt, "Debian should use apt-get")
	}
}

func getRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	require.NoError(t, err)
	
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	
	wd, _ := os.Getwd()
	return filepath.Join(wd, "../..")
}

func getDockerfilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(getRepoRoot(t), "Dockerfile")
}

func readDockerfile(t *testing.T) string {
	t.Helper()
	dockerfilePath := getDockerfilePath(t)
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err, "Failed to read Dockerfile")
	return string(content)
}

func TestUserManagementCommands(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches)
	
	baseImage := strings.ToLower(matches[1])
	isAlpine := strings.Contains(baseImage, "alpine")
	
	hasUseradd := strings.Contains(content, "useradd")
	hasAdduser := strings.Contains(content, "adduser")
	
	if isAlpine {
		assert.False(t, hasUseradd, "Alpine should use adduser not useradd")
	}
	assert.True(t, hasUseradd || hasAdduser, "Should create non-root user")
}

func TestNonRootUserCreation(t *testing.T) {
	content := readDockerfile(t)
	browserUserRegex := regexp.MustCompile(`(?i)(adduser|useradd).*browser`)
	assert.True(t, browserUserRegex.MatchString(content),
		"Should create browser user for security")
}

func TestRequiredDependencies(t *testing.T) {
	content := strings.ToLower(readDockerfile(t))
	requiredDeps := []string{"curl", "wget", "git"}
	for _, dep := range requiredDeps {
		assert.Contains(t, content, dep, "Should install %s", dep)
	}
}

func TestProjectSpecificDependencies(t *testing.T) {
	content := strings.ToLower(readDockerfile(t))
	projectDeps := []string{"java", "nodejs", "python", "gcloud"}
	for _, dep := range projectDeps {
		assert.Contains(t, content, dep, "Should reference %s", dep)
	}
}

func TestCloudSDKVersion(t *testing.T) {
	content := readDockerfile(t)
	envRegex := regexp.MustCompile(`(?m)^\s*ENV\s+CLOUD_SDK_VERSION\s*=\s*([0-9.]+)`)
	matches := envRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "CLOUD_SDK_VERSION should be set")
	version := matches[1]
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	assert.True(t, versionRegex.MatchString(version),
		"CLOUD_SDK_VERSION should follow semver: %s", version)
}

func TestLayerCount(t *testing.T) {
	content := readDockerfile(t)
	runRegex := regexp.MustCompile(`(?m)^RUN\s`)
	matches := runRegex.FindAllString(content, -1)
	runCount := len(matches)
	assert.LessOrEqual(t, runCount, 10,
		"Too many RUN instructions (%d)", runCount)
}

func TestDebianPackagesOnAlpine(t *testing.T) {
	content := readDockerfile(t)
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches)
	
	baseImage := strings.ToLower(matches[1])
	if !strings.Contains(baseImage, "alpine") {
		t.Skip("Test only for Alpine")
	}
	
	debianPkgs := []string{"python3-crcmod", "firefox-esr", "tox"}
	for _, pkg := range debianPkgs {
		if strings.Contains(content, pkg) {
			t.Errorf("Found Debian package '%s' on Alpine", pkg)
		}
	}
}