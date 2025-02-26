package rafay

import (
	"context"
	"encoding/json"
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

func dataCloudCredential() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataCloudCredentialRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        cloudCredentialSchema(),
	}
}

func cloudCredentialSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Name of the cloud credential",
		},
		"project": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Project of the cloud credential",
		},
		"metadata": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Standard object metadata",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"spec": {
			Type:        schema.TypeMap,
			Computed:    true,
			Description: "Specification of the cloud credential",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func dataCloudCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataCloudCredentialRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	name := d.Get("name").(string)
	project := d.Get("project").(string)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Credentials().Get(ctx, options.GetOptions{
		Name:    name,
		Project: project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	metadata := map[string]interface{}{
		"name":       ag.Metadata.Name,
		"project":    ag.Metadata.Project,
		"modifiedAt": ag.Metadata.ModifiedAt.String(),
	}

	// Initialize the sharing map if it is enabled
	var sharing map[string]interface{}
	if ag.Spec.Sharing != nil {

		if ag.Spec.Sharing.Enabled {
			sharing = map[string]interface{}{
				"enabled":  ag.Spec.Sharing.Enabled,
				"projects": []interface{}{},
			}

			// Populate sharing projects if present
			if ag.Spec.Sharing.Projects != nil {
				for _, proj := range ag.Spec.Sharing.Projects {
					sharing["projects"] = append(sharing["projects"].([]interface{}), map[string]interface{}{
						"name": proj.Name,
					})
				}
			}
		}
	}

	// Prepare the spec map, conditionally including the sharing field
	spec := map[string]interface{}{
		"type":     ag.Spec.Type,
		"provider": ag.Spec.Provider,
	}

	// Only include the sharing field if it's enabled and not nil
	if sharing != nil {
		// Convert the sharing map to JSON string
		sharingJSON, err := json.Marshal(sharing)
		if err != nil {
			return diag.FromErr(err)
		}
		spec["sharing"] = string(sharingJSON)
	}

	credentials := map[string]interface{}{}
	switch v := ag.Spec.Credentials.(type) {
	case *infrapb.CredentialsSpec_AwsRole:
		// Access AWS Role credentials
		awsRole := v.AwsRole
		// Populate credentials map with relevant fields
		credentials["arn"] = awsRole.Arn
		credentials["externalId"] = awsRole.ExternalId
		credentials["accountId"] = awsRole.AccountId
	case *infrapb.CredentialsSpec_AwsAccessKey:
		awsAccessKey := v.AwsAccessKey
		credentials["accessId"] = awsAccessKey.AccessId
		credentials["secretKey"] = awsAccessKey.SecretKey
		credentials["sessionToken"] = awsAccessKey.SessionToken
	case *infrapb.CredentialsSpec_Gcp:
		// Access GCP credentials
		gcp := v.Gcp
		credentials["file"] = gcp.File
	case *infrapb.CredentialsSpec_Azure:
		// Access Azure credentials
		azure := v.Azure
		if azure != nil {
			credentials["tenantId"] = azure.TenantId
			credentials["subscriptionId"] = azure.SubscriptionId
			credentials["clientId"] = azure.ClientId
		}
	case *infrapb.CredentialsSpec_SshRemote:
		sshRemote := v.SshRemote
		credentials["username"] = sshRemote.Username
		credentials["port"] = sshRemote.Port

		var agentNames []string
		for _, agent := range v.SshRemote.Agents {
			agentNames = append(agentNames, agent.Name)
		}

		credentials["agents"] = strings.Join(agentNames, ",")
	}

	credentialsJSON, err := json.Marshal(credentials)
	if err != nil {
		return diag.FromErr(err)
	}
	spec["credentials"] = string(credentialsJSON)

	if err := d.Set("metadata", metadata); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("spec", spec); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return diags
}
