# Copyright 2024 The WPT Dashboard Project. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

import os
import re
import unittest
from pathlib import Path


class RequirementsTest(unittest.TestCase):
    """Test suite for validating requirements.txt dependencies."""

    @classmethod
    def setUpClass(cls):
        """Load requirements.txt once for all tests."""
        requirements_path = Path(__file__).parent / 'requirements.txt'
        with open(requirements_path, 'r') as f:
            cls.requirements_content = f.read()
        
        # Parse requirements into a dict
        cls.requirements = {}
        for line in cls.requirements_content.split('\n'):
            line = line.strip()
            if line and not line.startswith('#'):
                match = re.match(r'^([a-zA-Z0-9\-_]+)==([0-9.]+)', line)
                if match:
                    package, version = match.groups()
                    cls.requirements[package] = version

    def test_requirements_file_exists(self):
        """Test that requirements.txt exists and is readable."""
        requirements_path = Path(__file__).parent / 'requirements.txt'
        self.assertTrue(requirements_path.exists(),
                       "requirements.txt should exist")
        self.assertTrue(requirements_path.is_file(),
                       "requirements.txt should be a file")

    def test_grpcio_version(self):
        """Test that grpcio is pinned to compatible version 1.53.2."""
        self.assertIn('grpcio', self.requirements,
                     "grpcio should be in requirements")
        self.assertEqual(self.requirements['grpcio'], '1.53.2',
                        "grpcio should be version 1.53.2 for compatibility")

    def test_grpcio_status_matches_grpcio(self):
        """Test that grpcio-status version matches grpcio version."""
        self.assertIn('grpcio-status', self.requirements,
                     "grpcio-status should be in requirements")
        self.assertEqual(self.requirements['grpcio-status'],
                        self.requirements['grpcio'],
                        "grpcio-status version should match grpcio version")

    def test_no_release_candidate_versions(self):
        """Test that no packages use release candidate versions."""
        rc_pattern = re.compile(r'==\d+\.\d+.*-rc', re.IGNORECASE)
        rc_packages = []
        
        for line in self.requirements_content.split('\n'):
            if rc_pattern.search(line):
                rc_packages.append(line.strip())
        
        self.assertEqual(len(rc_packages), 0,
                        f"Should not use RC versions, found: {rc_packages}")

    def test_critical_dependencies_present(self):
        """Test that all critical dependencies are present."""
        critical_deps = [
            'flask',
            'google-cloud-datastore',
            'google-cloud-storage',
            'gunicorn',
            'requests',
        ]
        
        for dep in critical_deps:
            self.assertIn(dep, self.requirements,
                         f"Critical dependency {dep} should be in requirements")

    def test_google_api_core_version(self):
        """Test that google-api-core is present with version."""
        self.assertIn('google-api-core', self.requirements,
                     "google-api-core should be in requirements")
        version = self.requirements['google-api-core']
        self.assertRegex(version, r'^\d+\.\d+\.\d+$',
                        "google-api-core should have semantic version")

    def test_flask_version(self):
        """Test that Flask is using a stable version."""
        self.assertIn('flask', self.requirements,
                     "Flask should be in requirements")
        version = self.requirements['flask']
        major = int(version.split('.')[0])
        self.assertGreaterEqual(major, 3,
                               "Flask should be version 3.x or higher")

    def test_gunicorn_version(self):
        """Test that gunicorn is present and versioned."""
        self.assertIn('gunicorn', self.requirements,
                     "gunicorn should be in requirements")
        version = self.requirements['gunicorn']
        self.assertRegex(version, r'^\d+\.\d+\.\d+$',
                        "gunicorn should have semantic version")

    def test_security_packages_present(self):
        """Test that security-related packages are present."""
        security_deps = [
            'certifi',
            'urllib3',
        ]
        
        for dep in security_deps:
            self.assertIn(dep, self.requirements,
                         f"Security package {dep} should be in requirements")

    def test_google_cloud_packages_consistent(self):
        """Test that Google Cloud packages are at compatible versions."""
        google_packages = {
            k: v for k, v in self.requirements.items()
            if k.startswith('google-')
        }
        
        self.assertGreater(len(google_packages), 3,
                          "Should have multiple Google Cloud packages")
        
        # Check that google-auth is present and versioned
        self.assertIn('google-auth', self.requirements,
                     "google-auth should be in requirements")

    def test_typing_extensions_present(self):
        """Test that typing-extensions is present for type hints."""
        self.assertIn('typing-extensions', self.requirements,
                     "typing-extensions should be in requirements")

    def test_dev_tools_present(self):
        """Test that development tools are present."""
        dev_tools = [
            'flake8',
            'mypy',
        ]
        
        for tool in dev_tools:
            self.assertIn(tool, self.requirements,
                         f"Development tool {tool} should be in requirements")

    def test_werkzeug_version(self):
        """Test that Werkzeug is compatible with Flask."""
        self.assertIn('werkzeug', self.requirements,
                     "werkzeug should be in requirements")
        version = self.requirements['werkzeug']
        major = int(version.split('.')[0])
        self.assertGreaterEqual(major, 3,
                               "werkzeug should be version 3.x for Flask 3.x")

    def test_no_conflicting_versions(self):
        """Test that there are no obvious version conflicts."""
        # Check that grpcio and grpcio-status are the same
        grpcio_ver = self.requirements.get('grpcio')
        grpcio_status_ver = self.requirements.get('grpcio-status')
        
        if grpcio_ver and grpcio_status_ver:
            self.assertEqual(grpcio_ver, grpcio_status_ver,
                           "grpcio and grpcio-status must have matching versions")

    def test_protobuf_version_compatible(self):
        """Test that protobuf version is compatible with grpcio."""
        self.assertIn('protobuf', self.requirements,
                     "protobuf should be in requirements")
        
        protobuf_version = self.requirements['protobuf']
        major = int(protobuf_version.split('.')[0])
        
        # Protobuf 4.x is compatible with grpcio 1.53.x
        self.assertGreaterEqual(major, 4,
                               "protobuf should be version 4.x or higher")

    def test_all_versions_pinned(self):
        """Test that all package versions are pinned (not using >= or ~=)."""
        for line in self.requirements_content.split('\n'):
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                if '>=' in line or '~=' in line or '!=' in line:
                    # Exception for zipp which uses >= for security
                    if not line.startswith('zipp>='):
                        self.fail(f"Package should use == pinning: {line}")

    def test_zipp_security_pin(self):
        """Test that zipp has security pin as noted in requirements."""
        zipp_line = None
        for line in self.requirements_content.split('\n'):
            if line.strip().startswith('zipp'):
                zipp_line = line
                break
        
        self.assertIsNotNone(zipp_line, "zipp should be in requirements")
        self.assertIn('>=', zipp_line,
                     "zipp should use >= for security vulnerability fix")
        self.assertIn('3.19.1', zipp_line,
                     "zipp should be pinned to 3.19.1 or higher")

    def test_requests_version_secure(self):
        """Test that requests is using a secure version."""
        self.assertIn('requests', self.requirements,
                     "requests should be in requirements")
        version = self.requirements['requests']
        major = int(version.split('.')[0])
        minor = int(version.split('.')[1])
        
        # requests 2.32.x includes security fixes
        self.assertEqual(major, 2, "requests should be version 2.x")
        self.assertGreaterEqual(minor, 32,
                               "requests should be 2.32.x or higher for security")

    def test_filelock_present(self):
        """Test that filelock is present for concurrency control."""
        self.assertIn('filelock', self.requirements,
                     "filelock should be in requirements")

    def test_comments_indicate_dependencies(self):
        """Test that comments properly indicate which packages require others."""
        # Check that major packages have dependency comments
        packages_with_comments = []
        current_package = None
        
        for line in self.requirements_content.split('\n'):
            line = line.strip()
            if line and not line.startswith('#'):
                match = re.match(r'^([a-zA-Z0-9\-_]+)==', line)
                if match:
                    current_package = match.group(1)
            elif line.startswith('#') and 'via' in line.lower():
                if current_package:
                    packages_with_comments.append(current_package)
        
        # Should have dependency comments for at least some packages
        self.assertGreater(len(packages_with_comments), 10,
                          "Should have dependency comments for major packages")


if __name__ == '__main__':
    unittest.main()