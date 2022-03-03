package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/RafaySystems/rctl/pkg/cloudprovider"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"

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
				Type:        schema.TypeString,
				Description: "Name of cloud crredential",
				Required:    true,
				ForceNew:    true,
			},
			"projectname": {
				Type:        schema.TypeString,
				Description: "Name of the project",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the project",
				Optional:    true,
			},
			"providertype": {
				Type:        schema.TypeString,
				Description: "Type of the cloud provider. Accepted values are 'AWS', 'AZURE', 'GCP', 'MINIO'",
				Required:    true,
			},
			"awscredtype": {
				Type:        schema.TypeString,
				Description: "AWS cloud provider credential type. Accepted values are 'rolearn', 'accesskey'",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "Cloud credential type. Accepted values are 'cluster-provisioning', 'data-backup' ",
				Required:    true,
			},
			"rolearn": {
				Type:        schema.TypeString,
				Description: "AWS role ARN.",
				Optional:    true,
			},
			"externalid": {
				Type:        schema.TypeString,
				Description: "External ID.",
				Optional:    true,
			},
			"accesskey": {
				Type:        schema.TypeString,
				Description: "AWS accesskey.",
				Optional:    true,
			},
			"secretkey": {
				Type:        schema.TypeString,
				Description: "AWS secret key.",
				Optional:    true,
			},
			"credfile": {
				Type:        schema.TypeString,
				Description: "Credential file.",
				Optional:    true,
			},
			"clientid": {
				Type:        schema.TypeString,
				Description: "Azure client ID.",
				Optional:    true,
			},
			"clientsecret": {
				Type:        schema.TypeString,
				Description: "Azure client secret.",
				Optional:    true,
			},
			"tenantid": {
				Type:        schema.TypeString,
				Description: "Azure tenant ID.",
				Optional:    true,
			},
			"subscriptionid": {
				Type:        schema.TypeString,
				Description: "Azure subscription ID.",
				Optional:    true,
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
			if d.Get("awscredtype").(string) == "rolearn" {
				if d.Get("rolearn").(string) == "" {
					return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
				}
				log.Printf("create cloud credential with name %s, %s", d.Get("name").(string), project.ID)
				s, err := cloudprovider.CreateAWSCloudRoleCredentials(
					d.Get("name").(string),
					project.ID, d.Get("rolearn").(string),
					d.Get("externalid").(string),
					1)
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			} else {
				if d.Get("accesskey").(string) == "" {
					return diag.FromErr(fmt.Errorf("accesskey cannot be empty"))
				}
				if d.Get("secretkey").(string) == "" {
					return diag.FromErr(fmt.Errorf("secretkey cannot be empty"))
				}
				s, err := cloudprovider.CreateAWSCloudAccessKeyCredentials(
					d.Get("name").(string),
					project.ID,
					d.Get("accesskey").(string),
					d.Get("secretkey").(string), "",
					1)
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			}
		} else if d.Get("providertype").(string) == "GCP" {
			credFile := d.Get("credfile").(string)
			if !utils.FileExists(credFile) {
				log.Printf("file %s not exist", credFile)
				return diags
			}
			byteContents, err := ioutil.ReadFile(credFile)
			if err != nil {
				log.Printf("Error while reading GCP jsonfile  %s", err.Error())
				return diag.FromErr(err)
			}
			s, err := cloudprovider.CreateGCPCloudRoleCredentials(d.Get("name").(string), project.ID, string(byteContents))
			if err != nil {
				log.Printf("Error while creatGCPRole()  %s", err.Error())
				return diag.FromErr(err)
			}
			d.SetId(s.ID)
		} else if d.Get("providertype").(string) == "AZURE" {
			s, err := cloudprovider.CreateAzureCloudCredentials(d.Get("name").(string),
				project.ID,
				d.Get("clientid").(string),
				d.Get("clientsecret").(string),
				d.Get("subscriptionid").(string),
				d.Get("tenantid").(string),
				3)
			if err != nil {
				log.Printf("Error while create azure credentials  %s", err.Error())
				return diag.FromErr(err)
			}
			d.SetId(s.ID)
		} else {
			return diag.FromErr(fmt.Errorf("providertype is not correct for cluster-provisioning,( use AWS or GCP )"))
		}
	} else if d.Get("type").(string) == "data-backup" {
		if d.Get("providertype").(string) == "MINIO" {
			if d.Get("awscredtype").(string) == "rolearn" {
				if d.Get("rolearn").(string) == "" {
					return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
				}
				s, err := cloudprovider.CreateMinioCloudRoleCredentials(d.Get("name").(string), project.ID, d.Get("rolearn").(string), d.Get("externalid").(string))
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			} else {
				if d.Get("accesskey").(string) == "" {
					return diag.FromErr(fmt.Errorf("accesskey cannot be empty"))
				}
				if d.Get("secretkey").(string) == "" {
					return diag.FromErr(fmt.Errorf("secretkey cannot be empty"))
				}
				s, err := cloudprovider.CreateMinioCloudAccessKeyCredentials(d.Get("name").(string), project.ID, d.Get("accesskey").(string), d.Get("secretkey").(string), "")
				if err != nil {
					log.Printf("create cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			}
		}
	} else {
		return diag.FromErr(fmt.Errorf("type is not correct ( cluster-provisioning or data-backup )"))
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
		//return diag.FromErr(err)
		if err := d.Set("name", c.Name); err != nil {
			d.Set("name", "")
		}
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
	// delete and recreate
	cloudprovider.DeleteCloudProvider(d.Get("name").(string), project.ID)
	return resourceCloudCredentialCreate(ctx, d, m)

	// if d.Get("type").(string) == "cluster-provisioning" {
	// 	if d.Get("providertype").(string) == "AWS" {
	// 		if d.Get("awscredtype").(string) == "rolearn" {
	// 			if d.Get("rolearn").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
	// 			}
	// 			log.Printf("create cloud credential with name %s, %s", d.Get("name").(string), project.ID)
	// 			s, err := cloudprovider.UpdateAWSCloudRoleCredentials(d.Get("name").(string),
	// 				project.ID,
	// 				d.Get("rolearn").(string),
	// 				d.Get("externalid").(string),
	// 				1)
	// 			if err != nil {
	// 				log.Printf("create cloud credential error %s", err.Error())
	// 				return diag.FromErr(err)
	// 			}
	// 			d.SetId(s.ID)
	// 		} else {
	// 			if d.Get("accesskey").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("accesskey cannot be empty"))
	// 			}
	// 			if d.Get("secretkey").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("secretkey cannot be empty"))
	// 			}
	// 			s, err := cloudprovider.UpdateAWSCloudAccessKeyCredentials(
	// 				d.Get("name").(string),
	// 				project.ID, d.Get("accesskey").(string),
	// 				d.Get("secretkey").(string),
	// 				"",
	// 				1)
	// 			if err != nil {
	// 				log.Printf("create cloud credential error %s", err.Error())
	// 				return diag.FromErr(err)
	// 			}
	// 			d.SetId(s.ID)
	// 		}
	// 	} else if d.Get("providertype").(string) == "GCP" {
	// 		credFile := d.Get("credfile").(string)
	// 		if !utils.FileExists(credFile) {
	// 			log.Printf("file %s not exist", credFile)
	// 			return diags
	// 		}
	// 		byteContents, err := ioutil.ReadFile(credFile)
	// 		if err != nil {
	// 			log.Printf("Error while reading GCP jsonfile  %s", err.Error())
	// 			return diag.FromErr(err)
	// 		}
	// 		s, err := cloudprovider.CreateGCPCloudRoleCredentials(
	// 			d.Get("name").(string),
	// 			project.ID,
	// 			string(byteContents))
	// 		if err != nil {
	// 			log.Printf("Error while creatGCPRole()  %s", err.Error())
	// 			return diag.FromErr(err)
	// 		}
	// 		d.SetId(s.ID)
	// 	} else if d.Get("providertype").(string) == "AZURE" {
	// 		s, err := cloudprovider.UpdateAzureCloudCredentials(
	// 			d.Get("name").(string),
	// 			project.ID,
	// 			d.Get("clientid").(string),
	// 			d.Get("clientsecret").(string),
	// 			d.Get("subscriptionid").(string),
	// 			d.Get("tenantid").(string))
	// 		if err != nil {
	// 			log.Printf("Error while create azure credentials  %s", err.Error())
	// 			return diag.FromErr(err)
	// 		}
	// 		d.SetId(s.ID)
	// 	} else {
	// 		return diag.FromErr(fmt.Errorf("providertype is not correct for cluster-provisioning,( use AWS or GCP )"))
	// 	}
	// } else if d.Get("type").(string) == "data-backup" {
	// 	if d.Get("providertype").(string) == "MINIO" {
	// 		if d.Get("awscredtype").(string) == "rolearn" {
	// 			if d.Get("rolearn").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
	// 			}
	// 			s, err := cloudprovider.CreateMinioCloudRoleCredentials(d.Get("name").(string), project.ID, d.Get("rolearn").(string), d.Get("externalid").(string))
	// 			if err != nil {
	// 				log.Printf("create cloud credential error %s", err.Error())
	// 				return diag.FromErr(err)
	// 			}
	// 			d.SetId(s.ID)
	// 		} else {
	// 			if d.Get("accesskey").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("accesskey cannot be empty"))
	// 			}
	// 			if d.Get("secretkey").(string) == "" {
	// 				return diag.FromErr(fmt.Errorf("secretkey cannot be empty"))
	// 			}
	// 			s, err := cloudprovider.CreateMinioCloudAccessKeyCredentials(d.Get("name").(string), project.ID, d.Get("accesskey").(string), d.Get("secretkey").(string), "")
	// 			if err != nil {
	// 				log.Printf("create cloud credential error %s", err.Error())
	// 				return diag.FromErr(err)
	// 			}
	// 			d.SetId(s.ID)
	// 		}
	// 	}
	// } else {
	// 	return diag.FromErr(fmt.Errorf("type is not correct ( cluster-provisioning or data-backup )"))
	// }
	// log.Printf("resource cloud credential created ")

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
		//return diag.FromErr(errDel)
	}

	return diags
}
