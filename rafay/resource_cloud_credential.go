package rafay

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/pkg/cloudprovider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceCloudCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudCredentialCreate,
		ReadContext:   resourceCloudCredentialRead,
		UpdateContext: resourceCloudCredentialUpdate,
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
			"projectname": {
				Type:	schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"rolearn": {
				Type: schema.TypeString,
				Required: true,
			},
			"credtype": {
				Type: schema.TypeInt,
				Required: true,
			},
			"externalid": {
				Type: schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceCloudCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resp, err := project.GetProjectByName(d.Get("projectname").(string))
        if err != nil {
                fmt.Print("project  does not exist")
		return diags
	}

        project, err := project.NewProjectFromResponse([]byte(resp))
        if err != nil {
                fmt.Printf("project  does not exist")
		return diags
        }

	log.Printf("create cloud cred with name %s, %s", d.Get("name").(string), project.ID )
	s, err := cloudprovider.CreateAWSCloudRoleCredentials(d.Get("name").(string), project.ID, d.Get("rolearn").(string), d.Get("externalid").(string), 1 )
	if err != nil {
		log.Printf("create cloud credential error %s %s", err.Error(),s)
		return diag.FromErr(err)
	}
        log.Printf("resource cloud credential created %s", s.ID)
        d.SetId(s.ID)

	return diags
}

func resourceCloudCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
        if err != nil {
                fmt.Print("project does not exist")
                return diags
        }

        project, err := project.NewProjectFromResponse([]byte(resp))
        if err != nil {
                fmt.Printf("project does not exist")
                return diags
        }
	c, err := cloudprovider.GetCloudProvider(d.Get("name").(string), project.ID )
	if err != nil {
		log.Printf("error while getCloudprovider %s", err.Error())
		return diag.FromErr(err)
	}
        if err := d.Set("name", c.Name); err != nil {
                log.Printf("get group set name error %s", err.Error())
                return diag.FromErr(err)
        }

	return diags
}

func resourceCloudCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	log.Printf("get cloudprovider " )
	return diags
}

func resourceCloudCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource project delete id %s", d.Id())

        resp, err := project.GetProjectByName(d.Get("projectname").(string))
        if err != nil {
                fmt.Print("project  does not exist")
                return diags
        }

        project, err := project.NewProjectFromResponse([]byte(resp))
        if err != nil {
                fmt.Printf("project  does not exist")
                return diags
        }

	errDel := cloudprovider.DeleteCloudProvider(d.Get("name").(string),project.ID)
	if errDel != nil {
		log.Printf("delete project error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}
