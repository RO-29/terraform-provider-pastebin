# Terraform Pastebin Provider

This Terraform provider allows you to manage pastebin pastes as infrastructure resources.

## Features

- **Create pastes** with various options (expiration, formatting, compression, etc.)
- **Read existing pastes** using data sources
- **Support for attachments** and password-protected pastes
- **Burn-after-reading** functionality
- **Discussion/comments** support

## Installation

### From Source

1. Build the provider:
   ```bash
   make build-terraform
   ```

2. Install locally for development:
   ```bash
   make install-terraform-provider
   ```

### Using the Provider

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    pastebin = {
      source = "RO-29/pastebin"
    }
  }
}
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

## Development

To contribute to this provider:

1. Build and test:
   ```bash
   make build-terraform
   make install-terraform-provider
   ```

2. Run examples:
   ```bash
   cd terraform/examples
   terraform init
   terraform plan
   ```

## License

This provider is part of the  project and follows the same license terms.
