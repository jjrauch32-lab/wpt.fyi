//go:build small

// Copyright 2024 The WPT Dashboard Project. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package docker_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper functions

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

func readDockerfile(t *testing.T, path string) string {
	t.Helper()
	repoRoot := getRepoRoot(t)
	dockerfilePath := filepath.Join(repoRoot, path)
	content, err := os.ReadFile(dockerfilePath)
	require.NoError(t, err, "Failed to read %s", path)
	return string(content)
}

// Main Dockerfile Tests

func TestMainDockerfileExists(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.NotEmpty(t, content, "Dockerfile should not be empty")
}

func TestMainDockerfileBaseImage(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Dockerfile should contain FROM instruction")

	baseImage := matches[1]
	assert.True(t, strings.HasPrefix(baseImage, "golang:"),
		"Base image should be golang, got: %s", baseImage)
}

func TestMainDockerfileUsesBookworm(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should contain FROM instruction")

	baseImage := strings.ToLower(matches[1])
	assert.Contains(t, baseImage, "bookworm",
		"Should use Debian Bookworm, got: %s", baseImage)
	assert.NotContains(t, baseImage, "alpine",
		"Should not use Alpine (incompatible with apt-get commands)")
}

func TestMainDockerfileGolangVersion(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+golang:([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should specify golang version")

	version := matches[1]
	versionRegex := regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)
	assert.True(t, versionRegex.MatchString(version),
		"Version should follow semver format: %s", version)

	// Validate it's the expected version (1.25.3)
	assert.Equal(t, "1.25.3", version,
		"Should use golang 1.25.3 for compatibility")
}

func TestMainDockerfileUsesAptGet(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.Contains(t, content, "apt-get",
		"Dockerfile should use apt-get (Debian package manager)")
	assert.NotContains(t, content, "apk add",
		"Dockerfile should not use apk (Alpine package manager)")
}

func TestMainDockerfileUsesUseradd(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.Contains(t, content, "useradd",
		"Dockerfile should use useradd (Debian user management)")
	
	// Verify only one useradd command (no duplicates)
	useradds := regexp.MustCompile(`useradd`).FindAllString(content, -1)
	assert.Len(t, useradds, 1,
		"Should have exactly one useradd command, found %d", len(useradds))
}

func TestMainDockerfileBrowserUserCreation(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	browserUserRegex := regexp.MustCompile(`useradd.*browser`)
	assert.True(t, browserUserRegex.MatchString(content),
		"Should create browser user for security")

	// Verify user ID is 9999
	uidRegex := regexp.MustCompile(`--uid\s+9999`)
	assert.True(t, uidRegex.MatchString(content),
		"Browser user should have UID 9999")
}

func TestMainDockerfileRequiredDependencies(t *testing.T) {
	content := strings.ToLower(readDockerfile(t, "Dockerfile"))
	
	requiredDeps := []string{
		"curl",
		"wget",
		"git",
		"firefox-esr",
		"nodejs",
		"python3",
		"sudo",
		"xvfb",
	}
	
	for _, dep := range requiredDeps {
		assert.Contains(t, content, dep,
			"Should install required dependency: %s", dep)
	}
}

func TestMainDockerfileJavaInstallation(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.Contains(t, content, "java-11-amazon-corretto-jdk",
		"Should install Java 11 Amazon Corretto JDK")
	assert.Contains(t, content, "corretto.aws",
		"Should add Corretto repository")
}

func TestMainDockerfileCloudSDKVersion(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	envRegex := regexp.MustCompile(`(?m)^\s*ENV\s+CLOUD_SDK_VERSION\s*=\s*([0-9.]+)`)
	matches := envRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "CLOUD_SDK_VERSION should be set")
	
	version := matches[1]
	assert.Equal(t, "527.0.0", version,
		"Should use Cloud SDK version 527.0.0 (Java 11 compatible)")
}

func TestMainDockerfileCloudSDKComponents(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	requiredComponents := []string{
		"google-cloud-cli",
		"google-cloud-cli-app-engine-python",
		"google-cloud-cli-app-engine-python-extras",
		"google-cloud-cli-app-engine-go",
		"google-cloud-cli-datastore-emulator",
	}
	
	for _, component := range requiredComponents {
		assert.Contains(t, content, component,
			"Should install Cloud SDK component: %s", component)
	}
}

func TestMainDockerfileNodeVersion(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	nodeVersionRegex := regexp.MustCompile(`NODE_VERSION="([0-9]+\.x)"`)
	matches := nodeVersionRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "NODE_VERSION should be set")
	
	assert.Equal(t, "18.x", matches[1],
		"Should use Node.js 18.x")
}

func TestMainDockerfilePython3Version(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.Contains(t, content, "python3.11",
		"Should install Python 3.11")
}

func TestMainDockerfileNoDuplicateUserCreation(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	// Count occurrences of user creation commands
	userCreationCommands := []string{
		"useradd.*browser",
		"adduser.*browser",
	}
	
	totalMatches := 0
	for _, pattern := range userCreationCommands {
		matches := regexp.MustCompile(pattern).FindAllString(content, -1)
		totalMatches += len(matches)
	}
	
	assert.LessOrEqual(t, totalMatches, 1,
		"Should not have duplicate user creation commands")
}

func TestMainDockerfileLayerOptimization(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	// Count RUN instructions
	runRegex := regexp.MustCompile(`(?m)^RUN\s`)
	matches := runRegex.FindAllString(content, -1)
	runCount := len(matches)
	
	assert.LessOrEqual(t, runCount, 10,
		"Should have reasonable number of RUN instructions for layer optimization, got %d", runCount)
}

func TestMainDockerfileSudoersConfiguration(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	assert.Contains(t, content, "echo \"root ALL=(ALL:ALL) ALL\" > /etc/sudoers",
		"Should configure sudoers for PATH inheritance")
}

func TestMainDockerfileSecurityPractices(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	// Check for secure package installation
	assert.Contains(t, content, "--no-install-suggests",
		"Should use --no-install-suggests to minimize attack surface")
	
	// Check that browser user is created (non-root)
	assert.Contains(t, content, "browser",
		"Should create non-root browser user")
}

func TestMainDockerfileRepositoryConfiguration(t *testing.T) {
	content := readDockerfile(t, "Dockerfile")
	
	// Verify signed repositories
	repositories := []string{
		"signed-by=/usr/share/keyrings/corretto.gpg",
		"signed-by=/usr/share/keyrings/nodesource.gpg",
		"signed-by=/usr/share/keyrings/cloud.google.gpg",
	}
	
	for _, repo := range repositories {
		assert.Contains(t, content, repo,
			"Should use signed repository: %s", repo)
	}
}

// Results Processor Dockerfile Tests

func TestResultsProcessorDockerfileExists(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	assert.NotEmpty(t, content, "results-processor/Dockerfile should not be empty")
}

func TestResultsProcessorDockerfileBaseImage(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should contain FROM instruction")
	
	baseImage := matches[1]
	assert.True(t, strings.HasPrefix(baseImage, "python:"),
		"Base image should be python, got: %s", baseImage)
}

func TestResultsProcessorDockerfilePythonVersion(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+python:([0-9]+\.[0-9]+\.[0-9]+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should specify python version")
	
	version := matches[1]
	assert.Equal(t, "3.11.14", version,
		"Should use Python 3.11.14 (stable, not RC)")
}

func TestResultsProcessorDockerfileUsesBookworm(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	fromRegex := regexp.MustCompile(`(?m)^\s*FROM\s+(\S+)`)
	matches := fromRegex.FindStringSubmatch(content)
	require.NotEmpty(t, matches, "Should contain FROM instruction")
	
	baseImage := strings.ToLower(matches[1])
	assert.Contains(t, baseImage, "bookworm",
		"Should use Debian Bookworm for stability")
	assert.NotContains(t, baseImage, "rc",
		"Should not use release candidate version")
	assert.NotContains(t, baseImage, "trixie",
		"Should use stable Bookworm, not testing Trixie")
}

func TestResultsProcessorDockerfileVirtualenv(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, content, "virtualenv",
		"Should use virtualenv for dependency isolation")
	assert.Contains(t, content, "virtualenv -p python3.11",
		"Should create virtualenv with Python 3.11")
	assert.Contains(t, content, "VIRTUAL_ENV",
		"Should set VIRTUAL_ENV environment variable")
}

func TestResultsProcessorDockerfileGcloudSDK(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, content, "sdk.cloud.google.com",
		"Should install gcloud SDK")
	assert.Contains(t, content, "--disable-prompts",
		"Should install gcloud non-interactively")
	assert.Contains(t, content, "/opt/google-cloud-sdk/bin",
		"Should add gcloud to PATH")
}

func TestResultsProcessorDockerfileGCECacheRemoval(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, content, "rm -f $HOME/.config/gcloud/gce",
		"Should remove GCE cache to prevent authentication issues")
}

func TestResultsProcessorDockerfileWorkdir(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	workdirRegex := regexp.MustCompile(`(?m)^\s*WORKDIR\s+/app`)
	assert.True(t, workdirRegex.MatchString(content),
		"Should set WORKDIR to /app")
}

func TestResultsProcessorDockerfileGunicornConfiguration(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, content, "gunicorn",
		"Should use gunicorn as WSGI server")
	assert.Contains(t, content, "--workers 2",
		"Should use 2 workers (1 for tasks, 1 for health checks)")
	assert.Contains(t, content, "--timeout 7200",
		"Should have long timeout for processing tasks")
}

func TestResultsProcessorDockerfileDependencies(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	requiredDeps := []string{
		"python3-crcmod",
		"python3-virtualenv",
	}
	
	for _, dep := range requiredDeps {
		assert.Contains(t, content, dep,
			"Should install required dependency: %s", dep)
	}
}

func TestResultsProcessorDockerfileCleanup(t *testing.T) {
	content := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, content, "apt-get clean",
		"Should clean apt cache to reduce image size")
}

// Cross-validation tests

func TestBothDockerfilesUseConsistentDebianVersion(t *testing.T) {
	mainContent := readDockerfile(t, "Dockerfile")
	processorContent := readDockerfile(t, "results-processor/Dockerfile")
	
	mainBookworm := strings.Contains(strings.ToLower(mainContent), "bookworm")
	processorBookworm := strings.Contains(strings.ToLower(processorContent), "bookworm")
	
	assert.True(t, mainBookworm && processorBookworm,
		"Both Dockerfiles should consistently use Debian Bookworm")
}

func TestNeitherDockerfileUsesAlpine(t *testing.T) {
	mainContent := strings.ToLower(readDockerfile(t, "Dockerfile"))
	processorContent := strings.ToLower(readDockerfile(t, "results-processor/Dockerfile"))
	
	assert.NotContains(t, mainContent, "alpine",
		"Main Dockerfile should not use Alpine (incompatible)")
	assert.NotContains(t, processorContent, "alpine",
		"Processor Dockerfile should not use Alpine (incompatible)")
}

func TestBothDockerfilesUsePython311(t *testing.T) {
	mainContent := readDockerfile(t, "Dockerfile")
	processorContent := readDockerfile(t, "results-processor/Dockerfile")
	
	assert.Contains(t, mainContent, "python3.11",
		"Main Dockerfile should reference Python 3.11")
	assert.Contains(t, processorContent, "python3.11",
		"Processor Dockerfile should use Python 3.11")
}