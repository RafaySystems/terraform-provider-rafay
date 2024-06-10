package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/cloudprovider"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type clusterType int

const (
	awsProviderInt clusterType = iota + 1
	gcpProviderInt
	azureProviderInt
	minioProviderInt
	vsphereProviderInt
)

func resourceCloudCredential() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCloudCredentialCreate,
		ReadContext:   resourceCloudCredentialRead,
		UpdateContext: resourceCloudCredentialUpdate,
		DeleteContext: resourceCloudCredentialDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCloudCredentialImport,
		},

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
			"project": {
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
				Sensitive:   true,
			},
			"accesskey": {
				Type:        schema.TypeString,
				Description: "AWS accesskey.",
				Optional:    true,
				Sensitive:   true,
			},
			"secretkey": {
				Type:        schema.TypeString,
				Description: "AWS secret key.",
				Optional:    true,
				Sensitive:   true,
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
				Sensitive:   true,
			},
			"clientsecret": {
				Type:        schema.TypeString,
				Description: "Azure client secret.",
				Optional:    true,
				Sensitive:   true,
			},
			"tenantid": {
				Type:        schema.TypeString,
				Description: "Azure tenant ID.",
				Optional:    true,
				Sensitive:   true,
			},
			"subscriptionid": {
				Type:        schema.TypeString,
				Description: "Azure subscription ID.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func resourceCloudCredentialCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceCredentialsCreate create starts")
	diags := resourceCredentialsUpsert(ctx, d, m)
	return diags
}

func resourceCloudCredentialRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := project.GetProjectByName(d.Get("project").(string))
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

	var typeStr string
	if c.Type == int(cloudprovider.CredTypeClusterProvisioning) {
		typeStr = "cluster-provisioning"
	} else if c.Type == int(cloudprovider.CredTypeDataBackup) {
		typeStr = "data-backup"
	}
	if typeStr != "" {
		if err := d.Set("type", typeStr); err != nil {
			log.Printf("get cloud credential set type error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	var providerString, credType string
	if c.Provider == int(awsProviderInt) {
		providerString = "AWS"
		if c.CredentialType == int(cloudprovider.CredTypeAccessKey) {
			credType = "accesskey"
		} else if c.CredentialType == int(cloudprovider.CredTypeRole) {
			credType = "rolearn"
		}

	} else if c.Provider == int(azureProviderInt) {
		providerString = "AZURE"
	} else if c.Provider == int(gcpProviderInt) {
		providerString = "GCP"
	} else if c.Provider == int(minioProviderInt) {
		providerString = "MINIO"
	}
	if providerString != "" {
		if err := d.Set("providertype", providerString); err != nil {
			log.Printf("get cloud credential set providertype error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	if credType != "" {
		if err := d.Set("awscredtype", credType); err != nil {
			log.Printf("get cloud credential set awscredtype error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	if typeStr == "cluster-provisioning" || typeStr == "data-backup" {
		switch providerString {
		case "AWS":
			if credType == "accesskey" {
				if err := setCredentialAttrState(d, "accesskey", c.AccessKey); err != nil {
					return diag.FromErr(err)
				}
			} else if credType == "rolearn" {
				if err := setCredentialAttrState(d, "rolearn", c.RoleName); err != nil {
					return diag.FromErr(err)
				}
				if err := setCredentialAttrState(d, "externalid", c.ExternalID); err != nil {
					return diag.FromErr(err)
				}
			}
		case "AZURE":
			if err := setCredentialAttrState(d, "clientid", c.ClientID); err != nil {
				return diag.FromErr(err)
			}
		case "GCP":
			/*TODO: Handle credentials remote read here*/
		case "MINIO":
			/*TODO: Handle credentials remote read here*/
		}
	}
	return diags
}

func resourceCloudCredentialUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Credentials update starts...")
	return resourceCredentialsUpsert(ctx, d, m)
}

func resourceCredentialsUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	resp, err := project.GetProjectByName(d.Get("project").(string))
	if err != nil {
		fmt.Print("project  does not exist")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}

	providerExists := false
	_, err = cloudprovider.GetCloudProvider(d.Get("name").(string), project.ID)
	if err == nil {
		providerExists = true
	}

	if d.Get("type").(string) == "cluster-provisioning" || d.Get("type").(string) == "data-backup" {
		credType := cloudprovider.CredTypeClusterProvisioning
		if d.Get("type").(string) == "data-backup" {
			credType = cloudprovider.CredTypeDataBackup
		}
		if d.Get("providertype").(string) == "AWS" {
			if d.Get("awscredtype").(string) == "rolearn" {
				if d.Get("rolearn").(string) == "" {
					return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
				}
				log.Printf("create/update cloud credential with name %s, %s", d.Get("name").(string), project.ID)
				var s models.CloudProvider
				if !providerExists {
					s, err = cloudprovider.CreateAWSCloudRoleCredentials(
						d.Get("name").(string),
						project.ID, d.Get("rolearn").(string),
						d.Get("externalid").(string),
						credType)
				} else {
					s, err = cloudprovider.UpdateAWSCloudRoleCredentials(
						d.Get("name").(string),
						project.ID, d.Get("rolearn").(string),
						d.Get("externalid").(string),
						credType)
				}
				if err != nil {
					log.Printf("create/update cloud credential error %s", err.Error())
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
				var s models.CloudProvider
				if !providerExists {
					s, err = cloudprovider.CreateAWSCloudAccessKeyCredentials(
						d.Get("name").(string),
						project.ID,
						d.Get("accesskey").(string),
						d.Get("secretkey").(string), "",
						credType)
				} else {
					s, err = cloudprovider.UpdateAWSCloudAccessKeyCredentials(
						d.Get("name").(string),
						project.ID,
						d.Get("accesskey").(string),
						d.Get("secretkey").(string), "",
						credType)
				}
				if err != nil {
					log.Printf("create/update cloud credential error %s", err.Error())
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
			var s models.CloudProvider
			if !providerExists {
				s, err = cloudprovider.CreateGCPCloudRoleCredentials(d.Get("name").(string), project.ID, string(byteContents))
			} else {
				s1, err1 := cloudprovider.UpdateGCPCloudRoleCredentials(d.Get("name").(string), project.ID, string(byteContents))
				if err1 != nil {
					err = err1
				} else if s1 != nil {
					s = *s1
				}
			}
			if err != nil {
				log.Printf("create/update cloud credential error %s", err.Error())
				return diag.FromErr(err)
			}
			d.SetId(s.ID)
		} else if d.Get("providertype").(string) == "AZURE" {
			var s *models.CloudProvider
			if !providerExists {
				s, err = cloudprovider.CreateAzureCloudCredentials(d.Get("name").(string),
					project.ID,
					d.Get("clientid").(string),
					d.Get("clientsecret").(string),
					d.Get("subscriptionid").(string),
					d.Get("tenantid").(string),
					credType)
			} else {
				s, err = cloudprovider.UpdateAzureCloudCredentials(d.Get("name").(string),
					project.ID,
					d.Get("clientid").(string),
					d.Get("clientsecret").(string),
					d.Get("subscriptionid").(string),
					d.Get("tenantid").(string))
			}
			if err != nil {
				log.Printf("create/update cloud credential error  %s", err.Error())
				return diag.FromErr(err)
			}
			d.SetId(s.ID)
		} else if d.Get("providertype").(string) == "MINIO" {
			if d.Get("awscredtype").(string) == "rolearn" {
				return diag.Errorf("Minio + Role ARN is not supported for data-backup.")
				// if d.Get("rolearn").(string) == "" {
				// 	return diag.FromErr(fmt.Errorf("RoleARN cannot be empty"))
				// }
				// s, err := cloudprovider.CreateMinioCloudRoleCredentials(d.Get("name").(string), project.ID, d.Get("rolearn").(string), d.Get("externalid").(string))
				// if err != nil {
				// 	log.Printf("create cloud credential error %s", err.Error())
				// 	return diag.FromErr(err)
				// }
				// d.SetId(s.ID)
			} else {
				if d.Get("accesskey").(string) == "" {
					return diag.FromErr(fmt.Errorf("accesskey cannot be empty"))
				}
				if d.Get("secretkey").(string) == "" {
					return diag.FromErr(fmt.Errorf("secretkey cannot be empty"))
				}
				var s models.CloudProvider
				if !providerExists {
					s, err = cloudprovider.CreateMinioCloudAccessKeyCredentials(d.Get("name").(string), project.ID, d.Get("accesskey").(string), d.Get("secretkey").(string), "")
				} else {
					log.Printf("update cloud credential is not supported for provider type MINIO")
					return diag.FromErr(errors.New("update cloud credential is not supported for provider type MINIO"))
				}
				if err != nil {
					log.Printf("create/update cloud credential error %s", err.Error())
					return diag.FromErr(err)
				}
				d.SetId(s.ID)
			}
		} else {
			return diag.FromErr(fmt.Errorf("providertype is not correct. use AWS or GCP or AZURE"))
		}
	} else {
		return diag.FromErr(fmt.Errorf("type is not correct ( cluster-provisioning or data-backup )"))
	}
	log.Printf("resource cloud credential created/updated.")
	return diags
}

func resourceCloudCredentialDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource project delete id %s", d.Id())

	resp, err := project.GetProjectByName(d.Get("project").(string))
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

func resourceCloudCredentialImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceCloudCredentialImport idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceCloudCredentialImport d.Id:", d.Id())
	log.Println("resourceCloudCredentialImport d_debug", d_debug)

	resp, err := project.GetProjectByName(idParts[1])
	if err != nil {
		log.Println("resourceCloudCredentialImport project does not exist ", err)
		return nil, err
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Println("resourceCloudCredentialImport project resp prase failed ", err)
		return nil, err
	}

	c, err := cloudprovider.GetCloudProvider(idParts[0], project.ID)
	if err != nil {
		log.Printf("failed to get cloud provider %s", err.Error())
		return nil, err
	}

	if err := d.Set("name", c.Name); err != nil {
		return nil, err
	}

	if err := d.Set("project", idParts[1]); err != nil {
		return nil, err
	}

	d.SetId(c.ID)
	return []*schema.ResourceData{d}, nil
}

func setCredentialAttrState(d *schema.ResourceData, key string, value string) error {
	if err := d.Set(key, value); err != nil {
		log.Printf("get cloud credential set %s error. Error: %s", key, err.Error())
		return fmt.Errorf("get cloud credential set %s error. Error: %s", key, err.Error())
	}
	return nil
}
func getCredentialAttrState(d *schema.ResourceData, key string) string {
	value, ok := d.Get(key).(string)
	if !ok {
		value = ""
	}
	return value
}
