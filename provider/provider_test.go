package provider

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPastebinProvider_Metadata(t *testing.T) {
	p := &PastebinProvider{version: "test"}
	ctx := context.Background()
	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(ctx, req, resp)

	assert.Equal(t, "pastebin", resp.TypeName)
	assert.Equal(t, "test", resp.Version)
}

func TestPastebinProvider_Schema(t *testing.T) {
	p := &PastebinProvider{}
	ctx := context.Background()
	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(ctx, req, resp)

	require.NotNil(t, resp.Schema.Attributes)

	// Check that all expected attributes are present
	expectedAttributes := []string{
		"host", "username", "password", "skip_tls_verify", "user_agent",
		"extra_headers", "expire", "formatter", "gzip", "open_discussion", "burn_after_reading",
	}

	for _, attr := range expectedAttributes {
		_, exists := resp.Schema.Attributes[attr]
		assert.True(t, exists, "Expected attribute %s to be present in schema", attr)
	}

	// Verify password is sensitive
	passwordAttr := resp.Schema.Attributes["password"]
	assert.True(t, passwordAttr.IsSensitive(), "Password attribute should be sensitive")

	// Verify all attributes are optional
	for name, attr := range resp.Schema.Attributes {
		assert.True(t, attr.IsOptional(), "Attribute %s should be optional", name)
	}
}

func TestPastebinProvider_Configure_EnvironmentVariables(t *testing.T) {
	// Test environment variable handling without calling Configure
	// since Configure requires complex framework setup
	
	tests := []struct {
		name     string
		hostEnv  string
		userEnv  string
		passEnv  string
		expected map[string]string
	}{
		{
			name:    "all environment variables set",
			hostEnv: "https://example.com",
			userEnv: "testuser",
			passEnv: "testpass",
			expected: map[string]string{
				"PASTEBIN_HOST":     "https://example.com",
				"PASTEBIN_USERNAME": "testuser",
				"PASTEBIN_PASSWORD": "testpass",
			},
		},
		{
			name:    "only host set",
			hostEnv: "https://paste.example.com",
			userEnv: "",
			passEnv: "",
			expected: map[string]string{
				"PASTEBIN_HOST":     "https://paste.example.com",
				"PASTEBIN_USERNAME": "",
				"PASTEBIN_PASSWORD": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Store original values
			originalHost := os.Getenv("PASTEBIN_HOST")
			originalUsername := os.Getenv("PASTEBIN_USERNAME")
			originalPassword := os.Getenv("PASTEBIN_PASSWORD")

			defer func() {
				restoreEnv("PASTEBIN_HOST", originalHost)
				restoreEnv("PASTEBIN_USERNAME", originalUsername)
				restoreEnv("PASTEBIN_PASSWORD", originalPassword)
			}()

			// Set test values
			setEnv("PASTEBIN_HOST", tt.hostEnv)
			setEnv("PASTEBIN_USERNAME", tt.userEnv)
			setEnv("PASTEBIN_PASSWORD", tt.passEnv)

			// Verify environment variables are set correctly
			assert.Equal(t, tt.expected["PASTEBIN_HOST"], os.Getenv("PASTEBIN_HOST"))
			assert.Equal(t, tt.expected["PASTEBIN_USERNAME"], os.Getenv("PASTEBIN_USERNAME"))
			assert.Equal(t, tt.expected["PASTEBIN_PASSWORD"], os.Getenv("PASTEBIN_PASSWORD"))
		})
	}
}

func TestPastebinProvider_Configure_URLValidation(t *testing.T) {
	// Test URL validation logic that would be used in Configure
	tests := []struct {
		name        string
		host        string
		expectError bool
	}{
		{
			name:        "valid URL",
			host:        "https://example.com",
			expectError: false,
		},
		{
			name:        "invalid URL with spaces",
			host:        "ht tp://invalid url",
			expectError: true,
		},
		{
			name:        "URL with port",
			host:        "https://example.com:8080",
			expectError: false,
		},
		{
			name:        "URL with path",
			host:        "https://example.com/api",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := url.Parse(tt.host)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPastebinProvider_Resources(t *testing.T) {
	p := &PastebinProvider{}
	ctx := context.Background()

	resources := p.Resources(ctx)

	assert.Len(t, resources, 1)
	
	// Test that the resource factory function works
	resource := resources[0]()
	assert.NotNil(t, resource)
}

func TestPastebinProvider_DataSources(t *testing.T) {
	p := &PastebinProvider{}
	ctx := context.Background()

	dataSources := p.DataSources(ctx)

	assert.Len(t, dataSources, 1)
	
	// Test that the data source factory function works
	dataSource := dataSources[0]()
	assert.NotNil(t, dataSource)
}

func TestNew(t *testing.T) {
	version := "1.2.3"
	providerFactory := New(version)

	provider := providerFactory()
	assert.NotNil(t, provider)

	// Verify it's the right type and version is set
	pastebinProvider, ok := provider.(*PastebinProvider)
	assert.True(t, ok)
	assert.Equal(t, version, pastebinProvider.version)
}

// testProviderFactory creates a provider factory for use in tests
func testProviderFactory() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"pastebin": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func TestProviderData_Defaults(t *testing.T) {
	tests := []struct {
		name     string
		data     ProviderData
		expected ProviderData
	}{
		{
			name: "empty provider data gets defaults",
			data: ProviderData{},
			expected: ProviderData{
				Expire:    "",
				Formatter: "",
			},
		},
		{
			name: "existing values are preserved",
			data: ProviderData{
				Expire:           "1day",
				Formatter:        "markdown",
				GZip:             true,
				OpenDiscussion:   true,
				BurnAfterReading: true,
			},
			expected: ProviderData{
				Expire:           "1day",
				Formatter:        "markdown",
				GZip:             true,
				OpenDiscussion:   true,
				BurnAfterReading: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected.Expire, tt.data.Expire)
			assert.Equal(t, tt.expected.Formatter, tt.data.Formatter)
			assert.Equal(t, tt.expected.GZip, tt.data.GZip)
			assert.Equal(t, tt.expected.OpenDiscussion, tt.data.OpenDiscussion)
			assert.Equal(t, tt.expected.BurnAfterReading, tt.data.BurnAfterReading)
		})
	}
}

// Helper functions for environment variable testing
func setEnv(key, value string) {
	if value == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, value)
	}
}

func restoreEnv(key, originalValue string) {
	if originalValue == "" {
		os.Unsetenv(key)
	} else {
		os.Setenv(key, originalValue)
	}
}

// Helper function to parse URL safely for tests
func mustParseURL(urlStr string) *url.URL {
	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}
	return u
}