package rafay

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataCloudCredentialV3() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataCloudCredentialV3Read,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"projectname": {
				Description: "Project name from where blueprints to be listed",
				Type:        schema.TypeString,
				Required:    true,
			},
			"credentials_list": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of cloud credentials",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metadata": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Metadata of the credential",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"spec": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Spec of the credential",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"credentials": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Credentials of the provider",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataCloudCredentialV3Read(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataCloudCredentialV3Read ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	project := d.Get("projectname").(string)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	credential, err := client.InfraV3().Credentials().List(ctx, options.ListOptions{
		Project: project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	credentialsList := make([]interface{}, len(credential.Items))
	for i, cred := range credential.Items {
		credentialMap := make(map[string]interface{})
		metadataMap := make(map[string]interface{})
		specMap := make(map[string]interface{})
		credentialsMap := make(map[string]interface{})

		// Populate metadataMap and specMap from 'cred'
		metadataMap["name"] = cred.Metadata.Name
		metadataMap["project"] = cred.Metadata.Project
		// ... populate other metadata fields

		specMap["type"] = cred.Spec.Type
		specMap["provider"] = cred.Spec.Provider
		// ... populate other spec fields

		credentialMap["metadata"] = metadataMap
		credentialMap["spec"] = specMap

		switch v := cred.Spec.Credentials.(type) {
		case *infrapb.CredentialsSpec_Azure:
			credentialsMap["tenantId"] = v.Azure.TenantId
			credentialsMap["subscriptionId"] = v.Azure.SubscriptionId
			credentialsMap["clientId"] = v.Azure.ClientId
			credentialsMap["clientSecret"] = v.Azure.ClientSecret

		case *infrapb.CredentialsSpec_Vsphere:
			credentialsMap["gatewayId"] = v.Vsphere.GatewayId
			credentialsMap["vsphereServer"] = v.Vsphere.VsphereServer
			credentialsMap["username"] = v.Vsphere.Username
			credentialsMap["password"] = v.Vsphere.Password

		case *infrapb.CredentialsSpec_AwsAccessKey:
			credentialsMap["type"] = "AccessKey"
			credentialsMap["accessId"] = v.AwsAccessKey.AccessId
			credentialsMap["secretKey"] = v.AwsAccessKey.SecretKey

		case *infrapb.CredentialsSpec_AwsRole:
			credentialsMap["type"] = v.AwsRole.Type //Setting type here
			credentialsMap["arn"] = v.AwsRole.Arn
			credentialsMap["externalId"] = v.AwsRole.ExternalId
			credentialsMap["accountId"] = v.AwsRole.AccountId

		case *infrapb.CredentialsSpec_Gcp:
			credentialsMap["file"] = v.Gcp.File

		case *infrapb.CredentialsSpec_SshRemote:
			// Handle SSH Remote credentials.
			var agentNames []string
			for _, agent := range v.SshRemote.Agents {
				agentNames = append(agentNames, agent.Name)
			}
			credentialsMap["agents"] = strings.Join(agentNames, ",")
			credentialsMap["privateKey"] = v.SshRemote.PrivateKey
			credentialsMap["passphrase"] = v.SshRemote.Passphrase
			credentialsMap["username"] = v.SshRemote.Username
			credentialsMap["port"] = v.SshRemote.Port
		}

		credentialMap["credentials"] = credentialsMap

		credentialsList[i] = credentialMap
	}

	if err := d.Set("credentials_list", credentialsList); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(project)

	return diags
}
