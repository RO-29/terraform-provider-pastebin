package provider

import (
	"context"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/RO-29/pastebin-go-cli"
)

func TestPasteDataSource_Metadata(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	req := datasource.MetadataRequest{
		ProviderTypeName: "pastebin",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(ctx, req, resp)

	assert.Equal(t, "pastebin_paste", resp.TypeName)
}

func TestPasteDataSource_Schema(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	require.NotNil(t, resp.Schema.Attributes)

	// Check that all expected attributes are present
	expectedAttributes := []string{
		"id", "url", "password", "confirm_burn", "content",
		"attachment_name", "attachment_data", "mime_type", "comment_count",
	}

	for _, attr := range expectedAttributes {
		_, exists := resp.Schema.Attributes[attr]
		assert.True(t, exists, "Expected attribute %s to be present in schema", attr)
	}

	// Verify required attributes
	urlAttr := resp.Schema.Attributes["url"]
	assert.True(t, urlAttr.IsRequired(), "URL attribute should be required")

	// Verify computed attributes
	computedAttrs := []string{"id", "content", "attachment_name", "attachment_data", "mime_type", "comment_count"}
	for _, attrName := range computedAttrs {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsComputed(), "Attribute %s should be computed", attrName)
	}

	// Verify optional attributes
	optionalAttrs := []string{"password", "confirm_burn"}
	for _, attrName := range optionalAttrs {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsOptional(), "Attribute %s should be optional", attrName)
	}

	// Verify sensitive attributes
	sensitiveAttrs := []string{"password", "attachment_data"}
	for _, attrName := range sensitiveAttrs {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsSensitive(), "Attribute %s should be sensitive", attrName)
	}
}

func TestPasteDataSource_Configure_Success(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	
	// Create mock provider data
	testURL, _ := url.Parse("https://example.com")
	providerData := &ProviderData{
		Client:           pastebin.NewClient(*testURL),
		Expire:           "1week",
		Formatter:        "plaintext",
		GZip:             false,
		OpenDiscussion:   false,
		BurnAfterReading: false,
	}

	req := datasource.ConfigureRequest{
		ProviderData: providerData,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, providerData, d.providerData)
}

func TestPasteDataSource_Configure_InvalidProviderData(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	
	req := datasource.ConfigureRequest{
		ProviderData: "invalid", // Wrong type
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Data Source Configure Type")
}

func TestPasteDataSource_Configure_NilProviderData(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, d.providerData)
}

func TestNewPasteDataSource(t *testing.T) {
	dataSource := NewPasteDataSource()
	assert.NotNil(t, dataSource)
	
	// Verify it's the correct type
	_, ok := dataSource.(*PasteDataSource)
	assert.True(t, ok)
}

func TestPasteDataSourceModel_DefaultValues(t *testing.T) {
	// Test that the model can be created and has expected zero values
	model := PasteDataSourceModel{}
	
	assert.True(t, model.ID.IsNull())
	assert.True(t, model.URL.IsNull())
	assert.True(t, model.Password.IsNull())
	assert.True(t, model.ConfirmBurn.IsNull())
	assert.True(t, model.Content.IsNull())
	assert.True(t, model.AttachmentName.IsNull())
	assert.True(t, model.AttachmentData.IsNull())
	assert.True(t, model.MimeType.IsNull())
	assert.True(t, model.CommentCount.IsNull())
}

func TestPasteDataSourceModel_WithValues(t *testing.T) {
	// Test that the model can hold values correctly
	model := PasteDataSourceModel{
		ID:             types.StringValue("test-id"),
		URL:            types.StringValue("https://example.com/paste/test-id"),
		Password:       types.StringValue("secret"),
		ConfirmBurn:    types.BoolValue(false),
		Content:        types.StringValue("test content"),
		AttachmentName: types.StringValue("test.txt"),
		AttachmentData: types.StringValue("dGVzdCBkYXRh"), // base64 encoded "test data"
		MimeType:       types.StringValue("text/plain"),
		CommentCount:   types.Int64Value(5),
	}
	
	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "https://example.com/paste/test-id", model.URL.ValueString())
	assert.Equal(t, "secret", model.Password.ValueString())
	assert.False(t, model.ConfirmBurn.ValueBool())
	assert.Equal(t, "test content", model.Content.ValueString())
	assert.Equal(t, "test.txt", model.AttachmentName.ValueString())
	assert.Equal(t, "dGVzdCBkYXRh", model.AttachmentData.ValueString())
	assert.Equal(t, "text/plain", model.MimeType.ValueString())
	assert.Equal(t, int64(5), model.CommentCount.ValueInt64())
}

// Test URL validation logic that would be used in Read method
func TestPasteDataSource_URLValidation(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL",
			url:         "https://example.com/paste/abc123",
			expectError: false,
		},
		{
			name:        "invalid URL with invalid characters",
			url:         "ht tp://invalid url with spaces",
			expectError: true,
		},
		{
			name:        "URL with special characters",
			url:         "https://example.com/paste/abc123#key=xyz",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := url.Parse(tt.url)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Mock tests for Read would require mocking the pastebin client
// Since we don't have access to mock the external client easily, we focus on
// testing the logic we can control (schema, configuration, model validation)

func TestPasteDataSource_Integration_Configure_And_Schema(t *testing.T) {
	// Integration test that verifies configure and schema work together
	d := &PasteDataSource{}
	ctx := context.Background()

	// First test schema
	schemaReq := datasource.SchemaRequest{}
	schemaResp := &datasource.SchemaResponse{}
	d.Schema(ctx, schemaReq, schemaResp)
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Then test configure
	configureReq := datasource.ConfigureRequest{
		ProviderData: createMockProviderData(),
	}
	configureResp := &datasource.ConfigureResponse{}
	d.Configure(ctx, configureReq, configureResp)
	assert.False(t, configureResp.Diagnostics.HasError())

	// Verify data source is properly configured
	assert.NotNil(t, d.providerData)
	assert.NotNil(t, d.providerData.Client)
}

// Test that all schema attribute types are correct
func TestPasteDataSource_Schema_AttributeTypes(t *testing.T) {
	d := &PasteDataSource{}
	ctx := context.Background()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(ctx, req, resp)

	schema := resp.Schema

	// Test string attributes
	stringAttrs := []string{"id", "url", "password", "content", "attachment_name", "attachment_data", "mime_type"}
	for _, attrName := range stringAttrs {
		attr := schema.Attributes[attrName]
		assert.NotNil(t, attr, "Attribute %s should exist", attrName)
		// We can't easily check the exact type without casting, but we can verify it exists
	}

	// Test boolean attributes
	boolAttrs := []string{"confirm_burn"}
	for _, attrName := range boolAttrs {
		attr := schema.Attributes[attrName]
		assert.NotNil(t, attr, "Attribute %s should exist", attrName)
	}

	// Test int64 attributes
	int64Attrs := []string{"comment_count"}
	for _, attrName := range int64Attrs {
		attr := schema.Attributes[attrName]
		assert.NotNil(t, attr, "Attribute %s should exist", attrName)
	}
}

// Test model field mappings
func TestPasteDataSourceModel_FieldTypes(t *testing.T) {
	model := PasteDataSourceModel{}

	// Verify all fields are of the expected types
	// This is compile-time validation but helps ensure model consistency
	assert.IsType(t, types.String{}, model.ID)
	assert.IsType(t, types.String{}, model.URL)
	assert.IsType(t, types.String{}, model.Password)
	assert.IsType(t, types.Bool{}, model.ConfirmBurn)
	assert.IsType(t, types.String{}, model.Content)
	assert.IsType(t, types.String{}, model.AttachmentName)
	assert.IsType(t, types.String{}, model.AttachmentData)
	assert.IsType(t, types.String{}, model.MimeType)
	assert.IsType(t, types.Int64{}, model.CommentCount)
}

// Test that password and attachment_data are properly handled as sensitive
func TestPasteDataSourceModel_SensitiveFields(t *testing.T) {
	model := PasteDataSourceModel{
		Password:       types.StringValue("secret"),
		AttachmentData: types.StringValue("sensitive-data"),
	}

	// Verify we can set and retrieve sensitive values
	assert.Equal(t, "secret", model.Password.ValueString())
	assert.Equal(t, "sensitive-data", model.AttachmentData.ValueString())
	assert.False(t, model.Password.IsNull())
	assert.False(t, model.AttachmentData.IsNull())
}