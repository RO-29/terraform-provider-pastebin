package provider

import (
	"context"
	"crypto/tls"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/RO-29/pastebin-go-cli"
)

// Ensure PastebinProvider satisfies various provider interfaces.
var _ provider.Provider = &PastebinProvider{}

// PastebinProvider defines the provider implementation.
type PastebinProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// PastebinProviderModel describes the provider data model.
type PastebinProviderModel struct {
	Host             types.String `tfsdk:"host"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	SkipTLSVerify    types.Bool   `tfsdk:"skip_tls_verify"`
	UserAgent        types.String `tfsdk:"user_agent"`
	ExtraHeaders     types.Map    `tfsdk:"extra_headers"`
	Expire           types.String `tfsdk:"expire"`
	Formatter        types.String `tfsdk:"formatter"`
	GZip             types.Bool   `tfsdk:"gzip"`
	OpenDiscussion   types.Bool   `tfsdk:"open_discussion"`
	BurnAfterReading types.Bool   `tfsdk:"burn_after_reading"`
}

func (p *PastebinProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "pastebin"
	resp.Version = p.version
}

func (p *PastebinProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Pastebin instance host URL",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for basic authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for basic authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"skip_tls_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification",
				Optional:            true,
			},
			"user_agent": schema.StringAttribute{
				MarkdownDescription: "Custom User-Agent header",
				Optional:            true,
			},
			"extra_headers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Extra HTTP headers to include in requests",
				Optional:            true,
			},
			"expire": schema.StringAttribute{
				MarkdownDescription: "Default expiration time for pastes",
				Optional:            true,
			},
			"formatter": schema.StringAttribute{
				MarkdownDescription: "Default formatter for pastes (plaintext, markdown, syntaxhighlighting)",
				Optional:            true,
			},
			"gzip": schema.BoolAttribute{
				MarkdownDescription: "Enable gzip compression by default",
				Optional:            true,
			},
			"open_discussion": schema.BoolAttribute{
				MarkdownDescription: "Enable discussion on pastes by default",
				Optional:            true,
			},
			"burn_after_reading": schema.BoolAttribute{
				MarkdownDescription: "Enable burn after reading by default",
				Optional:            true,
			},
		},
	}
}

func (p *PastebinProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PastebinProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration values are now available.
	host := os.Getenv("PASTEBIN_HOST")
	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}

	username := os.Getenv("PASTEBIN_USERNAME")
	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	password := os.Getenv("PASTEBIN_PASSWORD")
	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	userAgent := "terraform-provider-pastebin/" + p.version
	if !data.UserAgent.IsNull() {
		userAgent = data.UserAgent.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Unknown Pastebin Host",
			"The provider cannot create the Pastebin API client as there is an unknown configuration value for the Pastebin host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the PASTEBIN_HOST environment variable.",
		)
		return
	}

	hostURL, err := url.Parse(host)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Pastebin Host",
			"The provided host URL is invalid: "+err.Error(),
		)
		return
	}

	// Create client options
	clientOptions := []pastebin.Option{
		pastebin.WithUserAgent(userAgent),
	}

	if username != "" || password != "" {
		clientOptions = append(clientOptions, pastebin.WithBasicAuth(username, password))
	}

	if !data.SkipTLSVerify.IsNull() && data.SkipTLSVerify.ValueBool() {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		clientOptions = append(clientOptions, pastebin.WithTLSConfig(tlsConfig))
	}

	if !data.ExtraHeaders.IsNull() {
		headers := make(map[string]string)
		resp.Diagnostics.Append(data.ExtraHeaders.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}

		for k, v := range headers {
			clientOptions = append(clientOptions, pastebin.WithCustomHeaderField(k, v))
		}
	}

	// Create the client
	client := pastebin.NewClient(*hostURL, clientOptions...)

	// Create provider data struct
	providerData := &ProviderData{
		Client:           client,
		Expire:           data.Expire.ValueString(),
		Formatter:        data.Formatter.ValueString(),
		GZip:             data.GZip.ValueBool(),
		OpenDiscussion:   data.OpenDiscussion.ValueBool(),
		BurnAfterReading: data.BurnAfterReading.ValueBool(),
	}

	// Set defaults if not specified
	if providerData.Expire == "" {
		providerData.Expire = "1week"
	}
	if providerData.Formatter == "" {
		providerData.Formatter = "plaintext"
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PastebinProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPasteResource,
	}
}

func (p *PastebinProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPasteDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PastebinProvider{
			version: version,
		}
	}
}

// ProviderData contains the configured client and default settings
type ProviderData struct {
	Client           *pastebin.Client
	Expire           string
	Formatter        string
	GZip             bool
	OpenDiscussion   bool
	BurnAfterReading bool
}
