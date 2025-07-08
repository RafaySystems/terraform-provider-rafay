package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/user"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/pkg/errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type namespaceSpecTranspose struct {
	Psp                       *infrapb.NamespacePSP              `protobuf:"bytes,1,opt,name=psp,proto3" json:"psp,omitempty"`
	Placement                 *commonpb.PlacementSpec            `protobuf:"bytes,2,opt,name=placement,proto3" json:"placement,omitempty"`
	Drift                     *commonpb.DriftSpec                `protobuf:"bytes,3,opt,name=drift,proto3" json:"drift,omitempty"`
	NetworkPolicyParms        *infrapb.NetworkPolicyParams       `protobuf:"bytes,4,opt,name=network_policy_params,proto3" json:"networkPolicyParams,omitempty"`
	NamespaceMeshPolicyParams *infrapb.NamespaceMeshPolicyParams `protobuf:"bytes,4,opt,name=namespace_mesh_policy_params,proto3" json:"namespaceMeshPolicyParams,omitempty"`
	ResourceQuotas            *infrapb.NamespaceResourceQuotas   `protobuf:"bytes,4,opt,name=resourceQuotas,proto3" json:"resourceQuotas,omitempty"`
	LimitRange                *infrapb.NamespaceLimitRange       `protobuf:"bytes,5,opt,name=limitRange,proto3" json:"limitRange,omitempty"`
	// Types that are assignable to Artifact:
	//	*NamespaceSpec_Uploaded
	//	*NamespaceSpec_Repo
	//Artifact isNamespaceSpec_Artifact `protobuf_oneof:"artifact"`

	Artifact struct {
		Repository string `protobuf:"bytes,1,opt,name=repository,proto3" json:"repository,omitempty"`
		Revision   string `protobuf:"bytes,2,opt,name=revision,proto3" json:"revision,omitempty"`
		Path       *File  `protobuf:"bytes,3,opt,name=path,proto3" json:"path,omitempty"`
	} `json:"artifact,omitempty"`
}

func resourceNamespace() *schema.Resource {
	modSchema := resource.NamespaceSchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
	return &schema.Resource{
		CreateContext: resourceNamespaceCreate,
		ReadContext:   resourceNamespaceRead,
		UpdateContext: resourceNamespaceUpdate,
		DeleteContext: resourceNamespaceDelete,
		Importer: &schema.ResourceImporter{
			State: resourceNamespaceImport,
		},
		// Set timeouts of 17 mins for Create and Update functions as it takes approx 15 mins for the namespace publish to be marked as failed.
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(17 * time.Minute),
			Update: schema.DefaultTimeout(17 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

func resourceNamespaceImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceNamespaceImport idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceNamespaceImport d.Id:", d.Id())
	log.Println("resourceNamespaceImport d_debug", d_debug)

	namespace, err := expandNamespace(d)
	if err != nil {
		log.Printf("namespace expandNamespace error")
		//return nil, err
	}
	log.Println("import1")
	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	namespace.Metadata = &metaD
	log.Println("import pre flatten")
	err = d.Set("metadata", flattenMetaData(namespace.Metadata))
	if err != nil {
		log.Println("import set err")
		return nil, err
	}
	log.Println("import post flatten")
	d.SetId(namespace.Metadata.Name)
	log.Println("import post set id")
	return []*schema.ResourceData{d}, nil
}

func resourceNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespace create starts")
	create := isNamespaceAlreadyExists(ctx, d)
	diags := resourceNamespaceUpsert(ctx, d, m)

	if diags.HasError() && !create {
		if checkStandardInputTextError(diags[0].Summary) {
			return diags
		}
		namespaceCreateError := diag.FromErr(fmt.Errorf("%s", diags[0].Summary))
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespace create got error, perform cleanup")
		ns, err := expandNamespace(d)
		if err != nil {
			log.Printf("namespace expandNamespace error")
			return namespaceCreateError
		}
		if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
			defer ResetImpersonateUser()
			asUser := d.Get("impersonate").(string)
			// check user role : impersonation not allowed for a user
			// with ORG Admin role
			isOrgAdmin, err := user.IsOrgAdmin(asUser)
			if err != nil {
				return namespaceCreateError
			}
			if isOrgAdmin {
				return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
			}
			config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
			if err != nil {
				return namespaceCreateError
			}
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diag.FromErr(err)
		}

		err = client.InfraV3().Namespace().Delete(ctx, options.DeleteOptions{
			Name:    ns.Metadata.Name,
			Project: ns.Metadata.Project,
		})
		if err != nil {
			log.Println("Error while namespace cleanup :", err.Error())
			return namespaceCreateError
		}
	}
	return diags
}

func resourceNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespace update starts")
	return resourceNamespaceUpsert(ctx, d, m)
}

func resourceNamespaceUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespace upsert starts")

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	if d.State() != nil && d.State().ID != "" {
		n := GetMetaName(d)
		if n != "" && n != d.State().ID {
			log.Printf("metadata name change not supported")
			d.State().Tainted = true
			return diag.FromErr(fmt.Errorf("%s", "metadata name change not supported"))
		}
	}

	ns, err := expandNamespace(d)
	if err != nil {
		log.Printf("namespace expandNamespace error")
		return diag.FromErr(err)
	}

	err = validateResourceName(ns.Metadata.Name)
	if err != nil {
		return diag.FromErr(err)
	}

	artifactNames, err := ns.ArtifactList()
	if err != nil {
		err = errors.Wrapf(err, "unable to list artifacts")
		return diag.FromErr(err)
	}
	log.Println(" resourceNamespaceUpsert artifactNames ", artifactNames)

	for _, artifactName := range artifactNames {

		if !strings.HasPrefix(artifactName, "file://") {
			continue
		}
		//get full path of artifact
		artifactFullPath := filepath.Join(filepath.Dir("."), artifactName[7:])
		//retrieve artifact data
		artifactData, err := ioutil.ReadFile(artifactFullPath)
		if err != nil {
			err = fmt.Errorf("unable to read artifact at '%s'", artifactFullPath)
			return diag.FromErr(err)
		}
		//set artifact in namespace
		err = ns.ArtifactSet(artifactName, artifactData)
		if err != nil {
			err = fmt.Errorf("unable to set artifact %s at path '%s'", artifactName, artifactFullPath)
			return diag.FromErr(err)
		}
	}

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	w1 := spew.Sprintf("%+v", ns)
	log.Println("resourceNamespaceUpsert ns:", w1)
	err = client.InfraV3().Namespace().Apply(ctx, ns, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		// n1 := spew.Sprintf("%+v", ns)
		// log.Println("namespace apply ns:", n1)
		log.Println("namespace apply error:", err)
		return diag.FromErr(err)
	}

	// wait for publish
	for {
		ctxDeadlineTime, _ := ctx.Deadline()
		timeRemaining := time.Until(ctxDeadlineTime)

		nsStatus, err := client.InfraV3().Namespace().Status(ctx, options.StatusOptions{
			Name:    ns.Metadata.Name,
			Project: ns.Metadata.Project,
		})
		if err != nil {
			return diag.FromErr(err)
		}

		if timeRemaining < 30*time.Second {
			if len(nsStatus.Status.AssignedClusters) != (len(nsStatus.Status.FailedClusters) + len(nsStatus.Status.ReadyClusters)) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Warning,
					Summary:  "Failed to patch namespace",
					Detail:   fmt.Errorf("namespace: %s patch may not be complete", ns.Metadata.Name).Error(),
				})
				d.SetId(ns.Metadata.Name)
				return diags
			}
			break
		}

		//check if namespace can be placed on a cluster, if true break out of infinite loop
		if nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusOK ||
			nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusNotSet {
			break
		}

		if nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusPartiallyReady {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Failed to patch namespace",
				Detail:   fmt.Errorf("%s", nsStatus.Status.Reason).Error(),
			})
			d.SetId(ns.Metadata.Name)
			return diags
		}

		if nsStatus.Status.ConditionStatus == commonpb.ConditionStatus_StatusFailed && len(nsStatus.Status.AssignedClusters) == len(nsStatus.Status.FailedClusters) {
			return diag.FromErr(fmt.Errorf("%s to %s", "failed to publish namespace", nsStatus.Status.Reason))
		}

		time.Sleep(30 * time.Second)
	}
	d.SetId(ns.Metadata.Name)
	return diags
}

func resourceNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespace read starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	nsTFState, err := expandNamespace(d)
	if err != nil {
		log.Println("expandNamespace err:", err)
		//return diag.FromErr(err)
	}

	// XXX Debug
	n1 := spew.Sprintf("%+v", nsTFState)
	log.Println("resourceNamespaceRead nsTFState ", n1)

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ns, err := client.InfraV3().Namespace().Get(ctx, options.GetOptions{
		//Name:    nsTFState.Metadata.Name,
		Name:    meta.Name,
		Project: nsTFState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	log.Println("resourceNamespaceRead remoteState ", ns)
	// XXX Debug
	// n1 = spew.Sprintf("%+v", ns)
	// log.Println("resourceNamespaceRead ns", n1)
	/*
		if ns.Spec.ResourceQuotas != nil {
			log.Println("resourceNamespaceRead ns.Spec.ResourceQuotas Memory", ns.Spec.ResourceQuotas.MemoryRequests)
		}*/

	err = flattenNamespace(d, ns)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nsTFState, err := expandNamespace(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// n1 := spew.Sprintf("%+v", nsTFState)
	// log.Println("resourceNamespaceRead nsTFState", n1)

	if v, ok := d.Get("impersonate").(string); ok && len(v) > 0 {
		defer ResetImpersonateUser()
		asUser := d.Get("impersonate").(string)
		// check user role : impersonation not allowed for a user
		// with ORG Admin role
		isOrgAdmin, err := user.IsOrgAdmin(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
		if isOrgAdmin {
			return diag.FromErr(fmt.Errorf("%s", "--as-user cannot have ORGADMIN role"))
		}
		config.ApiKey, config.ApiSecret, err = user.GetUserAPIKey(asUser)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Namespace().Delete(ctx, options.DeleteOptions{
		Name:    nsTFState.Metadata.Name,
		Project: nsTFState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespace(in *schema.ResourceData) (*infrapb.Namespace, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespace empty input")
	}
	obj := &infrapb.Namespace{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}
	w1 := spew.Sprintf("%+v", obj.Metadata)
	log.Println("expandNamespace metadata ", w1)

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandNamespaceSpec(v)
		if err != nil {
			return obj, err
		}
		log.Println("expandNamespace got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Namespace"
	return obj, nil
}

func expandNamespaceSpec(p []interface{}) (*infrapb.NamespaceSpec, error) {
	obj := infrapb.NamespaceSpec{}
	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandNamespaceSpec empty input")
	}

	nst := namespaceSpecTranspose{}

	in := p[0].(map[string]interface{})
	if v, ok := in["psp"].([]interface{}); ok {
		nst.Psp = expandPsp(v)
	}

	if v, ok := in["network_policy_params"].([]interface{}); ok {
		nst.NetworkPolicyParms = expandNetworkPolicyParams(v)
	}

	if v, ok := in["namespace_mesh_policy_params"].([]interface{}); ok {
		nst.NamespaceMeshPolicyParams = expandNamespaceMeshPolicyParams(v)
	}

	if v, ok := in["placement"].([]interface{}); ok {
		nst.Placement = expandPlacement(v)
	}

	if v, ok := in["drift"].([]interface{}); ok {
		nst.Drift = expandDrift(v)
	}

	if v, ok := in["resource_quotas"].([]interface{}); ok {
		log.Println("resource_quotas v", v)
		nst.ResourceQuotas = expandNamespaceResourceQuotas(v)
		log.Println("nst.ResourceQuotas ", nst.ResourceQuotas)
	}

	if v, ok := in["limit_range"].([]interface{}); ok {
		nst.LimitRange = expandNamespaceLimitRange(v)
	}

	if vp, ok := in["artifact"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			log.Println("expandArtifact empty artifact")
		} else {
			ina := vp[0].(map[string]interface{})

			if v, ok := ina["path"].([]interface{}); ok && len(v) > 0 {
				nst.Artifact.Path = expandFile(v)
			}

			if v, ok := ina["repository"].(string); ok && len(v) > 0 {
				nst.Artifact.Repository = v
			}

			if v, ok := ina["revision"].(string); ok && len(v) > 0 {
				nst.Artifact.Revision = v
			}
		}
	}
	// XXX Debug
	s := spew.Sprintf("%+v", nst)
	log.Println("expandNamespaceSpec nst", s)

	jsonSpec, err := json.Marshal(nst)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("expandNamespaceSpec jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandNamespaceSpec artifact UnmarshalJSON error ", err)
		return nil, err
	}
	/*
		if obj.ResourceQuotas != nil && obj.ResourceQuotas.Requests != nil && obj.ResourceQuotas.Requests.Memory != nil {
			log.Println("expandNamespaceSpec Obj Memory ", obj.ResourceQuotas.Requests.Memory)
		}
	*/
	log.Println("expandNamespaceSpec Obj", obj)
	return &obj, nil
}

func expandPsp(p []interface{}) *infrapb.NamespacePSP {
	obj := &infrapb.NamespacePSP{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok {
		obj.Name = v
	}

	return obj
}

func expandNetworkPolicyParams(p []interface{}) *infrapb.NetworkPolicyParams {
	obj := &infrapb.NetworkPolicyParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["network_policy_enabled"].(bool); ok {
		obj.NetworkPolicyEnabled = v
	}

	if v, ok := in["policies"].([]interface{}); ok && len(v) > 0 {
		obj.Policies = expandNamespaceNetworkPolicyPolicies(v)
	}

	return obj
}

func expandNamespaceNetworkPolicyPolicies(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceNameAndVersionRef{}
	}

	out := make([]*commonpb.ResourceNameAndVersionRef, len(p))

	for i := range p {
		obj := commonpb.ResourceNameAndVersionRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out

}

func expandNamespaceMeshPolicyParams(p []interface{}) *infrapb.NamespaceMeshPolicyParams {
	obj := &infrapb.NamespaceMeshPolicyParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["mesh_enabled"].(bool); ok {
		obj.MeshEnabled = v
	}

	if v, ok := in["policies"].([]interface{}); ok && len(v) > 0 {
		obj.Policies = expandNamespaceMeshPolicies(v)
	}

	return obj
}

func expandNamespaceMeshPolicies(p []interface{}) []*commonpb.ResourceNameAndVersionRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceNameAndVersionRef{}
	}

	out := make([]*commonpb.ResourceNameAndVersionRef, len(p))

	for i := range p {
		obj := commonpb.ResourceNameAndVersionRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok {
			obj.Version = v
		}

		out[i] = &obj
	}

	return out

}

func expandNamespaceResourceQuotas(p []interface{}) *infrapb.NamespaceResourceQuotas {
	obj := &infrapb.NamespaceResourceQuotas{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cpu_requests"].(string); ok && len(v) > 0 {
		//obj.CpuRequests = expandQuantityString(v)
		obj.CpuRequests = v
	}

	if v, ok := in["memory_requests"].(string); ok && len(v) > 0 {
		//obj.MemoryRequests = expandQuantityString(v)
		obj.MemoryRequests = v
	}

	if v, ok := in["cpu_limits"].(string); ok && len(v) > 0 {
		//obj.CpuLimits = expandQuantityString(v)
		obj.CpuLimits = v
	}

	if v, ok := in["gpu_requests"].(string); ok && len(v) > 0 {
		obj.GpuRequests = v
	}

	if v, ok := in["gpu_limits"].(string); ok && len(v) > 0 {
		obj.GpuLimits = v
	}

	if v, ok := in["memory_limits"].(string); ok && len(v) > 0 {
		//obj.MemoryLimits = expandQuantityString(v)
		obj.MemoryLimits = v
	}

	if v, ok := in["config_maps"].(string); ok && len(v) > 0 {
		//obj.ConfigMaps = expandQuantityString(v)
		obj.ConfigMaps = v
	}

	if v, ok := in["persistent_volume_claims"].(string); ok && len(v) > 0 {
		//obj.PersistentVolumeClaims = expandQuantityString(v)
		obj.PersistentVolumeClaims = v
	}

	if v, ok := in["secrets"].(string); ok && len(v) > 0 {
		//obj.Secrets = expandQuantityString(v)
		obj.Secrets = v
	}

	if v, ok := in["services"].(string); ok && len(v) > 0 {
		//obj.Services = expandQuantityString(v)
		obj.Services = v
	}

	if v, ok := in["services_load_balancers"].(string); ok && len(v) > 0 {
		//obj.ServicesLoadBalancers = expandQuantityString(v)
		obj.ServicesLoadBalancers = v
	}

	if v, ok := in["services_node_ports"].(string); ok && len(v) > 0 {
		//obj.ServicesNodePorts = expandQuantityString(v)
		obj.ServicesNodePorts = v
	}

	if v, ok := in["storage_requests"].(string); ok && len(v) > 0 {
		//obj.StorageRequests = expandQuantityString(v)
		obj.StorageRequests = v
	}

	if v, ok := in["pods"].(string); ok && len(v) > 0 {
		//obj.Pods = expandQuantityString(v)
		obj.Pods = v
	}

	if v, ok := in["replication_controllers"].(string); ok && len(v) > 0 {
		//obj.ReplicationControllers = expandQuantityString(v)
		obj.ReplicationControllers = v
	}

	if v, ok := in["ephemeral_storage_limits"].(string); ok && len(v) > 0 {
		//obj.ReplicationControllers = expandQuantityString(v)
		obj.EphemeralStorageLimits = v
	}

	if v, ok := in["ephemeral_storage_requests"].(string); ok && len(v) > 0 {
		//obj.ReplicationControllers = expandQuantityString(v)
		obj.EphemeralStorageRequests = v
	}

	log.Println("expandNamespaceResourceQuotas obj ", obj)
	return obj
}

func expandNamespaceLimitRange(p []interface{}) *infrapb.NamespaceLimitRange {
	obj := &infrapb.NamespaceLimitRange{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	//log.Println("expandNamespaceLimitRange")
	in := p[0].(map[string]interface{})
	if v, ok := in["pod"].([]interface{}); ok {
		obj.Pod = expandNamespaceLimitRangeConfig(v)
	}

	if v, ok := in["container"].([]interface{}); ok {
		obj.Container = expandNamespaceLimitRangeConfig(v)
	}

	return obj
}

func expandNamespaceLimitRangeConfig(p []interface{}) *infrapb.NamespaceLimitRangeConfig {
	obj := &infrapb.NamespaceLimitRangeConfig{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["min"].([]interface{}); ok {
		obj.Min = expandResourceQuantityString(v)
	}

	if v, ok := in["max"].([]interface{}); ok {
		obj.Max = expandResourceQuantityString(v)
	}

	if v, ok := in["default"].([]interface{}); ok {
		obj.Default = expandResourceQuantityString(v)
	}

	if v, ok := in["default_request"].([]interface{}); ok {
		obj.DefaultRequest = expandResourceQuantity1170(v)
	}

	//log.Println("expandNamespaceLimitRangeConfig <<")
	if v, ok := in["ratio"].([]interface{}); ok {
		//log.Println("expandNamespaceLimitRangeConfig ")
		obj.Ratio = expandResourceRatio(v)
	}

	return obj
}

func expandResourceRatio(p []interface{}) *commonpb.ResourceRatio {
	//log.Println("expandResourceRatio ")
	obj := &commonpb.ResourceRatio{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	//w1 := spew.Sprintf("%+v", in)
	//log.Println("expandResourceRatio << in", w1)
	if v, ok := in["memory"].(float64); ok {
		obj.Memory = float32(v)
	}

	if v, ok := in["cpu"].(float64); ok {
		obj.Cpu = float32(v)
	}

	//log.Println("expandResourceRatio obj ", obj)
	return obj
}

// Flatteners

func flattenNamespace(d *schema.ResourceData, in *infrapb.Namespace) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	// XXX Debug
	//w1 := spew.Sprintf("%+v", in.Spec)
	//log.Println("flattenNamespaceSpec before ", w1)
	var ret []interface{}
	ret, err = flattenNamespaceSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	//w1 = spew.Sprintf("%+v", ret)
	//log.Println("flattenNamespaceSpec after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenPSP(in *infrapb.NamespacePSP) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	return []interface{}{obj}
}

func flattenNamespaceResourceQuotas(in *infrapb.NamespaceResourceQuotas) []interface{} {
	if in == nil {
		return nil
	}

	retNil := true
	obj := make(map[string]interface{})

	if len(in.ConfigMaps) > 0 {
		obj["config_maps"] = in.ConfigMaps
		retNil = false
	}

	if len(in.CpuLimits) > 0 {
		obj["cpu_limits"] = in.CpuLimits
		retNil = false
	}
	if len(in.CpuRequests) > 0 {
		obj["cpu_requests"] = in.CpuRequests
		retNil = false
	}
	if len(in.MemoryLimits) > 0 {
		obj["memory_limits"] = in.MemoryLimits
		retNil = false
	}
	if len(in.GpuRequests) > 0 {
		obj["gpu_requests"] = in.GpuRequests
		retNil = false
	}
	if len(in.GpuLimits) > 0 {
		obj["gpu_limits"] = in.GpuLimits
		retNil = false
	}
	if len(in.MemoryRequests) > 0 {
		obj["memory_requests"] = in.MemoryRequests
		retNil = false
	}
	if len(in.PersistentVolumeClaims) > 0 {
		obj["persistent_volume_claims"] = in.PersistentVolumeClaims
		retNil = false
	}
	if len(in.Pods) > 0 {
		obj["pods"] = in.Pods
		retNil = false
	}
	if len(in.ReplicationControllers) > 0 {
		obj["replication_controllers"] = in.ReplicationControllers
		retNil = false
	}
	if len(in.Secrets) > 0 {
		obj["secrets"] = in.Secrets
		retNil = false
	}
	if len(in.Services) > 0 {
		obj["services"] = in.Services
		retNil = false
	}
	if len(in.ServicesLoadBalancers) > 0 {
		obj["services_load_balancers"] = in.ServicesLoadBalancers
		retNil = false
	}
	if len(in.ServicesNodePorts) > 0 {
		obj["services_node_ports"] = in.ServicesNodePorts
		retNil = false
	}
	if len(in.StorageRequests) > 0 {
		obj["storage_requests"] = in.StorageRequests
		retNil = false
	}
	if len(in.EphemeralStorageLimits) > 0 {
		obj["ephemeral_storage_limits"] = in.EphemeralStorageLimits
		retNil = false
	}
	if len(in.EphemeralStorageRequests) > 0 {
		obj["ephemeral_storage_requests"] = in.EphemeralStorageRequests
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenNamespaceLimitRange(in *infrapb.NamespaceLimitRange) []interface{} {
	if in == nil {
		return nil
	}
	retNil := true
	obj := make(map[string]interface{})

	if in.Pod != nil {
		obj["pod"] = flattenNamespaceLimitRangeConfig(in.Pod)
		retNil = false
	}

	if in.Container != nil {
		obj["container"] = flattenNamespaceLimitRangeConfig(in.Container)
		retNil = false
	}

	if retNil {
		return nil
	}

	return []interface{}{obj}
}

func flattenNamespaceLimitRangeConfig(in *infrapb.NamespaceLimitRangeConfig) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})

	if in.Min != nil {
		obj["min"] = flattenResourceQuantity(in.Min)
	}

	if in.Max != nil {
		obj["max"] = flattenResourceQuantity(in.Max)
	}

	if in.Default != nil {
		obj["default"] = flattenResourceQuantity(in.Default)
	}

	if in.Default != nil {
		obj["default_request"] = flattenResourceQuantity(in.DefaultRequest)
	}

	//log.Println("flattenNamespaceLimitRangeConfig ", in.Ratio)
	if in.Ratio != nil {
		obj["ratio"] = flattenRatio(in.Ratio)
	}

	return []interface{}{obj}
}

func flattenNamespaceSpec(in *infrapb.NamespaceSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceSpec empty input")
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenNamespaceSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenNamespaceSpec MarshalJSON error", err)
	}

	log.Println("flattenNamespaceSpec jsonBytes ", string(jsonBytes))

	nsat := namespaceSpecTranspose{}
	err = json.Unmarshal(jsonBytes, &nsat)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Psp != nil {
		obj["psp"] = flattenPSP(in.Psp)
	}

	if in.Placement != nil {
		obj["placement"] = flattenPlacement(in.Placement)
	}

	if in.Drift != nil {
		obj["drift"] = flattenDrift(in.Drift)
	}

	if in.ResourceQuotas != nil {
		obj["resource_quotas"] = flattenNamespaceResourceQuotas(in.ResourceQuotas)
	}

	if in.LimitRange != nil {
		obj["limit_range"] = flattenNamespaceLimitRange(in.LimitRange)
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenNamespaceArtifact before ", w1)

	var ret []interface{}
	ret, err = flattenNamespaceArtifact(&nsat, v)
	if err != nil {
		log.Println("flattenNamespaceArtifact error ", err)
		return nil, err
	}
	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenNamespaceArtifact after ", w1)
	obj["artifact"] = ret

	v, ok = obj["network_policy_params"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	if nsat.NetworkPolicyParms != nil {
		ret, err = flattenNetworkPolicyParams(nsat.NetworkPolicyParms, v)
		if err != nil {
			log.Println("flattenNamespaceArtifact error ", err)
			return nil, err
		}
		obj["network_policy_params"] = ret
	}

	v, ok = obj["namespace_mesh_policy_params"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	if nsat.NamespaceMeshPolicyParams != nil {
		ret, err = flattenNamespaceMeshPolicyParams(nsat.NamespaceMeshPolicyParams, v)
		if err != nil {
			log.Println("flattenNamespaceArtifact error ", err)
			return nil, err
		}
		obj["namespace_mesh_policy_params"] = ret
	}

	return []interface{}{obj}, nil
}

func flattenNetworkPolicies(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "empty network policy")
	}

	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}

		out[i] = obj
	}

	return out, nil
}

func flattenNetworkPolicyParams(in *infrapb.NetworkPolicyParams, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "networkpolicyparams empty")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["network_policy_enabled"] = in.NetworkPolicyEnabled
	if len(in.Policies) > 0 {
		v, ok := obj["policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret, err := flattenNetworkPolicies(in.Policies, v)
		if err != nil {
			log.Println("flattenNamespaceArtifact error ", err)
			return nil, err
		}
		obj["policies"] = ret
	}

	return []interface{}{obj}, nil
}

func flattenNamespaceMeshPolicies(in []*commonpb.ResourceNameAndVersionRef, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, nil
	}

	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}

		out[i] = obj
	}

	return out, nil
}

func flattenNamespaceMeshPolicyParams(in *infrapb.NamespaceMeshPolicyParams, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "namespacemeshpolicyparams empty")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if obj["mesh_enabled"] == nil {
		if !in.MeshEnabled && len(in.Policies) <= 0 {
			return nil, nil
		}
	}
	obj["mesh_enabled"] = in.MeshEnabled
	if len(in.Policies) > 0 {
		v, ok := obj["policies"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		ret, err := flattenNamespaceMeshPolicies(in.Policies, v)
		if err != nil {
			log.Println("flattenNamespaceArtifact error ", err)
			return nil, err
		}
		obj["policies"] = ret
	}

	return []interface{}{obj}, nil
}

func flattenNamespaceArtifact(nsat *namespaceSpecTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true
	if len(nsat.Artifact.Repository) > 0 {
		obj["repository"] = nsat.Artifact.Repository
		retNil = false
	}

	if len(nsat.Artifact.Revision) > 0 {
		obj["revision"] = nsat.Artifact.Revision
		retNil = false
	}

	if nsat.Artifact.Path != nil {
		obj["path"] = flattenFile(nsat.Artifact.Path)
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}

func isNamespaceAlreadyExists(ctx context.Context, d *schema.ResourceData) bool {

	meta := GetMetaData(d)
	if meta == nil {
		return false
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return false
	}

	_, err = client.InfraV3().Namespace().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: meta.Project,
	})
	if err != nil {
		return false
	}
	return true
}
