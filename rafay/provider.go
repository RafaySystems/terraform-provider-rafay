package rafay

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"strings"

	rctlconfig "github.com/RafaySystems/rctl/pkg/config"
	rctlcontext "github.com/RafaySystems/rctl/pkg/context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func New(_ string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"provider_config_file": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("RAFAY_PROVIDER_CONFIG", "~/.rafay/cli/config.json"),
				},
				"ignore_insecure_tls_error": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"rafay_project":          resourceProject(),
				"rafay_group":            resourceGroup(),
				"rafay_groupassociation": resourceGroupAssociation(),
				"rafay_cloud_credential": resourceCloudCredential(),
				"rafay_eks_cluster":      resourceEKSCluster(),
				"rafay_addon":            resourceAddon(),
				"rafay_blueprint":        resourceBluePrint(),
			},
			DataSourcesMap:       map[string]*schema.Resource{},
			ConfigureContextFunc: providerConfigure,
		}

		return p
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

func providerConfigure(ctx context.Context, rd *schema.ResourceData) (interface{}, diag.Diagnostics) {
	var diags diag.Diagnostics

	config_file := rd.Get("provider_config_file").(string)
	ignoreTlsError := rd.Get("ignore_insecure_tls_error").(bool)

	log.Printf("rafay provider config file %s", config_file)
	cliCtx := rctlcontext.GetContext()
	if config_file != "" {
		var err error

		config_file = strings.TrimSpace(config_file)
		if config_file[0] == '~' {
			config_file, err = expandHomeDir(config_file)
		} else {
			config_file, err = filepath.Abs(config_file)
		}

		if err == nil {
			log.Printf("rafay provider config file absolute path %s", config_file)
			configPath := filepath.Dir(config_file)
			fileName := filepath.Base(config_file)
			cliCtx.ConfigFile = fileName
			cliCtx.ConfigDir = configPath
		} else {
			log.Println("failed to get rafay provider config absolute path error:", err)
			log.Println("provider will use default config file ~/.rafay/cli/config.json")
		}
	} else {
		log.Println("provider will use default config file ~/.rafay/cli/config.json")
	}

	err := rctlconfig.InitConfig(cliCtx)

	if err != nil {
		log.Printf("rafay provider config init error %s", err.Error())
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create rafay provider",
			Detail:   "Unable to init config for authenticated rafay provider",
		})
		return nil, diags
	}

	if ignoreTlsError {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	return rctlconfig.GetConfig(), diags

}
