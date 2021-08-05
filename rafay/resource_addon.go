package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"github.com/RafaySystems/rctl/pkg/addon"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"addontype": {
				Type:     schema.TypeString,
				Required: true,
			},
			"versionname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"yamlfilepath": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"chartfile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valuesfile": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"configmap": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"secret": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"statefulset": {
				Type:     schema.TypeString,
				Optional: true,
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
	if d.Get("addontype").(string) == "yaml" {
		if d.Get("namespace").(string) == "" {
			return diag.FromErr(fmt.Errorf("namespace cannot be empty for yaml "))
		}
		_, errCreate := addon.CreateAddon(d.Get("namespace").(string), d.Get("name").(string), project.ID, "NativeYaml")
		if errCreate != nil {
			log.Printf("Error while CreateAddon %s", errCreate.Error())
			return diag.FromErr(errCreate)
		}
	} else if d.Get("addontype").(string) == "helm" {
		if d.Get("namespace").(string) == "" {
			return diag.FromErr(fmt.Errorf("namespace cannot be empty for helm"))
		}
		_, errCreate := addon.CreateAddon(d.Get("namespace").(string), d.Get("name").(string), project.ID, "Helm")
		if errCreate != nil {
			log.Printf("Error while CreateAddon %s", errCreate.Error())
			return diag.FromErr(errCreate)
		}
	} else if d.Get("addontype").(string) == "helm3" {
		if d.Get("namespace").(string) == "" {
			return diag.FromErr(fmt.Errorf("namespace cannot be empty for helm3"))
		}
		_, errCreate := addon.CreateAddon(d.Get("namespace").(string), d.Get("name").(string), project.ID, "NativeHelm")
		if errCreate != nil {
			log.Printf("Error while CreateAddon %s", errCreate.Error())
			return diag.FromErr(errCreate)
		}
	} else if d.Get("addontype").(string) == "alertmanager" {
		_, errCreate := addon.CreateManagedAddon(d.Get("name").(string), project.ID)
		if errCreate != nil {
			log.Printf("Error while CreateManageaddon %s", errCreate.Error())
			return diag.FromErr(errCreate)
		}
	} else {
		log.Printf("addontype must be nativaYaml/helm/helm3/alertmanager")
		return diags
	}
	addonVersionName := d.Get("versionname").(string)
	var addonID string
	if d.Get("addontype").(string) == "alertmanager" {
		_, err := addon.GetManagedAddon(d.Get("name").(string), project.ID)
		if err != nil {
			log.Printf("error while GetAddon %s", err.Error())
			return diag.FromErr(err)
		}
		d.SetId(d.Get("name").(string))

	} else {
		s, err := addon.GetAddon(d.Get("name").(string), project.ID)
		if err != nil {
			log.Printf("error while GetAddon %s", err.Error())
			return diag.FromErr(err)
		}
		addonID = s.ID
		d.SetId(s.ID)
	}
	if d.Get("addontype").(string) == "nativeYaml" {
		if !utils.FileExists(YamlConfigFilePath) {
			log.Printf("file %s does not exist", YamlConfigFilePath)
			return diags
		}
		if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
			log.Printf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath))
			return diags
		}
		errversion := addon.CreateAddonVersion(addonID, addonVersionName, project.ID, YamlConfigFilePath, "", "", models.RepoArtifactMeta{})
		if errversion != nil {
			log.Printf("error while createAddonVersion() %s", errversion.Error())
			return diag.FromErr(errversion)
		}
	} else if d.Get("addontype").(string) == "helm" {
		if !utils.FileExists(chartfile) {
			log.Printf("file %s does not exist", chartfile)
			return diags
		}
		if valuesfile == "" {
			log.Printf("Valuesfile cannot be empty")
			return diags
		}

		errversion := addon.CreateAddonVersion(addonID, addonVersionName, project.ID, chartfile, valuesfile, "", models.RepoArtifactMeta{})
		if errversion != nil {
			log.Printf("error while createAddonVersion() %s", errversion.Error())
			return diag.FromErr(errversion)
		}
	} else if d.Get("addontype").(string) == "helm3" {
		if !utils.FileExists(chartfile) {
			log.Printf("file %s does not exist", chartfile)
			return diags
		}
		if valuesfile == "" {
			log.Printf("Valuesfile cannot be empty")
			return diags
		}
		errversion := addon.CreateAddonVersion(addonID, addonVersionName, project.ID, chartfile, valuesfile, "", models.RepoArtifactMeta{})
		if errversion != nil {
			log.Printf("error while createAddonVersion() %s", errversion.Error())
			return diag.FromErr(errversion)
		}
	} else if d.Get("addontype").(string) == "alertmanager" {
		configmap := d.Get("configmap").(string)
		configuration := d.Get("configuration").(string)
		secret := d.Get("secret").(string)
		statefulset := d.Get("statefulset").(string)

		if configmap == "" && configuration == "" && secret == "" && statefulset == "" {
			log.Printf("for alertmanager addons, you must provide one or more of the fields")
			return diags
		}
		var cm, s, st, c []byte
		var err error

		if configmap != "" {
			cm, err = ioutil.ReadFile(configmap)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if configuration != "" {
			c, err = ioutil.ReadFile(configuration)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if secret != "" {
			s, err = ioutil.ReadFile(secret)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if statefulset != "" {
			st, err = ioutil.ReadFile(statefulset)
			if err != nil {
				return diag.FromErr(err)
			}
		}
		_, errversion := addon.CreateManagedAddonVersion(d.Get("name").(string), d.Get("versionname").(string), string(cm), string(c), string(s), string(st), project.ID)
		if err != nil {
			log.Printf("error While createManageaddonversion %s", errversion.Error())
			return diag.FromErr(errversion)
		}
	}
	log.Printf("resource addon created ")

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
	if d.Get("addontype").(string) == "alertmanager" {
		_, err := addon.GetManagedAddon(d.Get("name").(string), project.ID)
		if err != nil {
			log.Printf("error in getAddon %s", err.Error())
			return diag.FromErr(err)
		}
		if err := d.Set("name", d.Get("name").(string)); err != nil {
			log.Printf("set name error %s", err.Error())
			return diag.FromErr(err)
		}
	} else {
		c, err := addon.GetAddon(d.Get("name").(string), project.ID)
		if err != nil {
			log.Printf("error in getAddon %s", err.Error())
			return diag.FromErr(err)
		}
		if err := d.Set("name", c.Name); err != nil {
			log.Printf("set name error %s", err.Error())
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update addon resource")
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
	if d.Get("addontype").(string) == "alertmanager" {
		errDel := addon.DeleteManagedAddon(d.Get("name").(string), project.ID)
		if errDel != nil {
			log.Printf("delete addon error %s", errDel.Error())
			return diag.FromErr(errDel)
		}
	} else {
		errDel := addon.DeleteAddon(d.Get("name").(string), project.ID)
		if errDel != nil {
			log.Printf("delete addon error %s", errDel.Error())
			return diag.FromErr(errDel)
		}
	}
	return diags
}
