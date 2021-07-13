package rafay

import (
	"context"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
//	"github.com/RafaySystems/rctl/pkg/models"
//	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/cloudprovider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/* type cloudProviderType int

const (
        CredTypeClusterProvisioning cloudProviderType = 1
        CredTypeDataBackup          cloudProviderType = 2
}*/

func resourceCloudCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudCredentialCreate,
		ReadContext:   resourceCloudCredentialRead,
//		UpdateContext: resourceCloudCredentialUpdate,
		DeleteContext: resourceCloudCredentialDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"roleARN": {
				Type: schema.TypeString,
				Required: true,
			},
			"credType": {
				Type: schema.TypeInt,
				Required: true,
			},
			"externalId": {
				Type: schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCloudCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	p := config.GetConfig().ProjectID

	log.Printf("create cloud cred with name %s", d.Get("name").(string))
//	_, err := cloudprovider.CreateAWSCloudRoleCredentials(d.Get("name").(string), p, d.Get("roleARN").(string), d.Get("externalId").(string), d.Get("credType").(cloudProviderType))
	_, err := cloudprovider.CreateAWSCloudRoleCredentials(d.Get("name").(string), p, d.Get("roleARN").(string), d.Get("externalId").(string), 1 )
	if err != nil {
		log.Printf("create cloud credential error %s", err.Error())
		return diag.FromErr(err)
	}

	c, err := cloudprovider.GetCloudProvider(d.Get("name").(string), p )
	if err != nil {
		log.Printf("get cloudprovider after creation failed, error %s", err.Error())
		return diag.FromErr(err)
	}

		log.Printf("get cloudprovider after creation %s",c.ID )

	return diags
}


func resourceCloudCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("resource project read id %s", d.Id())
	//resp, err := project.GetProjectByName(d.Get("name").(string))
	resp, err := getProjectById(d.Id())
	if err != nil {
		log.Printf("get project by id, error %s", err.Error())
		return diag.FromErr(err)
	}

	p, err := getProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("get project response error %s", err.Error())
		return diag.FromErr(err)
	} else if p == nil {
		d.SetId("")
		return diags
	}

	if err := d.Set("name", p.Name); err != nil {
		log.Printf("read project set name error %s", err.Error())
		return diag.FromErr(err)
	}

	if len(p.Description) > 0 {
		if err := d.Set("description", p.Description); err != nil {
			log.Printf("read project set description error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceCloudCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource project delete id %s", d.Id())

	err := cloudprovider.DeleteCloudProvider(d.Get("name").(string),d.Id())
	if err != nil {
		log.Printf("delete project error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}
