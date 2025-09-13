package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/RafaySystems/rctl/pkg/cluster"
	config "github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/terraform-provider-rafay/internal/resource_eks_cluster"
	"github.com/RafaySystems/terraform-provider-rafay/rafay"

	"github.com/RafaySystems/rctl/pkg/clusterctl"
	glogger "github.com/RafaySystems/rctl/pkg/log"
	"github.com/go-yaml/yaml"
)

var _ resource.Resource = (*eksClusterResource)(nil)

func NewEksClusterResource() resource.Resource {
	return &eksClusterResource{}
}

type eksClusterResource struct{}

func (r *eksClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_eks_cluster"
}

func (r *eksClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resource_eks_cluster.EksClusterResourceSchema(ctx)
}

func (r *eksClusterResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resource_eks_cluster.EksClusterModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ccList := make([]resource_eks_cluster.ClusterConfigValue, 0, len(data.ClusterConfig.Elements()))
	d := data.ClusterConfig.ElementsAs(ctx, &ccList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	cc := ccList[0]
	_ = cc

	if !cc.NodeGroups.IsNull() && !cc.NodeGroupsMap.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid Configuration",
			"Only one of 'node_groups' or 'node_groups_map' can be set at a time. Please remove one of them.",
		)
	}
}

func (r *eksClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newCluster, d := resource_eks_cluster.ExpandEksCluster(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	newClusterConfig, d := resource_eks_cluster.ExpandEksClusterConfig(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	tflog.Debug(ctx, "cluster value", map[string]any{"newCluster": newCluster, "newClusterConfig": newClusterConfig})

	// Specific to create flow: If `spec.sharing` specified then
	// set "cluster_sharing_external" to false.
	var cse string
	if newCluster.Spec != nil && newCluster.Spec.Sharing != nil {
		cse = "false"
	}

	// Call API to create EKS cluster
	clusterName := newCluster.Metadata.Name
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(newCluster); err != nil {
		log.Printf("error encoding cluster: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}
	if err := encoder.Encode(newClusterConfig); err != nil {
		log.Printf("error encoding cluster config: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster config, got error: %s", err))
		return
	}
	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	response, err := clusterctl.Apply(logger, rctlConfig, clusterName, b.Bytes(), false, false, false, false, uaDef, cse)
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}

	// wait until cluster is ready
	projectName := newCluster.Metadata.Project
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}
	log.Printf("process_filebytes response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse the cluster apply response, got error: %s", err))
		return
	}
	if res.TaskSetID == "" {
		return
	}
	time.Sleep(10 * time.Second)
	s, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
	if errGet != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
		return
	}

	data.Id = types.StringValue(s.ID)

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
LOOP:
	for {
		//Check for cluster operation timeout
		select {
		case <-ctx.Done():
			log.Println("Cluster operation stopped due to operation timeout.")
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("cluster operation stopped for cluster: `%s` due to operation timeout", clusterName))
			return
		case <-ticker.C:
			log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			check, errGet := cluster.GetCluster(newCluster.Metadata.Name, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
				return
			}
			edgeId := check.ID
			check, errGet = cluster.GetClusterWithEdgeID(edgeId, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
				return
			}
			rctlConfig.ProjectID = projectID
			statusResp, err := clusterctl.Status(logger, rctlConfig, res.TaskSetID)
			if err != nil {
				log.Println("status response parse error", err)
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster status, got error: %s", err))
				return
			}
			log.Println("statusResp:\n ", statusResp)
			sres := clusterCTLResponse{}
			err = json.Unmarshal([]byte(statusResp), &sres)
			if err != nil {
				log.Println("status response unmarshal error", err)
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse the cluster status response, got error: %s", err))
				return
			}
			if strings.Contains(sres.Status, "STATUS_COMPLETE") {
				log.Println("Checking in cluster conditions for blueprint sync success..")
				conditionsFailure, clusterReadiness, err := getClusterConditions(edgeId, projectID)
				if err != nil {
					log.Printf("error while getCluster %s", err.Error())
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster conditions, got error: %s", err))
					return
				}
				if conditionsFailure {
					log.Printf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName)
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName))
					return
				} else if clusterReadiness {
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", clusterName, projectName)
					break LOOP
				} else {
					log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
				}
			} else if strings.Contains(sres.Status, "STATUS_FAILED") {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("failed to create/update cluster while provisioning cluster %s %s", clusterName, statusResp))
				return
			} else {
				log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			}
		}
	}

	edgeDb, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get cluster", map[string]any{"name": clusterName, "pid": projectID})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", err))
		return
	}
	cseFromDb := edgeDb.Settings[clusterSharingExtKey]
	if cseFromDb != "true" {
		if newCluster.Spec.Sharing == nil && cseFromDb != "" {
			// reset cse as sharing is removed
			edgeDb.Settings[clusterSharingExtKey] = ""
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
				return
			}
			tflog.Error(ctx, "cse removed successfully")
		}
		if newCluster.Spec.Sharing != nil && cseFromDb != "false" {
			// explicitly set cse to false
			edgeDb.Settings[clusterSharingExtKey] = "false"
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
				return
			}
			tflog.Error(ctx, "cse set to false")
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "EKS Cluster Read existing data", map[string]interface{}{"data": data})

	clusterEls := make([]resource_eks_cluster.ClusterValue, 0, len(data.Cluster.Elements()))
	d := data.Cluster.ElementsAs(ctx, &clusterEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	if len(clusterEls) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Cluster block is missing")
		return
	}

	metadataEls := make([]resource_eks_cluster.MetadataValue, 0, len(clusterEls[0].Metadata.Elements()))
	d = clusterEls[0].Metadata.ElementsAs(ctx, &metadataEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	if len(metadataEls) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Metadata block is missing")
		return
	}

	mdO, d := metadataEls[0].ToObjectValue(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	var md resource_eks_cluster.MetadataType
	mdObj, d := md.ValueFromObject(ctx, mdO)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdValue, ok := mdObj.(resource_eks_cluster.MetadataValue)
	if !ok {
		resp.Diagnostics.AddError("Invalid Metadata", "Expected MetadataValue type but got a different type.")
		return
	}
	clusterName := mdValue.Name.ValueString()
	projectName := mdValue.Project.ValueString()
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}

	// Read API call logic
	c, err := cluster.GetCluster(clusterName, projectID, "")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", err))
		return
	}
	cse := c.Settings[clusterSharingExtKey]
	tflog.Info(ctx, "Got cluster from backend", map[string]interface{}{clusterSharingExtKey: cse})

	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlConfig, c.Name, projectID, uaDef)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster spec, got error: %s", err))
		return
	}
	tflog.Debug(ctx, "EKS Cluster Read API data", map[string]interface{}{"clusterSpecYaml": clusterSpecYaml})

	decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))
	clusterSpec := rafay.EKSCluster{}
	err = decoder.Decode(&clusterSpec)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode the cluster spec, got error: %s", err))
		return
	}

	// If the cluster sharing is managed by separate resource then
	// don't consider sharing from `rafay_eks_cluster`. Both
	// should not be present simultaneously.
	if cse == "true" {
		clusterSpec.Spec.Sharing = nil
	}

	clusterConfigSpec := rafay.EKSClusterConfig{}
	err = decoder.Decode(&clusterConfigSpec)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode the cluster config spec, got error: %s", err))
		return
	}
	tflog.Debug(ctx, "EKS Cluster Read API data", map[string]interface{}{
		"clusterSpec":       clusterSpec,
		"clusterConfigSpec": clusterConfigSpec,
	})

	// Update the model with the data from the API response
	diags := resource_eks_cluster.FlattenEksCluster(ctx, clusterSpec, &data)
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster, got error: %s", diags))
		return
	}

	diags = resource_eks_cluster.FlattenEksClusterConfig(ctx, clusterConfigSpec, &data)
	if diags.HasError() {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert the cluster config, got error: %s", diags))
		return
	}

	data.Id = types.StringValue(c.ID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterEls := make([]resource_eks_cluster.ClusterValue, 0, len(data.Cluster.Elements()))
	d := data.Cluster.ElementsAs(ctx, &clusterEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	if len(clusterEls) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Cluster block is missing")
		return
	}

	metadataEls := make([]resource_eks_cluster.MetadataValue, 0, len(clusterEls[0].Metadata.Elements()))
	d = clusterEls[0].Metadata.ElementsAs(ctx, &metadataEls, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	if len(metadataEls) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Metadata block is missing")
		return
	}

	mdO, d := metadataEls[0].ToObjectValue(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	var md resource_eks_cluster.MetadataType
	mdObj, d := md.ValueFromObject(ctx, mdO)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdValue, ok := mdObj.(resource_eks_cluster.MetadataValue)
	if !ok {
		resp.Diagnostics.AddError("Invalid Metadata", "Expected MetadataValue type but got a different type.")
		return
	}
	clusterName := mdValue.Name.ValueString()
	projectName := mdValue.Project.ValueString()
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}

	c, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", err))
		return
	}
	cse := c.Settings[clusterSharingExtKey]
	tflog.Info(ctx, "Got cluster from backend", map[string]interface{}{clusterSharingExtKey: cse})

	setSharing := false

	// Check if cse == true and `spec.sharing` specified. then
	// Error out here only before procedding. The next Upsert is
	// called by "Create" flow as well which is explicitly setting
	// cse to false if `spec.sharing` provided.
	if cse == "true" {
		// Load current state
		var state resource_eks_cluster.EksClusterModel
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

		specEls := make([]resource_eks_cluster.SpecValue, 0, len(clusterEls[0].Spec.Elements()))
		d = clusterEls[0].Spec.ElementsAs(ctx, &specEls, false)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		if len(specEls) < 1 {
			resp.Diagnostics.AddError("Invalid Configuration", "Spec block is missing")
			return
		}
		specO, d := specEls[0].ToObjectValue(ctx)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		var spec resource_eks_cluster.SpecType
		specObj, d := spec.ValueFromObject(ctx, specO)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		specValue, ok := specObj.(resource_eks_cluster.SpecValue)
		if !ok {
			resp.Diagnostics.AddError("Invalid Spec", "Expected SpecValue type but got a different type.")
			return
		}

		// state
		stClusterEls := make([]resource_eks_cluster.ClusterValue, 0, len(state.Cluster.Elements()))
		d = state.Cluster.ElementsAs(ctx, &stClusterEls, false)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}

		if len(stClusterEls) < 1 {
			resp.Diagnostics.AddError("Invalid Configuration", "Cluster block is missing")
			return
		}

		stSpecEls := make([]resource_eks_cluster.SpecValue, 0, len(stClusterEls[0].Spec.Elements()))
		d = stClusterEls[0].Spec.ElementsAs(ctx, &stSpecEls, false)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		if len(stSpecEls) < 1 {
			resp.Diagnostics.AddError("Invalid Configuration", "Spec block is missing")
			return
		}
		stSpecO, d := stSpecEls[0].ToObjectValue(ctx)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		var stSpec resource_eks_cluster.SpecType
		stSpecObj, d := stSpec.ValueFromObject(ctx, stSpecO)
		if d.HasError() {
			resp.Diagnostics.Append(d...)
			return
		}
		stSpecValue, ok := stSpecObj.(resource_eks_cluster.SpecValue)
		if !ok {
			resp.Diagnostics.AddError("Invalid Spec", "Expected SpecValue type but got a different type.")
			return
		}

		if !specValue.Sharing.Equal(stSpecValue.Sharing) {
			if !specValue.Sharing.IsNull() {
				resp.Diagnostics.AddError("Invalid Configuration", "Cluster sharing is currently managed through the external 'rafay_cluster_sharing' resource. To prevent configuration conflicts, please remove the sharing settings from the 'rafay_eks_cluster' resource and manage sharing exclusively via the external resource.")
				return
			}
		} else {
			// If the cluster sharing is managed externally, then (re-)populate sharing block.
			setSharing = true

		}
	}

	updatedCluster, d := resource_eks_cluster.ExpandEksCluster(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}

	if setSharing {
		logger := glogger.GetLogger()
		rctlCfg := config.GetConfig()
		clusterSpecYaml, err := clusterctl.GetClusterSpec(logger, rctlCfg, c.Name, projectID, uaDef)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster spec, got error: %s", err))
			return
		}
		log.Println("resourceEKSClusterUpdate clusterSpec ", clusterSpecYaml)

		decoder := yaml.NewDecoder(bytes.NewReader([]byte(clusterSpecYaml)))
		clusterSpec := rafay.EKSCluster{}
		if err := decoder.Decode(&clusterSpec); err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to decode the cluster spec, got error: %s", err))
			return
		}

		updatedCluster.Spec.Sharing = clusterSpec.Spec.Sharing
	}

	updatedClusterConfig, d := resource_eks_cluster.ExpandEksClusterConfig(ctx, data)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	tflog.Debug(ctx, "updated value", map[string]any{"updatedCluster": updatedCluster, "updatedClusterConfig": updatedClusterConfig})

	// Call API to update EKS cluster
	uClusterName := updatedCluster.Metadata.Name
	var b bytes.Buffer
	encoder := yaml.NewEncoder(&b)
	if err := encoder.Encode(updatedCluster); err != nil {
		log.Printf("error encoding cluster: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}
	if err := encoder.Encode(updatedClusterConfig); err != nil {
		log.Printf("error encoding cluster config: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster config, got error: %s", err))
		return
	}

	// Specific to create flow: If `spec.sharing` specified then
	// set "cluster_sharing_external" to false.
	var newCse string
	if updatedCluster.Spec != nil && updatedCluster.Spec.Sharing != nil {
		newCse = "false"
	}

	logger := glogger.GetLogger()
	rctlConfig := config.GetConfig()
	response, err := clusterctl.Apply(logger, rctlConfig, uClusterName, b.Bytes(), false, false, false, false, uaDef, newCse)
	if err != nil {
		log.Printf("cluster error 1: %s", err)
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to apply the cluster, got error: %s", err))
		return
	}

	// wait until cluster is ready
	log.Printf("process_filebytes response : %s", response)
	res := clusterCTLResponse{}
	err = json.Unmarshal([]byte(response), &res)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse the cluster apply response, got error: %s", err))
		return
	}
	if res.TaskSetID == "" {
		return
	}
	time.Sleep(10 * time.Second)
	s, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
	if errGet != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
		return
	}

	data.Id = types.StringValue(s.ID)

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()
LOOP:
	for {
		//Check for cluster operation timeout
		select {
		case <-ctx.Done():
			log.Println("Cluster operation stopped due to operation timeout.")
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("cluster operation stopped for cluster: `%s` due to operation timeout", clusterName))
			return
		case <-ticker.C:
			log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			check, errGet := cluster.GetCluster(updatedCluster.Metadata.Name, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
				return
			}
			edgeId := check.ID
			check, errGet = cluster.GetClusterWithEdgeID(edgeId, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s", errGet.Error())
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", errGet))
				return
			}
			rctlConfig.ProjectID = projectID
			statusResp, err := clusterctl.Status(logger, rctlConfig, res.TaskSetID)
			if err != nil {
				log.Println("status response parse error", err)
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster status, got error: %s", err))
				return
			}
			log.Println("statusResp:\n ", statusResp)
			sres := clusterCTLResponse{}
			err = json.Unmarshal([]byte(statusResp), &sres)
			if err != nil {
				log.Println("status response unmarshal error", err)
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse the cluster status response, got error: %s", err))
				return
			}
			if strings.Contains(sres.Status, "STATUS_COMPLETE") {
				log.Println("Checking in cluster conditions for blueprint sync success..")
				conditionsFailure, clusterReadiness, err := getClusterConditions(edgeId, projectID)
				if err != nil {
					log.Printf("error while getCluster %s", err.Error())
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster conditions, got error: %s", err))
					return
				}
				if conditionsFailure {
					log.Printf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName)
					resp.Diagnostics.AddError("Client Error", fmt.Sprintf("blueprint sync failed for edgename: %s and projectname: %s", clusterName, projectName))
					return
				} else if clusterReadiness {
					log.Printf("Cluster operation completed for edgename: %s and projectname: %s", clusterName, projectName)
					break LOOP
				} else {
					log.Println("Cluster Provisiong is Complete. Waiting for cluster to be Ready...")
				}
			} else if strings.Contains(sres.Status, "STATUS_FAILED") {
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("failed to create/update cluster while provisioning cluster %s %s", clusterName, statusResp))
				return
			} else {
				log.Printf("Cluster operation not completed for edgename: %s and projectname: %s. Waiting 60 seconds more for cluster to complete the operation.", clusterName, projectName)
			}
		}
	}

	edgeDb, err := cluster.GetCluster(clusterName, projectID, uaDef)
	if err != nil {
		tflog.Error(ctx, "failed to get cluster", map[string]any{"name": clusterName, "pid": projectID})
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get the cluster, got error: %s", err))
		return
	}
	cseFromDb := edgeDb.Settings[clusterSharingExtKey]
	if cseFromDb != "true" {
		if updatedCluster.Spec.Sharing == nil && cseFromDb != "" {
			// reset cse as sharing is removed
			edgeDb.Settings[clusterSharingExtKey] = ""
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
				return
			}
			tflog.Error(ctx, "cse removed successfully")
		}
		if updatedCluster.Spec.Sharing != nil && cseFromDb != "false" {
			// explicitly set cse to false
			edgeDb.Settings[clusterSharingExtKey] = "false"
			err := cluster.UpdateCluster(edgeDb, uaDef)
			if err != nil {
				tflog.Error(ctx, "failed to update cluster", map[string]any{"edgeObj": edgeDb})
				resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update the edge object, got error: %s", err))
				return
			}
			tflog.Error(ctx, "cse set to false")
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *eksClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data resource_eks_cluster.EksClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get cluster from state
	clusterList := make([]resource_eks_cluster.ClusterValue, 0, len(data.Cluster.Elements()))
	d := data.Cluster.ElementsAs(ctx, &clusterList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	if len(clusterList) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Cluster block is missing")
		return
	}
	mdList := make([]resource_eks_cluster.MetadataValue, 0, len(clusterList[0].Metadata.Elements()))
	d = clusterList[0].Metadata.ElementsAs(ctx, &mdList, false)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	if len(mdList) < 1 {
		resp.Diagnostics.AddError("Invalid Configuration", "Metadagta block is missing")
		return
	}
	mdO, d := mdList[0].ToObjectValue(ctx)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	var md resource_eks_cluster.MetadataType
	mdObj, d := md.ValueFromObject(ctx, mdO)
	if d.HasError() {
		resp.Diagnostics.Append(d...)
		return
	}
	mdValue, ok := mdObj.(resource_eks_cluster.MetadataValue)
	if !ok {
		resp.Diagnostics.AddError("Invalid Metadata", "Expected MetadataValue type but got a different type.")
		return
	}
	clusterName := mdValue.Name.ValueString()
	projectName := mdValue.Project.ValueString()
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to get project ID from name '%s', got error: %s", projectName, err))
		return
	}

	// Delete API call logic
	err = cluster.DeleteCluster(clusterName, projectID, false, uaDef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Client Error",
			fmt.Sprintf("Failed to delete cluster, got error: %s", err),
		)
	}

	ticker := time.NewTicker(time.Duration(60) * time.Second)
	defer ticker.Stop()

LOOP:
	for {
		select {
		case <-ctx.Done():
			log.Printf("Cluster Deletion for edgename: %s and projectname: %s got timeout out.", clusterName, projectName)
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("cluster deletion for edgename: %s and projectname: %s got timeout out", clusterName, projectName))
			return
		case <-ticker.C:
			check, errGet := cluster.GetCluster(clusterName, projectID, uaDef)
			if errGet != nil {
				log.Printf("error while getCluster %s, delete success", errGet.Error())
				break LOOP
			}
			if check == nil {
				break LOOP
			}
			log.Printf("Cluster Deletion is in progress for edgename: %s and projectname: %s. Waiting 60 seconds more for operation to complete.", clusterName, projectName)
		}
	}
	log.Printf("Cluster Deletion completes for edgename: %s and projectname: %s", clusterName, projectName)

}
