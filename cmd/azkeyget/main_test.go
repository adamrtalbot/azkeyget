package main

import (
	"os"
	"testing"
)

func TestGetEnvOrDefaultBool(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "environment variable true",
			envVar:       "TEST_BOOL_VAR",
			envValue:     "true",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "environment variable 1",
			envVar:       "TEST_BOOL_VAR",
			envValue:     "1",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "environment variable yes",
			envVar:       "TEST_BOOL_VAR",
			envValue:     "yes",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "environment variable on",
			envVar:       "TEST_BOOL_VAR",
			envValue:     "on",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "environment variable false",
			envVar:       "TEST_BOOL_VAR",
			envValue:     "false",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "environment variable not set",
			envVar:       "TEST_BOOL_VAR_NOT_SET",
			envValue:     "",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env var
			os.Unsetenv(tt.envVar)

			// Set env var if test requires it
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
			}

			result := getEnvOrDefaultBool(tt.envVar, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefaultBool(%s, %t) = %t; want %t",
					tt.envVar, tt.defaultValue, result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.envVar)
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "environment variable exists",
			envVar:       "TEST_VAR",
			envValue:     "env_value",
			defaultValue: "default_value",
			expected:     "env_value",
		},
		{
			name:         "environment variable empty",
			envVar:       "TEST_VAR_EMPTY",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "environment variable not set",
			envVar:       "TEST_VAR_NOT_SET",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing env var
			os.Unsetenv(tt.envVar)

			// Set env var if test requires it
			if tt.envValue != "" || tt.name == "environment variable empty" {
				os.Setenv(tt.envVar, tt.envValue)
			}

			result := getEnvOrDefault(tt.envVar, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnvOrDefault(%s, %s) = %s; want %s",
					tt.envVar, tt.defaultValue, result, tt.expected)
			}

			// Clean up
			os.Unsetenv(tt.envVar)
		})
	}
}

func TestCreateCredential(t *testing.T) {
	// Disable debug logging for tests
	originalDebug := debug
	debug = false
	defer func() { debug = originalDebug }()
	tests := []struct {
		name           string
		authMethod     string
		clientID       string
		clientSecret   string
		tenantID       string
		userAssignedID string
		shouldError    bool
		errorContains  string
	}{
		{
			name:        "default auth method",
			authMethod:  "default",
			shouldError: false,
		},
		{
			name:        "system managed identity",
			authMethod:  "system-mi",
			shouldError: false,
		},
		{
			name:        "user managed identity with client-id",
			authMethod:  "user-mi",
			clientID:    "test-client-id",
			shouldError: false,
		},
		{
			name:           "user managed identity with user-assigned-id",
			authMethod:     "user-mi",
			userAssignedID: "test-user-assigned-id",
			shouldError:    false,
		},
		{
			name:          "user managed identity without id",
			authMethod:    "user-mi",
			shouldError:   true,
			errorContains: "requires --client-id or --user-assigned-id",
		},
		{
			name:         "service principal with all params",
			authMethod:   "service-principal",
			clientID:     "test-client-id",
			clientSecret: "test-client-secret",
			tenantID:     "test-tenant-id",
			shouldError:  false,
		},
		{
			name:          "service principal missing client-id",
			authMethod:    "service-principal",
			clientSecret:  "test-client-secret",
			tenantID:      "test-tenant-id",
			shouldError:   true,
			errorContains: "requires --client-id, --client-secret, and --tenant-id",
		},
		{
			name:          "service principal missing client-secret",
			authMethod:    "service-principal",
			clientID:      "test-client-id",
			tenantID:      "test-tenant-id",
			shouldError:   true,
			errorContains: "requires --client-id, --client-secret, and --tenant-id",
		},
		{
			name:          "service principal missing tenant-id",
			authMethod:    "service-principal",
			clientID:      "test-client-id",
			clientSecret:  "test-client-secret",
			shouldError:   true,
			errorContains: "requires --client-id, --client-secret, and --tenant-id",
		},
		{
			name:          "unsupported auth method",
			authMethod:    "invalid-method",
			shouldError:   true,
			errorContains: "unsupported authentication method",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global variables as they would be set by cobra
			authMethod = tt.authMethod
			clientID = tt.clientID
			clientSecret = tt.clientSecret
			tenantID = tt.tenantID
			userAssignedID = tt.userAssignedID

			credential, err := createCredential()

			if tt.shouldError {
				if err == nil {
					t.Errorf("createCredential() expected error but got none")
					return
				}
				if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("createCredential() error = %v, should contain %s", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("createCredential() unexpected error: %v", err)
					return
				}
				// Verify that we got a non-nil credential
				if credential == nil {
					t.Errorf("createCredential() returned nil credential for method %s", tt.authMethod)
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(len(substr) == 0 ||
			findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func TestEnvironmentVariableIntegration(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected map[string]string
	}{
		{
			name: "all environment variables set",
			envVars: map[string]string{
				"AZURE_KEYVAULT_URL":         "https://test.vault.azure.net/",
				"AZURE_KEYVAULT_SECRET_NAME": "test-secret",
				"AZURE_AUTH_METHOD":          "system-mi",
				"AZURE_CLIENT_ID":            "test-client-id",
				"AZURE_CLIENT_SECRET":        "test-client-secret",
				"AZURE_TENANT_ID":            "test-tenant-id",
				"AZURE_USER_ASSIGNED_ID":     "test-user-assigned-id",
			},
			expected: map[string]string{
				"vault-url":        "https://test.vault.azure.net/",
				"secret":           "test-secret",
				"auth":             "system-mi",
				"client-id":        "test-client-id",
				"client-secret":    "test-client-secret",
				"tenant-id":        "test-tenant-id",
				"user-assigned-id": "test-user-assigned-id",
			},
		},
		{
			name:    "no environment variables set",
			envVars: map[string]string{},
			expected: map[string]string{
				"vault-url":        "",
				"secret":           "",
				"auth":             "default",
				"client-id":        "",
				"client-secret":    "",
				"tenant-id":        "",
				"user-assigned-id": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up environment
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
				os.Unsetenv(envVar)
			}

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Test getEnvOrDefault for each variable
			testCases := map[string]struct {
				envVar       string
				defaultValue string
			}{
				"vault-url":        {"AZURE_KEYVAULT_URL", ""},
				"secret":           {"AZURE_KEYVAULT_SECRET_NAME", ""},
				"auth":             {"AZURE_AUTH_METHOD", "default"},
				"client-id":        {"AZURE_CLIENT_ID", ""},
				"client-secret":    {"AZURE_CLIENT_SECRET", ""},
				"tenant-id":        {"AZURE_TENANT_ID", ""},
				"user-assigned-id": {"AZURE_USER_ASSIGNED_ID", ""},
			}

			for key, tc := range testCases {
				result := getEnvOrDefault(tc.envVar, tc.defaultValue)
				expected := tt.expected[key]
				if result != expected {
					t.Errorf("getEnvOrDefault(%s, %s) = %s; want %s",
						tc.envVar, tc.defaultValue, result, expected)
				}
			}

			// Clean up
			for _, envVar := range envVarsToClean {
				os.Unsetenv(envVar)
			}
		})
	}
}
