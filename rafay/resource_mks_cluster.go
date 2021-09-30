package rafay

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
	"sync"
	"path/filepath"
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/commands"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/RafaySystems/rctl/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMKSCluster() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMKSClusterCreate,
		ReadContext:   resourceMKSClusterRead,
		UpdateContext: resourceMKSClusterUpdate,
		DeleteContext: resourceMKSClusterDelete,

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
			"projectname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"waitflag": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceMKSClusterCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("create MKS cluster resource")
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

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	if !utils.FileExists(YamlConfigFilePath) {
		return diag.FromErr(fmt.Errorf("file %s does not exist", YamlConfigFilePath))
	}
	// make sure the file is a yaml file
	if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
		return diag.FromErr(fmt.Errorf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath)))
	}

	f, err := os.Open(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("file %s open Error", YamlConfigFilePath))
	}
	fc, errc := ioutil.ReadAll(f)
	if errc != nil {
		return diag.FromErr(fmt.Errorf("file %s open Error", YamlConfigFilePath))
	}

	var c commands.ClusterMKSYamlConfig
	err = yaml.Unmarshal(fc, &c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error While Yaml unmarshal "))
	}

	if c.Kind != "Cluster" {
		return diag.FromErr(fmt.Errorf("please provide a correct yaml config for resource cluster, kind was %s", c.Kind))
	}

	if c.Spec.Type != "mks" {
		return diag.FromErr(fmt.Errorf("cluster types in file is not mks, type is %s", c.Spec.Type))
	}

	if c.Metadata.Name == "" {
		return diag.FromErr(fmt.Errorf("clusterName can not be emtpy" ))
	}
	if c.Metadata.Name != d.Get("name").(string) {
		return diag.FromErr(fmt.Errorf("ClusterConfig name does not match config file"))
	}

	if c.Metadata.Project == "" {
		return diag.FromErr(fmt.Errorf("Project can not be empty" ))
	}
	if c.Metadata.Project != d.Get("projectname").(string) {
		return diag.FromErr(fmt.Errorf("project name does not match config file" ))
	}

	resp, errcluster := cluster.NewImportClusterMKS(c.Metadata.Name, c.Spec.Blueprint, c.Spec.Config.Location, project.ID, c.Spec.Config.K8sversion, c.Spec.Config.OperatingSystem, c.Spec.Config.DefaultStorageClass, c.Spec.Config.StorageClassMap[0].StoragePath,c.Spec.Config.StorageClassMap[1].StoragePath  )
	if errcluster != nil {
		return diag.FromErr(fmt.Errorf("cluster creation failed: %v",errcluster ))
	}
	clusterResp := models.Edge{}
	errs := json.Unmarshal([]byte(resp), &clusterResp)
	if errs != nil {
		return diag.FromErr(fmt.Errorf("there was an error while json unmarshalling: %v",errs))
	}

	if c.Spec.Config.AutoApproveNodes {
		clusterResp.AutoApproveNodes = true
	}

	if  c.Spec.Proxy.Enabled {
		ProxyConfig := &models.ProxyConfig {
			Enabled :  true,
			HttpProxy : c.Spec.Proxy.HttpProxy,
			HttpsProxy : c.Spec.Proxy.HttpsProxy,
			NoProxy : c.Spec.Proxy.NoProxy,
			BootstrapCA : c.Spec.Proxy.RootCA,
		}
		clusterResp.ProxyConfig = ProxyConfig
	}

	clusterResp.ProvisionParams["state"] = "PROVISION"
	resp, err = cluster.NewImportClusterMKSConfig(clusterResp )
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error While cluster config: %v",err) )
	}
	time.Sleep(30 * time.Second )
	err = commands.NodesWork(clusterResp , c )
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error nodeWork(): %v",err))
	}
	time.Sleep(60 * time.Second )

	getClusterResp, errGet := cluster.GetCluster(c.Metadata.Name, project.ID )
	if errGet != nil {
		return diag.FromErr(fmt.Errorf("GetCluster error: %v", errGet))
	}
	var wg sync.WaitGroup
	if !getClusterResp.AutoApproveNodes {
		for  _, nodedata := range getClusterResp.Nodes {
			wg.Add(1)
			go func(Nodedata interface{} ) error {
				defer wg.Done()
				jsonbody, errj := json.Marshal(Nodedata)
				if errj != nil {
					return fmt.Errorf("json marshal : %v\n",errj)
				}
				singlenodedata := models.NodeData{}
				errj = json.Unmarshal(jsonbody, &singlenodedata )
				if errj != nil {
					return fmt.Errorf("json unmarshal :%v\n",errj )
				}
				err = commands.ApproveNode(getClusterResp.ProjectID, getClusterResp.ID, singlenodedata.Id, c.Metadata.Name )
				if err != nil {
					return fmt.Errorf("error on ApproveNode() %v",err)
				}
				return nil
			}(nodedata)
		}
	}
	wg.Wait()

	getClusterResp, errGet = cluster.GetCluster(c.Metadata.Name, project.ID )
	if errGet != nil {
		return diag.FromErr(fmt.Errorf("GetCluster error: %v", errGet))
	}
	var wgconfig sync.WaitGroup
	for  _, nodedata := range getClusterResp.Nodes {
		wgconfig.Add(1)
		go func(Nodedata interface{} ) error {
			defer wgconfig.Done()
			jsonbody, errj := json.Marshal(Nodedata)
			if errj != nil {
				return fmt.Errorf("json marshal : %v\n",errj)
			}
			singlenodedata := models.NodeData{}
			errj = json.Unmarshal(jsonbody, &singlenodedata )
			if errj != nil {
				return fmt.Errorf("json unmarshal :%v\n",errj )
			}
			err = commands.WaitUntilNodeApprove( c.Metadata.Name , getClusterResp.ProjectID ,singlenodedata.Id )
			if err != nil {
				return fmt.Errorf("error while waiting approve: %v\n",err )
			}
			for _, nodeinfo := range c.Spec.Config.Nodes {
				if singlenodedata.Hostname == nodeinfo.HostName {
					singlenodedata.Roles = make([]string, len(nodeinfo.Roles))
					copy(singlenodedata.Roles, nodeinfo.Roles)
					err = commands.ConfigureNode(getClusterResp.ProjectID, getClusterResp.ID, singlenodedata.Id,  singlenodedata  )
					if err != nil {
						return fmt.Errorf("error on configureNode(): %v",err)
					}
				}
			}
			return nil
		}(nodedata)
	}
	wgconfig.Wait()

	err = commands.ProvisionNodes(getClusterResp.ProjectID, getClusterResp.ID )
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error while provision %v",err))
	}

	s, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
		return diag.FromErr(err)
	}
	if d.Get("waitflag").(string) == "1" {
		log.Printf("Cluster Provision may take upto 15-20 Minutes")
		for {
			check, errGet := cluster.GetCluster(d.Get("name").(string), project.ID)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				return diag.FromErr(errGet)
			}
			if check.Status == "READY" {
				break
			}
			if strings.Contains(check.Provision.Status, "FAILED") {
				return diag.FromErr(fmt.Errorf("Failed to create cluster while cluster provisioning"))
			}
			time.Sleep(40 * time.Second)
		}
	}

	log.Printf("resource eks cluster created %s", s.ID)
	d.SetId(s.ID)

	return diags
}

func resourceMKSClusterRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	c, err := cluster.GetCluster(d.Get("name").(string), project.ID)
	if err != nil {
		log.Printf("error in get cluster %s", err.Error())
		return diag.FromErr(err)
	}
	if err := d.Set("name", c.Name); err != nil {
		log.Printf("get group set name error %s", err.Error())
		return diag.FromErr(err)
	}

	return diags
}

func resourceMKSClusterUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("update MKS cluster resource")

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

	YamlConfigFilePath := d.Get("yamlfilepath").(string)

	if !utils.FileExists(YamlConfigFilePath) {
		return diag.FromErr(fmt.Errorf("file %s does not exist", YamlConfigFilePath))
	}

	// make sure the file is a yaml file
	if filepath.Ext(YamlConfigFilePath) != ".yml" && filepath.Ext(YamlConfigFilePath) != ".yaml" {
		return diag.FromErr(fmt.Errorf("file must a yaml file, file type is %s", filepath.Ext(YamlConfigFilePath)))
	}
	f, err := os.Open(YamlConfigFilePath)
	if err != nil {
		return diag.FromErr(fmt.Errorf("file %s open Error", YamlConfigFilePath))
	}
	fc, errc := ioutil.ReadAll(f)
	if errc != nil {
		return diag.FromErr(fmt.Errorf("file %s open Error", YamlConfigFilePath))
	}

	var c commands.ClusterMKSYamlConfig
	err = yaml.Unmarshal(fc, &c)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error While Yaml unmarshal "))
	}

	if c.Kind != "Cluster" {
		return diag.FromErr(fmt.Errorf("please provide a correct yaml config for resource cluster, kind was %s", c.Kind))
	}

	if c.Spec.Type != "mks" {
		return diag.FromErr(fmt.Errorf("cluster types in file is not mks, type is %s", c.Spec.Type))
	}

	if c.Metadata.Name == "" {
		return diag.FromErr(fmt.Errorf("clusterName can not be emtpy" ))
	}
	if c.Metadata.Name != d.Get("name").(string) {
		return diag.FromErr(fmt.Errorf("ClusterConfig name does not match config file"))
	}

	if c.Metadata.Project == "" {
		return diag.FromErr(fmt.Errorf("Project can not be empty" ))
	}
	if c.Metadata.Project != d.Get("projectname").(string) {
		return diag.FromErr(fmt.Errorf("project name does not match config file" ))
	}

	getClusterResp, errGet := cluster.GetCluster(c.Metadata.Name, project.ID )
	if errGet != nil {
		return diag.FromErr(fmt.Errorf("GetCluster error: %v", errGet))
	}
	var NodeFound bool
	for _, nodeinfo := range c.Spec.Config.Nodes {
		NodeFound = true
		for _, nodedata := range getClusterResp.Nodes {
			jsonbody, errj := json.Marshal(nodedata)
			if errj != nil {
				return diag.FromErr(fmt.Errorf("json marshal : %v\n",errj))
			}
			singlenodedata := models.NodeData{}
			errj = json.Unmarshal(jsonbody, &singlenodedata )
			if errj != nil {
				return diag.FromErr(fmt.Errorf("json unmarshal :%v\n",errj ))
			}
			if nodeinfo.HostName == singlenodedata.Hostname {
				NodeFound = false
				if singlenodedata.Status == "READY" {
					break
				}
				//TODO:if Any pending work for Node
			}
		}
		if NodeFound {
			var nodes commands.Nodes
			nodes.HostName = nodeinfo.HostName
			nodes.Ipaddress = nodeinfo.Ipaddress
			nodes.Ipv6address = nodeinfo.Ipv6address
			nodes.SSHPrivateKeyPath = nodeinfo.SSHPrivateKeyPath
			nodes.SSHPort = nodeinfo.SSHPort
			nodes.SSHUserName = nodeinfo.SSHUserName
			nodes.Roles = make([]string, len(nodeinfo.Roles))
			copy(nodes.Roles, nodeinfo.Roles)
			err :=  commands.SingleNodeWork(getClusterResp, nodes, project.ID)
			if err != nil {
				return diag.FromErr(fmt.Errorf("Error while updating cluster with new node %s: %v\n",nodeinfo.HostName, err ))
			}
		}
	}

	return diags
}

func resourceMKSClusterDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	errDel := cluster.DeleteCluster(d.Get("name").(string), project.ID)
	if errDel != nil {
		log.Printf("delete cluster error %s", errDel.Error())
		return diag.FromErr(errDel)
	}

	return diags
}
