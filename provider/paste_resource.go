package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RO-29/pastebin-go-cli"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PasteResource{}
var _ resource.ResourceWithImportState = &PasteResource{}

func NewPasteResource() resource.Resource {
	return &PasteResource{}
}

// PasteResource defines the resource implementation.
type PasteResource struct {
	providerData *ProviderData
}

// PasteResourceModel describes the resource data model.
type PasteResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Content          types.String `tfsdk:"content"`
	AttachmentName   types.String `tfsdk:"attachment_name"`
	Formatter        types.String `tfsdk:"formatter"`
	Expire           types.String `tfsdk:"expire"`
	Password         types.String `tfsdk:"password"`
	OpenDiscussion   types.Bool   `tfsdk:"open_discussion"`
	BurnAfterReading types.Bool   `tfsdk:"burn_after_reading"`
	GZip             types.Bool   `tfsdk:"gzip"`
	URL              types.String `tfsdk:"url"`
	DeleteToken      types.String `tfsdk:"delete_token"`
}

func (r *PasteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_paste"
}

func (r *PasteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Pastebin paste resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Paste identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the paste",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attachment_name": schema.StringAttribute{
				MarkdownDescription: "Name for the attachment (makes the paste an attachment)",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"formatter": schema.StringAttribute{
				MarkdownDescription: "Text formatter (plaintext, markdown, syntaxhighlighting)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("plaintext"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expire": schema.StringAttribute{
				MarkdownDescription: "Expiration time (5min, 10min, 1hour, 1day, 1week, 1month, 1year, never)",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("1week"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to protect the paste",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"open_discussion": schema.BoolAttribute{
				MarkdownDescription: "Enable discussion/comments on the paste",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"burn_after_reading": schema.BoolAttribute{
				MarkdownDescription: "Delete the paste after first read",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"gzip": schema.BoolAttribute{
				MarkdownDescription: "Enable gzip compression",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "URL of the created paste",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delete_token": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "Delete token for the paste",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *PasteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(*ProviderData)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.providerData = providerData
}

func (r *PasteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PasteResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider defaults if not specified
	formatter := data.Formatter.ValueString()
	if formatter == "" {
		formatter = r.providerData.Formatter
	}

	expire := data.Expire.ValueString()
	if expire == "" {
		expire = r.providerData.Expire
	}

	gzip := data.GZip.ValueBool()
	if data.GZip.IsNull() {
		gzip = r.providerData.GZip
	}

	openDiscussion := data.OpenDiscussion.ValueBool()
	if data.OpenDiscussion.IsNull() {
		openDiscussion = r.providerData.OpenDiscussion
	}

	burnAfterReading := data.BurnAfterReading.ValueBool()
	if data.BurnAfterReading.IsNull() {
		burnAfterReading = r.providerData.BurnAfterReading
	}

	// Prepare paste options
	compress := pastebin.CompressionAlgorithmNone
	if gzip {
		compress = pastebin.CompressionAlgorithmGZip
	}

	password := []byte(data.Password.ValueString())

	options := pastebin.CreatePasteOptions{
		AttachmentName:   data.AttachmentName.ValueString(),
		Formatter:        formatter,
		Expire:           expire,
		OpenDiscussion:   openDiscussion,
		BurnAfterReading: burnAfterReading,
		Compress:         compress,
		Password:         password,
	}

	// Create the paste
	result, err := r.providerData.Client.CreatePaste(ctx, []byte(data.Content.ValueString()), options)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create paste, got error: %s", err))
		return
	}

	// Save data into Terraform state
	data.ID = types.StringValue(result.PasteID)
	data.URL = types.StringValue(result.PasteURL.String())
	data.DeleteToken = types.StringValue(result.DeleteToken)

	// Set computed values based on what was actually used
	data.Formatter = types.StringValue(formatter)
	data.Expire = types.StringValue(expire)
	data.GZip = types.BoolValue(gzip)
	data.OpenDiscussion = types.BoolValue(openDiscussion)
	data.BurnAfterReading = types.BoolValue(burnAfterReading)

	// Write logs using the tflog package
	// tflog.Trace(ctx, "created a paste resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PasteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PasteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the URL to check if paste still exists
	pasteURL, err := url.Parse(data.URL.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse paste URL: %s", err))
		return
	}

	// Try to read the paste (this will fail if it doesn't exist or was burned)
	options := pastebin.ShowPasteOptions{
		Password:    []byte(data.Password.ValueString()),
		ConfirmBurn: false, // Don't actually read burn-after-reading pastes
	}

	_, err = r.providerData.Client.ShowPaste(ctx, *pasteURL, options)
	if err != nil {
		// If we can't read the paste, it might have been deleted or burned
		// Remove from state
		resp.State.RemoveResource(ctx)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PasteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Pastes are immutable, so any changes require replacement
	// This should not be called due to RequiresReplace plan modifiers
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"Paste resources are immutable and cannot be updated. Any changes require replacement.",
	)
}

func (r *PasteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PasteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Note: The pastebin API doesn't support deleting pastes via delete token in this implementation
	// In a real implementation, you would use the delete token to delete the paste
	// For now, we'll just remove it from state

	// If you had a delete API:
	// err := r.providerData.Client.DeletePaste(ctx, data.ID.ValueString(), data.DeleteToken.ValueString())
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete paste, got error: %s", err))
	//     return
	// }
}

func (r *PasteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
