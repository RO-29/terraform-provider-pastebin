package provider

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RO-29/pastebin-go-cli"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &PasteDataSource{}

func NewPasteDataSource() datasource.DataSource {
	return &PasteDataSource{}
}

// PasteDataSource defines the data source implementation.
type PasteDataSource struct {
	providerData *ProviderData
}

// PasteDataSourceModel describes the data source data model.
type PasteDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	URL            types.String `tfsdk:"url"`
	Password       types.String `tfsdk:"password"`
	ConfirmBurn    types.Bool   `tfsdk:"confirm_burn"`
	Content        types.String `tfsdk:"content"`
	AttachmentName types.String `tfsdk:"attachment_name"`
	AttachmentData types.String `tfsdk:"attachment_data"`
	MimeType       types.String `tfsdk:"mime_type"`
	CommentCount   types.Int64  `tfsdk:"comment_count"`
}

func (d *PasteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_paste"
}

func (d *PasteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pastebin paste data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Paste identifier (computed from URL)",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Full URL of the paste including master key",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to decrypt the paste (if password protected)",
				Optional:            true,
				Sensitive:           true,
			},
			"confirm_burn": schema.BoolAttribute{
				MarkdownDescription: "Confirm reading a burn-after-reading paste (will delete it)",
				Optional:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the paste",
				Computed:            true,
			},
			"attachment_name": schema.StringAttribute{
				MarkdownDescription: "Name of the attachment (if paste is an attachment)",
				Computed:            true,
			},
			"attachment_data": schema.StringAttribute{
				MarkdownDescription: "Base64 encoded attachment data (if paste is an attachment)",
				Computed:            true,
				Sensitive:           true,
			},
			"mime_type": schema.StringAttribute{
				MarkdownDescription: "MIME type of attachment (if paste is an attachment)",
				Computed:            true,
			},
			"comment_count": schema.Int64Attribute{
				MarkdownDescription: "Number of comments on the paste",
				Computed:            true,
			},
		},
	}
}

func (d *PasteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.providerData = providerData
}

func (d *PasteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PasteDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the paste URL
	pasteURL, err := url.Parse(data.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse paste URL: %s", err))
		return
	}

	// Prepare options
	password := []byte(data.Password.ValueString())
	confirmBurn := data.ConfirmBurn.ValueBool()

	options := pastebin.ShowPasteOptions{
		Password:    password,
		ConfirmBurn: confirmBurn,
	}

	// Read the paste
	result, err := d.providerData.Client.ShowPaste(ctx, *pasteURL, options)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read paste: %s", err))
		return
	}

	// Map response to data source model
	data.ID = types.StringValue(result.PasteID)
	data.Content = types.StringValue(string(result.Paste.Data))
	data.CommentCount = types.Int64Value(int64(result.CommentCount))

	// Handle attachment data if present
	if result.Paste.AttachmentName != "" {
		data.AttachmentName = types.StringValue(result.Paste.AttachmentName)
		data.MimeType = types.StringValue(result.Paste.MimeType)

		// Convert attachment to base64 string
		if len(result.Paste.Attachement) > 0 {
			data.AttachmentData = types.StringValue(base64.StdEncoding.EncodeToString(result.Paste.Attachement))
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
