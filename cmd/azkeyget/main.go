// Package main provides the azkeyget CLI tool for retrieving secrets from Azure Key Vault.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/spf13/cobra"
)

// Build-time variables set by GoReleaser
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var (
	vaultURL       string
	secretName     string
	authMethod     string
	clientID       string
	clientSecret   string
	tenantID       string
	userAssignedID string
	debug          bool
)

func main() {
	rootCmd := &cobra.Command{
		Use:     "azkeyget",
		Short:   "Get secrets from Azure Key Vault",
		Long:    "A CLI tool to retrieve secrets from Azure Key Vault with support for multiple authentication methods",
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		RunE:    getSecret,
	}

	rootCmd.Flags().StringVarP(&vaultURL, "vault-url", "v", getEnvOrDefault("AZURE_KEYVAULT_URL", ""), "Azure Key Vault URL (required, env: AZURE_KEYVAULT_URL)")
	rootCmd.Flags().StringVarP(&secretName, "secret", "s", getEnvOrDefault("AZURE_KEYVAULT_SECRET_NAME", ""), "Secret name to retrieve (required, env: AZURE_KEYVAULT_SECRET_NAME)")
	rootCmd.Flags().StringVarP(&authMethod, "auth", "a", getEnvOrDefault("AZURE_AUTH_METHOD", "default"), "Authentication method: default, system-mi, user-mi, service-principal (env: AZURE_AUTH_METHOD)")
	rootCmd.Flags().StringVar(&clientID, "client-id", getEnvOrDefault("AZURE_CLIENT_ID", ""), "Client ID for service principal or user-assigned managed identity (env: AZURE_CLIENT_ID)")
	rootCmd.Flags().StringVar(&clientSecret, "client-secret", getEnvOrDefault("AZURE_CLIENT_SECRET", ""), "Client secret for service principal authentication (env: AZURE_CLIENT_SECRET)")
	rootCmd.Flags().StringVar(&tenantID, "tenant-id", getEnvOrDefault("AZURE_TENANT_ID", ""), "Tenant ID for service principal authentication (env: AZURE_TENANT_ID)")
	rootCmd.Flags().StringVar(&userAssignedID, "user-assigned-id", getEnvOrDefault("AZURE_USER_ASSIGNED_ID", ""), "User-assigned managed identity client ID (env: AZURE_USER_ASSIGNED_ID)")
	rootCmd.Flags().BoolVar(&debug, "debug", getEnvOrDefaultBool("AZURE_DEBUG", false), "Enable debug logging (env: AZURE_DEBUG)")

	if err := rootCmd.MarkFlagRequired("vault-url"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking vault-url as required: %v\n", err)
		os.Exit(1)
	}
	if err := rootCmd.MarkFlagRequired("secret"); err != nil {
		fmt.Fprintf(os.Stderr, "Error marking secret as required: %v\n", err)
		os.Exit(1)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getSecret(_ *cobra.Command, _ []string) error {
	// Setup debug logging
	setupDebugLogging()

	debugLog("Starting azkeyget execution")
	debugLog("Configuration:")
	debugLog("  Vault URL: %s", vaultURL)
	debugLog("  Secret Name: %s", secretName)
	debugLog("  Auth Method: %s", authMethod)
	debugLog("  Debug Enabled: %t", debug)

	ctx := context.Background()

	debugLog("Creating credential with method: %s", authMethod)
	credential, err := createCredential()
	if err != nil {
		debugLog("Failed to create credential: %v", err)
		return fmt.Errorf("failed to create credential: %w", err)
	}
	debugLog("Successfully created credential")

	debugLog("Creating Key Vault client for URL: %s", vaultURL)
	client, err := azsecrets.NewClient(vaultURL, credential, nil)
	if err != nil {
		debugLog("Failed to create Key Vault client: %v", err)
		return fmt.Errorf("failed to create Key Vault client: %w", err)
	}
	debugLog("Successfully created Key Vault client")

	debugLog("Retrieving secret: %s", secretName)
	response, err := client.GetSecret(ctx, secretName, "", nil)
	if err != nil {
		debugLog("Failed to retrieve secret '%s': %v", secretName, err)
		return fmt.Errorf("failed to get secret '%s': %w", secretName, err)
	}
	debugLog("Successfully retrieved secret")

	if response.Value == nil {
		debugLog("Secret '%s' has no value", secretName)
		return fmt.Errorf("secret '%s' has no value", secretName)
	}

	debugLog("Secret retrieved successfully, outputting to stdout")
	fmt.Print(*response.Value)
	debugLog("Operation completed successfully")
	return nil
}

func getEnvOrDefault(envVar, defaultValue string) string {
	if value := os.Getenv(envVar); value != "" {
		return value
	}
	return defaultValue
}

func getEnvOrDefaultBool(envVar string, defaultValue bool) bool {
	if value := os.Getenv(envVar); value != "" {
		return value == "true" || value == "1" || value == "yes" || value == "on"
	}
	return defaultValue
}

// setupDebugLogging configures the debug logger
func setupDebugLogging() {
	if !debug {
		// Disable debug logging by directing to discard
		log.SetOutput(os.Stderr)
		log.SetFlags(0)
		log.SetPrefix("")
		return
	}

	// Setup debug logging to stderr with timestamp
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[DEBUG] azkeyget: ")
}

// debugLog outputs debug information if debug mode is enabled
func debugLog(format string, args ...interface{}) {
	if debug {
		log.Printf(format, args...)
	}
}

func createCredential() (azcore.TokenCredential, error) {
	debugLog("Creating credential for auth method: %s", authMethod)

	switch authMethod {
	case "default":
		debugLog("Using DefaultAzureCredential")
		return azidentity.NewDefaultAzureCredential(nil)

	case "system-mi":
		debugLog("Using system managed identity")
		return azidentity.NewManagedIdentityCredential(nil)

	case "user-mi":
		if userAssignedID != "" {
			debugLog("Using user-assigned managed identity with ID: %s", userAssignedID)
			options := &azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(userAssignedID),
			}
			return azidentity.NewManagedIdentityCredential(options)
		} else if clientID != "" {
			debugLog("Using user-assigned managed identity with client ID: %s", clientID)
			options := &azidentity.ManagedIdentityCredentialOptions{
				ID: azidentity.ClientID(clientID),
			}
			return azidentity.NewManagedIdentityCredential(options)
		}
		debugLog("User-assigned managed identity requires client ID or user-assigned ID")
		return nil, fmt.Errorf("user-assigned managed identity requires --client-id or --user-assigned-id")

	case "service-principal":
		if clientID == "" || clientSecret == "" || tenantID == "" {
			debugLog("Service principal authentication missing required parameters")
			debugLog("  Client ID provided: %t", clientID != "")
			debugLog("  Client Secret provided: %t", clientSecret != "")
			debugLog("  Tenant ID provided: %t", tenantID != "")
			return nil, fmt.Errorf("service principal authentication requires --client-id, --client-secret, and --tenant-id")
		}
		debugLog("Using service principal with client ID: %s, tenant ID: %s", clientID, tenantID)
		return azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)

	default:
		debugLog("Unsupported authentication method: %s", authMethod)
		return nil, fmt.Errorf("unsupported authentication method: %s", authMethod)
	}
}
