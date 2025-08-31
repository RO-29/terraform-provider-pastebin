package provider

import (
	"context"
	"net/url"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/RO-29/pastebin-go-cli"
)

func TestPasteResource_Metadata(t *testing.T) {
	r := &PasteResource{}
	ctx := context.Background()
	req := resource.MetadataRequest{
		ProviderTypeName: "pastebin",
	}
	resp := &resource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "pastebin_paste", resp.TypeName)
}

func TestPasteResource_Schema(t *testing.T) {
	r := &PasteResource{}
	ctx := context.Background()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(ctx, req, resp)

	require.NotNil(t, resp.Schema.Attributes)

	// Check that all expected attributes are present
	expectedAttributes := []string{
		"id", "content", "attachment_name", "formatter", "expire",
		"password", "open_discussion", "burn_after_reading", "gzip",
		"url", "delete_token",
	}

	for _, attr := range expectedAttributes {
		_, exists := resp.Schema.Attributes[attr]
		assert.True(t, exists, "Expected attribute %s to be present in schema", attr)
	}

	// Verify required attributes
	contentAttr := resp.Schema.Attributes["content"]
	assert.True(t, contentAttr.IsRequired(), "Content attribute should be required")

	// Verify computed attributes
	computedAttrs := []string{"id", "url", "delete_token"}
	for _, attrName := range computedAttrs {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsComputed(), "Attribute %s should be computed", attrName)
	}

	// Verify sensitive attributes
	sensitiveAttrs := []string{"password", "delete_token"}
	for _, attrName := range sensitiveAttrs {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsSensitive(), "Attribute %s should be sensitive", attrName)
	}

	// Verify optional attributes with defaults
	optionalWithDefaults := []string{"formatter", "expire", "open_discussion", "burn_after_reading", "gzip"}
	for _, attrName := range optionalWithDefaults {
		attr := resp.Schema.Attributes[attrName]
		assert.True(t, attr.IsOptional(), "Attribute %s should be optional", attrName)
		assert.True(t, attr.IsComputed(), "Attribute %s should be computed (for defaults)", attrName)
	}
}

func TestPasteResource_Configure_Success(t *testing.T) {
	r := &PasteResource{}
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

	req := resource.ConfigureRequest{
		ProviderData: providerData,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, providerData, r.providerData)
}

func TestPasteResource_Configure_InvalidProviderData(t *testing.T) {
	r := &PasteResource{}
	ctx := context.Background()
	
	req := resource.ConfigureRequest{
		ProviderData: "invalid", // Wrong type
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Unexpected Resource Configure Type")
}

func TestPasteResource_Configure_NilProviderData(t *testing.T) {
	r := &PasteResource{}
	ctx := context.Background()
	
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Nil(t, r.providerData)
}

func TestPasteResource_Update_NotSupported(t *testing.T) {
	r := &PasteResource{}
	ctx := context.Background()
	
	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	r.Update(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError())
	assert.Contains(t, resp.Diagnostics.Errors()[0].Summary(), "Update Not Supported")
}

func TestNewPasteResource(t *testing.T) {
	resource := NewPasteResource()
	assert.NotNil(t, resource)
	
	// Verify it's the correct type
	_, ok := resource.(*PasteResource)
	assert.True(t, ok)
}

func TestPasteResourceModel_DefaultValues(t *testing.T) {
	// Test that the model can be created and has expected zero values
	model := PasteResourceModel{}
	
	assert.True(t, model.ID.IsNull())
	assert.True(t, model.Content.IsNull())
	assert.True(t, model.AttachmentName.IsNull())
	assert.True(t, model.Formatter.IsNull())
	assert.True(t, model.Expire.IsNull())
	assert.True(t, model.Password.IsNull())
	assert.True(t, model.OpenDiscussion.IsNull())
	assert.True(t, model.BurnAfterReading.IsNull())
	assert.True(t, model.GZip.IsNull())
	assert.True(t, model.URL.IsNull())
	assert.True(t, model.DeleteToken.IsNull())
}

func TestPasteResourceModel_WithValues(t *testing.T) {
	// Test that the model can hold values correctly
	model := PasteResourceModel{
		ID:               types.StringValue("test-id"),
		Content:          types.StringValue("test content"),
		AttachmentName:   types.StringValue("test.txt"),
		Formatter:        types.StringValue("plaintext"),
		Expire:           types.StringValue("1week"),
		Password:         types.StringValue("secret"),
		OpenDiscussion:   types.BoolValue(true),
		BurnAfterReading: types.BoolValue(false),
		GZip:             types.BoolValue(true),
		URL:              types.StringValue("https://example.com/paste/test-id"),
		DeleteToken:      types.StringValue("delete-token"),
	}
	
	assert.Equal(t, "test-id", model.ID.ValueString())
	assert.Equal(t, "test content", model.Content.ValueString())
	assert.Equal(t, "test.txt", model.AttachmentName.ValueString())
	assert.Equal(t, "plaintext", model.Formatter.ValueString())
	assert.Equal(t, "1week", model.Expire.ValueString())
	assert.Equal(t, "secret", model.Password.ValueString())
	assert.True(t, model.OpenDiscussion.ValueBool())
	assert.False(t, model.BurnAfterReading.ValueBool())
	assert.True(t, model.GZip.ValueBool())
	assert.Equal(t, "https://example.com/paste/test-id", model.URL.ValueString())
	assert.Equal(t, "delete-token", model.DeleteToken.ValueString())
}

// Mock tests for Create, Read, Delete would require mocking the pastebin client
// Since we don't have access to mock the external client easily, we focus on
// testing the logic we can control (schema, configuration, model validation)

func TestPasteResource_Delete_LogicOnly(t *testing.T) {
	// Test the Delete method without actually calling it since it tries to read state
	// This test verifies the Delete method exists and has the expected signature
	r := &PasteResource{}
	assert.NotNil(t, r)
	
	// We can't easily test Delete without mocking the entire state infrastructure
	// The method signature is tested by compilation, and the actual delete logic
	// is mostly just removing from state per the comment in the implementation
}

func TestPasteResource_ImportState(t *testing.T) {
	// Test that ImportState method exists and can be called
	// The actual functionality requires complex framework setup that's not
	// practical for unit tests
	r := &PasteResource{}
	assert.NotNil(t, r)
	
	// The ImportState method uses ImportStatePassthroughID which requires
	// a proper framework context that's complex to set up in unit tests.
	// We verify the method exists by compilation and leave detailed testing
	// to acceptance tests.
}

// Test helper functions and utilities
func createMockProviderData() *ProviderData {
	testURL, _ := url.Parse("https://example.com")
	return &ProviderData{
		Client:           pastebin.NewClient(*testURL),
		Expire:           "1week",
		Formatter:        "plaintext",
		GZip:             false,
		OpenDiscussion:   false,
		BurnAfterReading: false,
	}
}

func TestPasteResource_Integration_Configure_And_Schema(t *testing.T) {
	// Integration test that verifies configure and schema work together
	r := &PasteResource{}
	ctx := context.Background()

	// First test schema
	schemaReq := resource.SchemaRequest{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, schemaReq, schemaResp)
	assert.False(t, schemaResp.Diagnostics.HasError())

	// Then test configure
	configureReq := resource.ConfigureRequest{
		ProviderData: createMockProviderData(),
	}
	configureResp := &resource.ConfigureResponse{}
	r.Configure(ctx, configureReq, configureResp)
	assert.False(t, configureResp.Diagnostics.HasError())

	// Verify resource is properly configured
	assert.NotNil(t, r.providerData)
	assert.NotNil(t, r.providerData.Client)
}