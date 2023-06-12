package rafay

import (
	"context"
	"fmt"
	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/config"
	"log"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterVaultDetails() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterVaultDetailsCreate,
		ReadContext:   resourceClusterVaultDetailsRead,
		UpdateContext: resourceClusterVaultDetailsUpdate,
		DeleteContext: resourceClusterVaultDetailsReadDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"cluster_name": {
				Description: "Cluster name for which the details are required to set up vault authentication.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"project_name": {
				Description: "Project of which cluster is part of",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"kubernetes_host": {
				Description: "Kubernetes Host of the Cluster needed to set vault authentication",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"kubernetes_ca_cert": {
				Description: "Ca Cert of the Cluster needed to set vault authentication",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceClusterVaultDetailsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func resourceClusterVaultDetailsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	return diags
}

func resourceClusterVaultDetailsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	clusterName := d.Get("cluster_name").(string)
	projectName := d.Get("project_name").(string)

	projectId, err := config.GetProjectIdByName(projectName)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Println("resourceClusterVaultDetails ", clusterName, " projectName ", projectId)

	clusterVaultDetails, err := cluster.GetClusterVaultDetails(clusterName, projectId)
	if err != nil {
		log.Printf("get Cluster Vault Details, error %s", err.Error())
		return diag.FromErr(err)
	}

	err = flattenClusterVaultDetails(d, clusterVaultDetails)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(clusterName)
	return diags
}

func flattenClusterVaultDetails(d *schema.ResourceData, in *cluster.ResponseGetVaultDetails) error {
	if in == nil {
		return fmt.Errorf("%s", "failed to get cluster vault details(empty)")
	}

	d.Set("kubernetes_host", in.KubernetesHost)
	d.Set("kubernetes_ca_cert", in.KubernetesCACert)

	return nil
}

func resourceClusterVaultDetailsReadDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	os.Remove(d.Id())
	return diags
}
