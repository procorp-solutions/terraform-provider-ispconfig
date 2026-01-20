# Installation Guide

## Prerequisites

- **Go 1.21 or later** (required for building)
- **Terraform 1.0 or later**
- ISP Config 3.x with remote API enabled

## Installing Go 1.21+

If you have an older version of Go (< 1.21), you'll need to upgrade:

### Linux

```bash
# Download and install Go 1.21+ (adjust version as needed)
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz

# Add to PATH (add to ~/.bashrc or ~/.profile for persistence)
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version
```

### macOS

```bash
# Using Homebrew
brew install go@1.21

# Or download from https://go.dev/dl/
```

### Windows

Download and install from https://go.dev/dl/

## Building the Provider

1. Clone the repository:

```bash
git clone https://github.com/procorp-solutions/ispconfig-terraform-provider.git
cd ispconfig-terraform-provider
```

2. Download dependencies:

```bash
go mod download
```

3. Build the provider:

```bash
go build -o terraform-provider-ispconfig
```

## Installing the Provider Locally

### For Development/Testing

Create the appropriate directory structure and copy the binary:

**Linux/macOS:**

```bash
# Determine your architecture
ARCH=$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
OS=$(uname -s | tr '[:upper:]' '[:lower:]')

# Create plugin directory
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/procorp-solutions/ispconfig/1.0.0/${OS}_${ARCH}

# Copy the binary
cp terraform-provider-ispconfig ~/.terraform.d/plugins/registry.terraform.io/procorp-solutions/ispconfig/1.0.0/${OS}_${ARCH}/
```

**Windows (PowerShell):**

```powershell
$ARCH = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$PluginDir = "$env:APPDATA\terraform.d\plugins\registry.terraform.io\procorp-solutions\ispconfig\1.0.0\windows_$ARCH"

New-Item -ItemType Directory -Force -Path $PluginDir
Copy-Item terraform-provider-ispconfig.exe $PluginDir\
```

### Configuring Terraform to Use the Local Provider

Create or update `~/.terraformrc` (Linux/macOS) or `%APPDATA%\terraform.rc` (Windows):

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/procorp-solutions/ispconfig" = "/path/to/terraform-provider-ispconfig"
  }

  # For all other providers, install them directly as normal.
  direct {}
}
```

Replace `/path/to/terraform-provider-ispconfig` with the directory containing the binary (not including the binary name itself).

## Verifying Installation

1. Create a test configuration file `test.tf`:

```hcl
terraform {
  required_providers {
    ispconfig = {
      source = "registry.terraform.io/procorp-solutions/ispconfig"
    }
  }
}

provider "ispconfig" {
  host     = "your-ispconfig-host:8080"
  username = "your-username"
  password = "your-password"
}
```

2. Initialize Terraform:

```bash
terraform init
```

3. If successful, you should see:

```
Terraform has been successfully initialized!
```

## Troubleshooting

### "Provider not found" Error

- Ensure the plugin directory structure is correct
- Check that the binary has execute permissions (Linux/macOS): `chmod +x terraform-provider-ispconfig`
- Verify the provider source in your configuration matches the installation path

### Build Errors

- **Go version too old**: Ensure you have Go 1.21 or later installed (`go version`)
- **Missing dependencies**: Run `go mod download` and try again
- **Compilation errors**: Ensure you're building on a compatible platform

### Runtime Errors

- **TLS/SSL errors**: Try setting `insecure = true` in the provider configuration if using self-signed certificates
- **Authentication errors**: Verify your credentials and that the remote API is enabled in ISP Config
- **Connection timeout**: Check that the ISP Config host is reachable and the port is correct

## Next Steps

See the [README.md](README.md) for usage examples and detailed documentation.

