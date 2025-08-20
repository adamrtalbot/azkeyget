package main

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// cleanTestEnvironment cleans all Azure environment variables for testing
func cleanTestEnvironment(t *testing.T) {
	envVarsToClean := []string{
		"AZURE_KEYVAULT_URL",
		"AZURE_KEYVAULT_SECRET_NAME",
		"AZURE_AUTH_METHOD",
		"AZURE_CLIENT_ID",
		"AZURE_CLIENT_SECRET",
		"AZURE_TENANT_ID",
		"AZURE_USER_ASSIGNED_ID",
	}

	for _, envVar := range envVarsToClean {
		if err := os.Unsetenv(envVar); err != nil {
			t.Fatalf("Failed to unset environment variable %s: %v", envVar, err)
		}
	}
}

// setTestEnvironment sets environment variables for testing
func setTestEnvironment(envVars map[string]string) {
	for key, value := range envVars {
		if err := os.Setenv(key, value); err != nil {
			// In test context, we can panic since this is a setup failure
			panic(err)
		}
	}
}

// validateHelpOutput validates that help command works correctly
func validateHelpOutput(t *testing.T, err error, stdout, stderr string) {
	if err != nil {
		t.Errorf("Help command should not error, got: %v", err)
	}
	if !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "Usage:") {
		t.Errorf("Help output should contain usage information")
	}
}

// validateErrorOutput validates error cases
func validateErrorOutput(t *testing.T, err error, stderr, expectedError string) {
	if err == nil {
		t.Errorf("Expected error but command succeeded")
		return
	}
	if expectedError != "" && !strings.Contains(stderr, expectedError) {
		t.Errorf("Error output %q should contain %q", stderr, expectedError)
	}
}

// validateSuccessOutput validates success cases
func validateSuccessOutput(t *testing.T, err error, stderr string) {
	if err != nil {
		t.Errorf("Unexpected error: %v\nStderr: %s", err, stderr)
	}
}

func TestCLIFlags(t *testing.T) {
	// Build the binary for testing
	buildCmd := exec.Command("go", "build", "-o", "azkeyget_test", ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer func() {
		if err := os.Remove("azkeyget_test"); err != nil {
			// Log but don't fail test for cleanup errors
			t.Logf("Failed to remove test binary: %v", err)
		}
	}()

	tests := []struct {
		name          string
		args          []string
		envVars       map[string]string
		expectError   bool
		errorContains string
	}{
		{
			name:          "missing vault-url flag",
			args:          []string{"--secret", "test-secret"},
			expectError:   true,
			errorContains: "required flag(s) \"vault-url\" not set",
		},
		{
			name:          "missing secret flag",
			args:          []string{"--vault-url", "https://test.vault.azure.net/"},
			expectError:   true,
			errorContains: "required flag(s) \"secret\" not set",
		},
		{
			name:        "help flag",
			args:        []string{"--help"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean and set environment
			cleanTestEnvironment(t)
			setTestEnvironment(tt.envVars)
			defer cleanTestEnvironment(t)

			// Execute command
			cmd := exec.Command("./azkeyget_test", tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Validate output based on test case
			switch tt.name {
			case "help flag":
				validateHelpOutput(t, err, stdout.String(), stderr.String())
			default:
				if tt.expectError {
					validateErrorOutput(t, err, stderr.String(), tt.errorContains)
				} else {
					validateSuccessOutput(t, err, stderr.String())
				}
			}
		})
	}
}

// TestCLIFlagPrecedence tests that CLI flags override environment variables
// This test is simplified to avoid hanging on Azure API calls
func TestCLIFlagPrecedence(t *testing.T) {
	t.Skip("Skipping CLI precedence test to avoid Azure API timeouts - precedence is tested in unit tests")
}
