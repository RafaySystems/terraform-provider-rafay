package rafay

import (
	"context"
	"fmt"
	"log"
	"time"
	"io/ioutil"

	"github.com/RafaySystems/rctl/pkg/cloudprovider"
	"github.com/RafaySystems/rctl/pkg/project"

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
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"providertype": {
				Type:	 schema.TypeString,
				Required: true,
			},
			"rolearn": {
				Type:     schema.TypeString,
				Required: true,
			},
			"credtype": {
				Type:     schema.TypeString,
				Required: true,
			},
			"externalid": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:	  schema.TypeString,
				Required: true,
			},
			"accesskey": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secretkey": {
				Type:	  schema.TypeString,
				Required: true,
			},
			"credfile": {
				Type:	schema.TypeString,
				Required: true,
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
		fmt.Printf("project does not exist")
		return diags
	}
	if d.Get("type").(string) == "cluster-provisioning" {
		if d.Get("providertype").(string) == "AWS" {
			if d.Get("credtype").(string)  == "rolearn" {
				log.Printf("create cloud credential with name %s, %s", d.Get("name").(string), project.ID)
				s, err := cloudprovider.CreateAWSCloudRoleCredentials(d.Get("name").(string), project.ID, d.Get("rolearn").(string), d.Get("externalid").(string), 1 )
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			} else {
				s, err := cloudprovider.CreateAWSCloudAccessKeyCredentials(d.Get("name").(string), project.ID, d.Get("accesskey").(string), d.Get("secretkey").(string), "", 0 )
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
                                }
				d.SetId(s.ID)
			}
		} else if d.Get("providertype").(string) == "GCP" {
			credFile := d.Get("credfile").(string)
			byteContents, err := ioutil.ReadFile(credFile)
			if err != nil {
				log.Printf("Error while reading GCP jsonfile  %s", err.Error())
				return diag.FromErr(err)
			}
			s,err := cloudprovider.CreateGCPCloudRoleCredentials(d.Get("name").(string),project.ID, string(byteContents))
			if err != nil {
				log.Printf("Error while creatGCPRole()  %s", err.Error())
				return diag.FromErr(err)
			}
			d.SetId(s.ID)
		} else {
			log.Printf("error provider name is not correct")
			return diags
		}
	} else if d.Get("type").(string) == "data-backup" {
		if d.Get("providertype").(string) == "MINIO" {
			if d.Get("credtype").(string) == "rolearn" {
				s, err := cloudprovider.CreateMinioCloudRoleCredentials(d.Get("name").(string), project.ID , d.Get("rolearn").(string) , d.Get("externalid").(string))
				if err  != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			} else {
				s, err := cloudprovider.CreateMinioCloudAccessKeyCredentials(d.Get("name").(string), project.ID , d.Get("accesskey").(string) , d.Get("secretkey").(string) , "")
				if err  != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			}
		}
	} else {
		log.Printf("type is not correct ( cluster-provisioning or data-backup")
		return diags
	}
	log.Printf("resource cloud credential created ")

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
	c, err := cloudprovider.GetCloudProvider(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("failed to get cloud provider %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("name", c.Name); err != nil {
		log.Printf("get cloud credential set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceCloudCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	log.Printf("update cloudcredentials ")
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
		fmt.Printf("project does not exist")
		return diags
	}

	errDel := cloudprovider.DeleteCloudProvider(d.Get("name").(string), project.ID)
	if errDel != nil {
		log.Printf("delete cloud credential error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}
