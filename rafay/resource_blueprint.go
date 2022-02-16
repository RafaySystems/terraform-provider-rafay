package rafay

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/blueprint"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"gopkg.in/yaml.v3"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	ClusterScoped   = "cluster-scoped"
	NamespaceScoped = "namespace-scoped"
)

type blueprintVersionYamlConfig struct {
	Kind     string `yaml:"kind"`
	Metadata struct {
		Name    string `yaml:"name"`
		Project string `yaml:"project"`
	} `yaml:"metadata"`
	Spec struct {
		Blueprint string `yaml:"blueprint"`
		Addons    []struct {
			Name      string   `yaml:"name"`
			Version   string   `yaml:"version"`
			DependsOn []string `yaml:"dependsOn"`
		} `yaml:"addons"`
		Psps                       []string `yaml:"psps"`
		PspScope                   string   `yaml:"pspScope"`
		RafayIngress               *bool    `yaml:"rafayIngress"`
		RafayMonitoringAndAlerting *bool    `yaml:"rafayMonitoringAndAlerting"`
		Kubevirt                   bool     `yaml:"kubevirt"`
		DriftAction                string   `yaml:"driftAction"`
		PrometheusCustomization    struct {
			NodeExporter struct {
				Disable         bool `yaml:"disable"`
				DiscoveryConfig struct {
					Namespace string            `yaml:"namespace"`
					Resource  string            `yaml:"resource"`
					Labels    map[string]string `yaml:"labels"`
				} `yaml:"discoveryConfig"`
			} `yaml:"nodeExporter"`
			HelmExporter struct {
				Disable         bool `yaml:"disable"`
				DiscoveryConfig struct {
					Namespace string            `yaml:"namespace"`
					Resource  string            `yaml:"resource"`
					Labels    map[string]string `yaml:"labels"`
				} `yaml:"discoveryConfig"`
			} `yaml:"helmExporter"`
			KubeStateMetrics struct {
				Disable         bool `yaml:"disable"`
				DiscoveryConfig struct {
					Namespace string            `yaml:"namespace"`
					Resource  string            `yaml:"resource"`
					Labels    map[string]string `yaml:"labels"`
				} `yaml:"discoveryConfig"`
			} `yaml:"kubeStateMetrics"`
			PrometheusAdapter struct {
				Disable bool `yaml:"disable"`
			} `yaml:"prometheusAdapter"`
			MetricsServer struct {
				Disable bool `yaml:"disable"`
			} `yaml:"metricsServer"`
			Resources struct {
				Limits struct {
					Memory string `yaml:"memory"`
					CPU    string `yaml:"cpu"`
				} `yaml:"limits"`
			} `yaml:"resources"`
		} `yaml:"prometheusCustomization"`
	} `yaml:"spec"`
}

func resourceBluePrint() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBluePrintCreate,
		ReadContext:   resourceBluePrintRead,
		UpdateContext: resourceBluePrintUpdate,
		DeleteContext: resourceBluePrintDelete,

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
			"yamlfileversion": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceBluePrintCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	log.Printf("YamlConfigFile %s", YamlConfigFilePath)

	if !utils.FileExists(YamlConfigFilePath) {
		log.Printf("file %s not exist", YamlConfigFilePath)
		return diags
	}
	if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
		log.Printf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath))
		return diags
	}
	f, err := os.Open(YamlConfigFilePath)
	if err != nil {
		log.Printf("Error while open Yaml %s", YamlConfigFilePath)
		return diag.FromErr(err)
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("error while Reading file")
		return diag.FromErr(err)
	}
	var b blueprintVersionYamlConfig
	err = yaml.Unmarshal(c, &b)
	if err != nil {
		log.Printf("error while unmarshal Yaml file ")
		return diag.FromErr(err)
	}
	if b.Metadata.Project == "" {
		log.Printf("project name should not be empty")
		return diags
	}
	// get project details
	log.Printf("project Name %s", b.Metadata.Project)
	resp, err := project.GetProjectByName(b.Metadata.Project)
	if err != nil {
		fmt.Print("project does not exist")
		return diag.FromErr(err)
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.FromErr(err)
	}
	addonDependency := make(map[string][]string)
	addons := make(map[string]string, len(b.Spec.Addons))
	for _, a := range b.Spec.Addons {
		if a.Version == "" {
			err = fmt.Errorf("version field is empty for addon %s", a.Name)
			return diag.FromErr(err)
		}
		if a.Name == "" {
			err = fmt.Errorf("name field is empty for addon version %s", a.Version)
			return diag.FromErr(err)
		}
		addons[a.Name] = a.Version
		addonDependency[a.Name] = a.DependsOn
	}
	log.Printf("addon len %d", len(b.Spec.Addons))
	if b.Spec.Blueprint == "" {
		err = fmt.Errorf(" Blueprint name cannot be empty ")
		return diag.FromErr(err)
	}
	if b.Metadata.Name == "" {
		err = fmt.Errorf(" Blueprint Metadataname cannot be empty ")
		return diag.FromErr(err)
	}
	if b.Spec.PspScope == "" {
		err = fmt.Errorf("psp scope must be supplied and must be one of %s or %s", ClusterScoped, NamespaceScoped)
		return diag.FromErr(err)
	} else if b.Spec.PspScope != ClusterScoped && b.Spec.PspScope != NamespaceScoped {
		err = fmt.Errorf("psp scope must be one of %s or %s, current value is %s", ClusterScoped, NamespaceScoped, b.Spec.PspScope)
		return diag.FromErr(err)
	}
	errCreate := blueprint.CreateBlueprint(b.Spec.Blueprint, project.ID)
	if errCreate != nil {
		log.Printf("Error While creating blueprint %s, %s", b.Spec.Blueprint, errCreate.Error())
		return diag.FromErr(errCreate)
	}
	errVersion := blueprint.CreateBlueprintVersion(b.Spec.Blueprint,
		project.ID, b.Metadata.Name, "", b.Spec.RafayIngress, b.Spec.RafayMonitoringAndAlerting,
		b.Spec.Kubevirt, addons, addonDependency, b.Spec.PspScope, b.Spec.Psps, b.Spec.DriftAction,
		b.Spec.PrometheusCustomization, "", "", "", "", "", false)
	if errVersion != nil {
		log.Printf("Error While creating blueprintversion %s, %s", b.Spec.Blueprint, errVersion.Error())
		return diag.FromErr(errVersion)
	}
	errpublish := blueprint.PublishBlueprint(b.Spec.Blueprint, b.Metadata.Name, "", project.ID)
	if errpublish != nil {
		log.Printf("Error While publish blueprintversion %s, %s", b.Spec.Blueprint, errpublish.Error())
		return diag.FromErr(errpublish)
	}
	d.SetId(b.Spec.Blueprint + "@" + b.Metadata.Name)
	return diags
}

func resourceBluePrintRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Read Blueprint %s", d.Id())
	s := strings.Split(d.Id(), "@")
	if len(s) < 2 {
		return diag.FromErr(fmt.Errorf("invalid blueprint Id"))
	}
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
	_, errGet := blueprint.GetBlueprint(s[0], project.ID)
	if errGet != nil {
		fmt.Printf("error while get blueprint %s", errGet.Error())
		return diag.FromErr(errGet)
	}
	if err := d.Set("name", s[0]); err != nil {
		log.Printf("set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceBluePrintUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	YamlConfigFilePath := d.Get("yamlfilepath").(string)
	log.Printf("YamlConfigFile %s", YamlConfigFilePath)

	if !utils.FileExists(YamlConfigFilePath) {
		log.Printf("file %s not exist", YamlConfigFilePath)
		return diags
	}
	if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
		log.Printf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath))
		return diags
	}
	f, err := os.Open(YamlConfigFilePath)
	if err != nil {
		log.Printf("Error while open Yaml %s", YamlConfigFilePath)
		return diag.FromErr(err)
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("error while Reading file")
		return diag.FromErr(err)
	}
	var b blueprintVersionYamlConfig
	err = yaml.Unmarshal(c, &b)
	if err != nil {
		log.Printf("error while unmarshal Yaml file ")
		return diag.FromErr(err)
	}
	if b.Metadata.Project == "" {
		log.Printf("project name should not be empty")
		return diags
	}
	// get project details
	log.Printf("project Name %s", b.Metadata.Project)
	resp, err := project.GetProjectByName(b.Metadata.Project)
	if err != nil {
		fmt.Print("project does not exist")
		return diag.FromErr(err)
	}
	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		fmt.Printf("project does not exist")
		return diag.FromErr(err)
	}
	addonDependency := make(map[string][]string)
	addons := make(map[string]string, len(b.Spec.Addons))
	for _, a := range b.Spec.Addons {
		if a.Version == "" {
			err = fmt.Errorf("version field is empty for addon %s", a.Name)
			return diag.FromErr(err)
		}
		if a.Name == "" {
			err = fmt.Errorf("name field is empty for addon version %s", a.Version)
			return diag.FromErr(err)
		}
		addons[a.Name] = a.Version
		addonDependency[a.Name] = a.DependsOn
	}
	log.Printf("addon len %d", len(b.Spec.Addons))
	if b.Spec.Blueprint == "" {
		err = fmt.Errorf(" Blueprint name cannot be empty ")
		return diag.FromErr(err)
	}
	if b.Metadata.Name == "" {
		err = fmt.Errorf(" Blueprint Metadataname cannot be empty ")
		return diag.FromErr(err)
	}
	if b.Spec.PspScope == "" {
		err = fmt.Errorf("psp scope must be supplied and must be one of %s or %s", ClusterScoped, NamespaceScoped)
		return diag.FromErr(err)
	} else if b.Spec.PspScope != ClusterScoped && b.Spec.PspScope != NamespaceScoped {
		err = fmt.Errorf("psp scope must be one of %s or %s, current value is %s", ClusterScoped, NamespaceScoped, b.Spec.PspScope)
		return diag.FromErr(err)
	}

	errVersion := blueprint.CreateBlueprintVersion(b.Spec.Blueprint, project.ID, b.Metadata.Name, "",
		b.Spec.RafayIngress, b.Spec.RafayMonitoringAndAlerting, b.Spec.Kubevirt,
		addons, addonDependency, b.Spec.PspScope, b.Spec.Psps, b.Spec.DriftAction,
		b.Spec.PrometheusCustomization, "", "", "", "", "", false)
	if errVersion != nil {
		log.Printf("Error While creating blueprintversion %s, %s", b.Spec.Blueprint, errVersion.Error())
		return diag.FromErr(errVersion)
	}
	errpublish := blueprint.PublishBlueprint(b.Spec.Blueprint, b.Metadata.Name, "", project.ID)
	if errpublish != nil {
		log.Printf("Error While publish blueprintversion %s, %s", b.Spec.Blueprint, errpublish.Error())
		return diag.FromErr(errpublish)
	}

	return diags
}

func resourceBluePrintDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("Read Blueprint1 %s", d.Get("name").(string))
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
	errDel := blueprint.DeleteBlueprint(d.Get("name").(string), project.ID)
	if errDel != nil {
		fmt.Printf("error while deleting blueprint %s", errDel.Error())
		return diag.FromErr(errDel)
	}
	return diags
}
