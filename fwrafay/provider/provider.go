package provider

import (
	"context"
	"crypto/tls"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	config "github.com/RafaySystems/rctl/pkg/config"
	rctlcontext "github.com/RafaySystems/rctl/pkg/context"

	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure RafayFwProvider satisfies terraform framework provider interfaces.
var _ provider.Provider = &RafayFwProvider{}

const TF_USER_AGENT = "terraform"

// RafayFwProvider defines the provider implementation using framework.
type RafayFwProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RafayFwProvider{
			version: version,
		}
	}
}

type RafayFwProviderModel struct {
	ProviderConfigFile     types.String        `tfsdk:"provider_config_file"`
	IgnoreInsecureTlsError basetypes.BoolValue `tfsdk:"ignore_insecure_tls_error"`
}

func (p *RafayFwProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"provider_config_file": schema.StringAttribute{
				Description: "Path to the rafay provider config file",
				Optional:    true,
			},
			"ignore_insecure_tls_error": schema.BoolAttribute{
				Optional:    true,
				Description: "Ignore insecure tls error",
			},
		},
	}

}

func (p *RafayFwProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "rafay"
	resp.Version = p.version
}

func (p *RafayFwProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {

	var data RafayFwProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	configFile := data.ProviderConfigFile.ValueString()
	ignoreTlsError := data.IgnoreInsecureTlsError

	tflog.Info(ctx, "rafay provider config file", map[string]interface{}{
		"config_file": configFile,
	})

	cliCtx := rctlcontext.GetContext()

	if configFile != "" {
		var err error

		configFile = strings.TrimSpace(configFile)
		if configFile[0] == '~' {
			configFile, err = expandHomeDir(configFile)
		} else {
			configFile, err = filepath.Abs(configFile)
		}

		if err == nil {
			tflog.Info(ctx, "rafay provider config file absolute path", map[string]interface{}{
				"config_file": configFile,
			})
			configPath := filepath.Dir(configFile)
			fileName := filepath.Base(configFile)
			cliCtx.ConfigFile = fileName
			cliCtx.ConfigDir = configPath
		} else {
			tflog.Error(ctx, "failed to get rafay provider config absolute path", map[string]interface{}{
				"error": err,
			})
			tflog.Info(ctx, "provider will use default config file ~/.rafay/cli/config.json")
		}
	} else {
		tflog.Info(ctx, "provider will use default config file ~/.rafay/cli/config.json")
	}

	err := config.InitConfig(cliCtx)

	if err != nil {
		tflog.Error(ctx, "rafay provider config init error", map[string]interface{}{
			"error": err.Error(),
		})
		resp.Diagnostics.AddError(
			"Unable to create rafay provider",
			"Unable to init config for authenticated rafay provider: "+err.Error(),
		)
		return
	}

	if ignoreTlsError.ValueBool() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())

	if err != nil {
		resp.Diagnostics.AddError("Unable to initialise the Client, Error", err.Error())
		return
	}
	// Save the client in the provider data
	resp.ResourceData = client
	resp.DataSourceData = client

}

func (p *RafayFwProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *RafayFwProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMksClusterResource,
	}
}

func expandHomeDir(path string) (string, error) {
	if len(path) == 0 || path[0] != '~' {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
