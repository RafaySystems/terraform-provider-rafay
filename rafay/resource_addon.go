package rafay

import (
	"context"
	"fmt"
	"log"
	"time"
	"path/filepath"

	"github.com/RafaySystems/rctl/pkg/addon"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonCreate,
		ReadContext:   resourceAddonRead,
		UpdateContext: resourceAddonUpdate,
		DeleteContext: resourceAddonDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"yamlfilepath": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:	schema.TypeString,
				Required: true,
			},
			"addontype": {
				Type:	schema.TypeString,
				Required: true,
			},
			"versionname": {
				Type:	schema.TypeString,
				Required: true,
			},
			"chartfile": {
				Type: schema.TypeString,
				Required: true,
			},
                        "valuesfile": {
                                Type: schema.TypeString,
                                Required: true,
                        },
		},
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	chartfile := d.Get("chartfile").(string)
	valuesfile := d.Get("valuesfile").(string)


        // get project details
        resp, err := project.GetProjectByName(d.Get("projectname").(string))
        if err != nil {
                fmt.Print("project does not exist")
                return diag.FromErr(err)
        }
        project, err := project.NewProjectFromResponse([]byte(resp))
        if err != nil {
                fmt.Printf("project does not exist")
                return diag.FromErr(err)
        }
	if "NativeYaml" == d.Get("addontype").(string) {
		_, errCreate := addon.CreateAddon(d.Get("namespace").(string), d.Get("name").(string), project.ID, "NativeYaml" )
		if errCreate != nil {
			log.Printf("Error while CreateAddon %s",err.Error())
			return diag.FromErr(errCreate)
		}
	} else if "Helm"  == d.Get("addontype").(string) {
                _, errCreate := addon.CreateAddon(d.Get("namespace").(string), d.Get("name").(string), project.ID, "Helm" )
                if errCreate != nil {
                        log.Printf("Error while CreateAddon %s",err.Error())
                        return diag.FromErr(errCreate)
                }

	} else {
		log.Printf("addontype must be NativaYaml/Helm" )
		return diags
	}
	addonVersionName := d.Get("versionname").(string)

	s, err := addon.GetAddon(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error while GetAddon %s", err.Error())
		return diag.FromErr(err)
	}
	if "NativeYaml" == d.Get("addontype").(string) {
		if !utils.FileExists(YamlConfigFilePath) {
			log.Printf("file %s does not exist", YamlConfigFilePath)
			return diags
		}
		if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
			log.Printf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath))
			return diags
		}
		errversion := addon.CreateAddonVersion(s.ID, addonVersionName, project.ID, YamlConfigFilePath , "", "", models.RepoArtifactMeta{} )
		if errversion != nil {
			log.Printf("error while createAddonVersion() %s", errversion.Error() )
			return diag.FromErr(errversion)
		}
	} else if "Helm" == d.Get("addontype").(string) {
		if !utils.FileExists(chartfile) {
			log.Printf("file %s does not exist", YamlConfigFilePath)
			return diags
		}
		if !utils.FileExists(valuesfile) {
			log.Printf("file %s does not exist", YamlConfigFilePath)
			return diags
		}

               errversion := addon.CreateAddonVersion(s.ID, addonVersionName, project.ID, chartfile ,valuesfile, "", models.RepoArtifactMeta{} )
                if errversion != nil {
                        log.Printf("error while createAddonVersion() %s", errversion.Error() )
                        return diag.FromErr(errversion)
                }

	}
	log.Printf("resource eks cluster created %s", s.ID)
	d.SetId(s.ID)

	return diags
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	resp, err := project.GetProjectByName(d.Get("projectname").(string))
	if err != nil {
		fmt.Print("project name missing in the resource")
		return diags
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diags
	}
	c, err := addon.GetAddon(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in getAddon %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("name", c.Name); err != nil {
                log.Printf("set name error %s", err.Error())
                return diag.FromErr(err)
        }

	return diags
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update EKS cluster resource")
	return diags
}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("resource cluster delete id %s", d.Id())

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

	errDel := addon.DeleteAddon(d.Get("name").(string), project.ID)
	if errDel != nil {
		log.Printf("delete addon error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}
