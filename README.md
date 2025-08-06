# azkeyget

A lightweight CLI tool for retrieving secrets from Azure Key Vault.

## Motivation

I wanted a simple tool that I could use in my scripts and CI/CD pipeline to retrieve secrets from Azure Key Vault without installing the full Azure CLI. Ideally, it would be a small binary I could install on a variety of platforms without additional configuration or dependencies, while being completely portable.

`azkeyget` is a small, portable CLI tool that retrieves secrets from Azure Key Vault. It's designed to be easy to use and package so you can quickly add it to machine images or containers.

## Installation

### Download the latest release

Download the latest release from the [releases page](https://github.com/adamrtalbot/azkeyget/releases).

### Build from Source

```bash
git clone <repository-url>
cd azkeyget
make build
# or
go build -o azkeyget ./cmd/azkeyget
```

### Prerequisites

- Go 1.23 or later
- Appropriate Azure permissions to access the target Key Vault

## Usage

### Basic Usage

```bash
azkeyget --vault-url <VAULT_URL> --secret <SECRET_NAME> [OPTIONS]
```

All parameters can be provided via command line flags or environment variables. Environment variables are used as defaults when CLI flags are not specified.

### Environment Variables

| Environment Variable | CLI Flag | Description |
|---------------------|----------|-------------|
| `AZURE_KEYVAULT_URL` | `--vault-url` | Azure Key Vault URL |
| `AZURE_KEYVAULT_SECRET_NAME` | `--secret` | Name of the secret to retrieve |
| `AZURE_AUTH_METHOD` | `--auth` | Authentication method |
| `AZURE_CLIENT_ID` | `--client-id` | Client ID for authentication |
| `AZURE_CLIENT_SECRET` | `--client-secret` | Client secret for service principal |
| `AZURE_TENANT_ID` | `--tenant-id` | Tenant ID for service principal |
| `AZURE_USER_ASSIGNED_ID` | `--user-assigned-id` | User-assigned managed identity client ID |
| `AZURE_DEBUG` | `--debug` | Enable debug logging (true/1/yes/on) |

### Authentication Methods

#### Default Authentication (Recommended)

Uses the Azure SDK's default credential chain, which tries multiple authentication methods in order:

```bash
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret

# Or using environment variables
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_KEYVAULT_SECRET_NAME=mysecret
azkeyget
```

#### System Managed Identity

For Azure VMs, App Service, or other Azure resources with system-assigned managed identity:

```bash
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret --auth system-mi

# Or using environment variables
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_KEYVAULT_SECRET_NAME=mysecret
export AZURE_AUTH_METHOD=system-mi
azkeyget
```

#### User-Assigned Managed Identity

For resources with user-assigned managed identity:

```bash
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret --auth user-mi --client-id 12345678-1234-1234-1234-123456789012

# Or using environment variables
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_KEYVAULT_SECRET_NAME=mysecret
export AZURE_AUTH_METHOD=user-mi
export AZURE_CLIENT_ID=12345678-1234-1234-1234-123456789012
azkeyget
```

#### Service Principal

For application authentication with client credentials:

```bash
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret --auth service-principal \
  --client-id YOUR_CLIENT_ID \
  --client-secret YOUR_CLIENT_SECRET \
  --tenant-id YOUR_TENANT_ID

# Or using environment variables
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_KEYVAULT_SECRET_NAME=mysecret
export AZURE_AUTH_METHOD=service-principal
export AZURE_CLIENT_ID=YOUR_CLIENT_ID
export AZURE_CLIENT_SECRET=YOUR_CLIENT_SECRET
export AZURE_TENANT_ID=YOUR_TENANT_ID
azkeyget
```

## Command Line Options

| Flag | Short | Environment Variable | Description | Required |
|------|-------|---------------------|-------------|----------|
| `--vault-url` | `-v` | `AZURE_KEYVAULT_URL` | Azure Key Vault URL | Yes* |
| `--secret` | `-s` | `AZURE_KEYVAULT_SECRET_NAME` | Name of the secret to retrieve | Yes* |
| `--auth` | `-a` | `AZURE_AUTH_METHOD` | Authentication method: `default`, `system-mi`, `user-mi`, `service-principal` | No (default: `default`) |
| `--client-id` | | `AZURE_CLIENT_ID` | Client ID for service principal or user-assigned managed identity | Conditional |
| `--client-secret` | | `AZURE_CLIENT_SECRET` | Client secret for service principal authentication | Conditional |
| `--tenant-id` | | `AZURE_TENANT_ID` | Tenant ID for service principal authentication | Conditional |
| `--user-assigned-id` | | `AZURE_USER_ASSIGNED_ID` | Alternative to `--client-id` for user-assigned managed identity | No |
| `--debug` | | `AZURE_DEBUG` | Enable debug logging | No |

*Required unless provided via environment variable

## Examples

### Get a database connection string
```bash
# Using CLI flags
DB_CONNECTION=$(azkeyget -v https://myvault.vault.azure.net/ -s database-connection-string)
echo "Connection string retrieved"

# Using environment variables
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_KEYVAULT_SECRET_NAME=database-connection-string
DB_CONNECTION=$(azkeyget)
echo "Connection string retrieved"
```

### Use in a script with error handling

```bash
#!/bin/bash
# Set up environment for the script
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_AUTH_METHOD=system-mi

SECRET=$(azkeyget --secret api-key 2>/dev/null)
if [ $? -eq 0 ]; then
    echo "Secret retrieved successfully"
    # Use $SECRET in your application
else
    echo "Failed to retrieve secret" >&2
    exit 1
fi
```


### CI/CD Pipeline with Service Principal

```bash
# Set in your CI/CD environment
export AZURE_KEYVAULT_URL=https://myvault.vault.azure.net/
export AZURE_AUTH_METHOD=service-principal
export AZURE_CLIENT_ID=${{ secrets.AZURE_CLIENT_ID }}
export AZURE_CLIENT_SECRET=${{ secrets.AZURE_CLIENT_SECRET }}
export AZURE_TENANT_ID=${{ secrets.AZURE_TENANT_ID }}

# Retrieve multiple secrets
API_KEY=$(azkeyget --secret api-key)
DB_PASSWORD=$(azkeyget --secret db-password)
```

## Permissions

The identity used for authentication must have the following Key Vault permissions:
- **Secret permissions**: `Get`

You can assign these permissions through:
- Azure RBAC: `Key Vault Secrets User` role
- Access policies: `Get` permission for secrets

## Error Handling

The tool returns appropriate exit codes:
- `0`: Success
- `1`: Error (authentication failure, secret not found, network issues, etc.)

Error messages are written to stderr, while the secret value is written to stdout.

## Default Azure Credential Chain

When using `--auth default` (or omitting the auth flag), the tool attempts authentication in this order:

1. **Environment variables**: `AZURE_CLIENT_ID`, `AZURE_CLIENT_SECRET`, `AZURE_TENANT_ID`
2. **Workload Identity**: For AKS workload identity
3. **Managed Identity**: System or user-assigned managed identity
4. **Azure CLI**: If logged in via `az login`
5. **Azure PowerShell**: If logged in via Azure PowerShell
6. **Visual Studio**: If logged in via Visual Studio
7. **VS Code**: If logged in via VS Code Azure extension

## Troubleshooting

### Common Issues

**Authentication failed**
- Verify the identity has proper Key Vault permissions
- Check that the authentication method matches your environment
- For service principal auth, verify client ID, secret, and tenant ID

**Secret not found**
- Verify the secret name is correct (case-sensitive)
- Check that the secret exists and is enabled
- Ensure the Key Vault URL is correct

**Network issues**
- Verify connectivity to the Key Vault
- Check firewall rules if using Key Vault network restrictions

### Debug Mode

Enable debug logging to get detailed information about the authentication and secret retrieval process:

```bash
# Using CLI flag
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret --debug

# Using environment variable
export AZURE_DEBUG=true
azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret
```

Debug output includes:

- Configuration details (vault URL, auth method, etc.)
- Authentication method being used
- Credential creation process
- Key Vault client creation
- Secret retrieval steps

**Note**: Debug information is written to stderr, while the secret value is still written to stdout, so you can still capture the secret value while debugging:

```bash
SECRET=$(azkeyget --vault-url https://myvault.vault.azure.net/ --secret mysecret --debug)
# Debug info goes to stderr, secret value is captured in $SECRET
```

## Testing

Run the test suite to ensure everything works correctly:

```bash
# Run all tests
go test -v

# Run only unit tests (faster, no external dependencies)
go test -v -run "^(TestGetEnvOrDefault|TestCreateCredential|TestEnvironmentVariableIntegration)$"

# Run with coverage
go test -v -cover
```

The test suite includes:
- **Unit tests** for environment variable handling and credential creation

## Development

### Development Dependencies

The project uses several development tools that are declared in `tools.go` and managed via `go.mod`:

- **golangci-lint** - Comprehensive Go linter
- **go-critic** - Go source code checker
- **goimports** - Tool to update Go import lines
- **gocyclo** - Cyclomatic complexity analyzer
- **revive** - Fast, configurable, extensible Go linter

### Installing Development Tools

```bash
# Install all development tools
make install-tools

# Or install individually
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/go-critic/go-critic/cmd/gocritic@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/mgechev/revive@latest
```

### Development Workflow

```bash
# Format code
make fmt

# Run linters
make lint

# Run tests
make test

# Run all checks
make check

# Build binary
make build
```

### Pre-commit Hooks

The project uses pre-commit hooks to ensure code quality. Install them with:

```bash
pip install pre-commit
pre-commit install

# Run hooks manually
pre-commit run --all-files
# Or use make
make pre-commit
```

## Contributing

Contributions are welcome! Please ensure all changes:

1. Include appropriate tests
2. Pass all linting checks (`make lint`)
3. Follow Go best practices
4. Include documentation updates if needed

The pre-commit hooks will automatically run formatting, linting, and tests before each commit.

## Project Structure

```
azkeyget/
├── cmd/azkeyget/           # Main application
├── .github/workflows/      # CI/CD workflows
├── Makefile               # Development tasks
├── README.md              # This file
└── LICENSE                # MIT License
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
