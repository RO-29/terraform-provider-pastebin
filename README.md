# Terraform Pastebin Provider

This Terraform provider allows you to manage pastebin pastes as infrastructure resources.

## Features

- **Create pastes** with various options (expiration, formatting, compression, etc.)
- **Read existing pastes** using data sources
- **Support for attachments** and password-protected pastes
- **Burn-after-reading** functionality
- **Discussion/comments** support

## Installation

### From Terraform Registry

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    pastebin = {
      source  = "RO-29/pastebin"
      version = "~> 1.0"
    }
  }
}
```

### From Source (Development)

1. Clone the repository:
   ```bash
   git clone https://github.com/RO-29/terraform-provider-pastebin.git
   cd terraform-provider-pastebin
   ```

2. Build the provider:
   ```bash
   make build
   ```

3. Install locally for development:
   ```bash
   make install-terraform-provider
   ```

## Configuration

The provider supports the following configuration options:

```hcl
provider "pastebin" {
  host              = "https://pastebin.example.tech"  # Required
  username          = var.pastebin_username            # Optional: for authenticated instances
  password          = var.pastebin_password            # Optional: for authenticated instances
  skip_tls_verify   = false                            # Optional: skip TLS verification
  user_agent        = "terraform-provider-pastebin"   # Optional: custom user agent

  # Extra HTTP headers
  extra_headers = {
    "X-Custom-Header" = "value"
  }

  # Default settings for resources
  expire            = "1week"
  formatter         = "plaintext"
  gzip              = true
  open_discussion   = false
  burn_after_reading = false
}
```

### Environment Variables

You can also configure the provider using environment variables:

- `PASTEBIN_HOST` - Pastebin instance host URL
- `PASTEBIN_USERNAME` - Username for authentication
- `PASTEBIN_PASSWORD` - Password for authentication

## Resources

### `pastebin_paste`

Creates and manages a pastebin paste.

```hcl
resource "pastebin_paste" "example" {
  content = "Hello, World!"

  formatter         = "plaintext"
  expire            = "1day"
  password          = "secret123"
  open_discussion   = true
  burn_after_reading = false
  gzip              = true
}
```

#### Arguments

- `content` (Required, String) - The content of the paste
- `attachment_name` (Optional, String) - Name for the attachment (makes the paste an attachment)
- `formatter` (Optional, String) - Text formatter: `plaintext`, `markdown`, `syntaxhighlighting`
- `expire` (Optional, String) - Expiration time: `5min`, `10min`, `1hour`, `1day`, `1week`, `1month`, `1year`, `never`
- `password` (Optional, String, Sensitive) - Password to protect the paste
- `open_discussion` (Optional, Boolean) - Enable discussion/comments on the paste
- `burn_after_reading` (Optional, Boolean) - Delete the paste after first read
- `gzip` (Optional, Boolean) - Enable gzip compression

#### Attributes

- `id` (String) - Paste identifier
- `url` (String) - URL of the created paste
- `delete_token` (String, Sensitive) - Delete token for the paste

## Data Sources

### `pastebin_paste`

Reads an existing pastebin paste.

```hcl
data "pastebin_paste" "existing" {
  url = "https://pastebin.example.tech/?abcd1234#EezApNVTTRUuEkt1jj7r9vSfewLBvUohDSXWuvPEs1bF"

  password     = "secret123"  # Optional: if password protected
  confirm_burn = true         # Optional: confirm reading burn-after-reading pastes
}
```

#### Arguments

- `url` (Required, String) - Full URL of the paste including master key
- `password` (Optional, String, Sensitive) - Password to decrypt the paste
- `confirm_burn` (Optional, Boolean) - Confirm reading a burn-after-reading paste (will delete it)

#### Attributes

- `id` (String) - Paste identifier
- `content` (String) - The content of the paste
- `attachment_name` (String) - Name of the attachment (if paste is an attachment)
- `attachment_data` (String, Sensitive) - Base64 encoded attachment data
- `mime_type` (String) - MIME type of attachment
- `comment_count` (Number) - Number of comments on the paste

## Examples

See the [examples](./examples/) directory for complete usage examples.

### Basic Paste

```hcl
resource "pastebin_paste" "example" {
  content = "Hello, Terraform!"
  expire  = "1day"
}

output "paste_url" {
  value = pastebin_paste.example.url
}
```

### Code with Syntax Highlighting

```hcl
resource "pastebin_paste" "code" {
  content   = file("${path.module}/script.py")
  formatter = "syntaxhighlighting"
  expire    = "1month"
}
```

### Password Protected

```hcl
resource "pastebin_paste" "secret" {
  content  = "Sensitive information"
  password = var.paste_password
  expire   = "1hour"
  burn_after_reading = true
}
```

### Reading Existing Paste

```hcl
data "pastebin_paste" "existing" {
  url = "https://pastebin.example.tech/?id#key"
}

output "content" {
  value = data.pastebin_paste.existing.content
}
```

### Advanced Usage

#### Managing Configuration Files

```hcl
# Store application configuration
resource "pastebin_paste" "app_config" {
  content = jsonencode({
    environment = var.environment
    database_url = var.database_url
    api_keys = {
      stripe = var.stripe_api_key
    }
  })
  password           = var.config_password
  expire             = "1month"
  burn_after_reading = false
  formatter          = "syntaxhighlighting"
}

# Retrieve and use the configuration
data "pastebin_paste" "current_config" {
  url      = pastebin_paste.app_config.url
  password = var.config_password
}

locals {
  config = jsondecode(data.pastebin_paste.current_config.content)
}
```

#### Sharing Build Artifacts

```hcl
# Upload build logs
resource "pastebin_paste" "build_log" {
  content           = file("${path.module}/build.log")
  attachment_name   = "build-${timestamp()}.log"
  expire           = "1week"
  open_discussion  = true
  gzip            = true
}

# Share compressed archives
resource "pastebin_paste" "release_archive" {
  content         = filebase64("${path.module}/dist/app-v${var.version}.tar.gz")
  attachment_name = "app-v${var.version}.tar.gz"
  expire          = "1year"
}
```

## Development

### Prerequisites

- [Go](https://golang.org/doc/install) 1.23+ 
- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.0+

### Building and Testing

1. Clone the repository:
   ```bash
   git clone https://github.com/RO-29/terraform-provider-pastebin.git
   cd terraform-provider-pastebin
   ```

2. Build the provider:
   ```bash
   make build
   ```

3. Install locally for testing:
   ```bash
   make install-terraform-provider
   ```

4. Run tests:
   ```bash
   make test           # Unit tests
   make test-coverage  # With coverage report
   make testacc        # Acceptance tests (requires TF_ACC=1)
   ```

### Testing Your Changes

After making changes, test with a minimal Terraform configuration:

```hcl
terraform {
  required_providers {
    pastebin = {
      source = "RO-29/pastebin"
    }
  }
}

provider "pastebin" {
  host = "https://pastebin.example.tech"
}

resource "pastebin_paste" "test" {
  content = "Test paste from Terraform!"
  expire  = "1day"
}

output "paste_url" {
  value = pastebin_paste.test.url
}
```

## Troubleshooting

### Common Issues

1. **Provider not found**: Ensure you've specified the correct source in your `required_providers` block:
   ```hcl
   terraform {
     required_providers {
       pastebin = {
         source  = "RO-29/pastebin"
         version = "~> 1.0"
       }
     }
   }
   ```

2. **Authentication errors**: Verify your host URL and credentials:
   - Check that `PASTEBIN_HOST` environment variable or `host` provider attribute is set correctly
   - For authenticated instances, ensure `PASTEBIN_USERNAME` and `PASTEBIN_PASSWORD` are set

3. **TLS certificate errors**: For self-hosted pastebin instances with self-signed certificates:
   ```hcl
   provider "pastebin" {
     host            = "https://pastebin.internal.company.com"
     skip_tls_verify = true  # Use only for testing/development
   }
   ```

4. **Build issues**: If you encounter build problems:
   - Ensure Go 1.23+ is installed
   - Check that all dependencies are available: `go mod download`
   - Clean build artifacts: `make clean && make build`

### Getting Help

- Check existing [issues](https://github.com/RO-29/terraform-provider-pastebin/issues) 
- Review the [pastebin-go-cli documentation](https://github.com/RO-29/pastebin-go-cli) for API-related questions
- Create a new issue with:
  - Terraform version (`terraform version`)
  - Provider version
  - Minimal reproduction case
  - Full error messages

## License

This project is licensed under the same terms as the [pastebin-go-cli](https://github.com/RO-29/pastebin-go-cli) library it depends on.
