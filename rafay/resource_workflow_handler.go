package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/structpb"
)

func resourceWorkflowHandler() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkflowHandlerCreate,
		ReadContext:   resourceWorkflowHandlerRead,
		UpdateContext: resourceWorkflowHandlerUpdate,
		DeleteContext: resourceWorkflowHandlerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWorkflowHandlerImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.WorkflowHandlerSchema.Schema,
	}
}

func resourceWorkflowHandlerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("workflow handler create")
	diags := resourceWorkflowHandlerUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandWorkflowHandler(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().WorkflowHandler().Delete(ctx, options.DeleteOptions{
			Name:    cc.Metadata.Name,
			Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceWorkflowHandlerUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("workflow handler upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().WorkflowHandler().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceWorkflowHandlerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("workflow handler read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	wh, err := client.EaasV1().WorkflowHandler().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: cc.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenWorkflowHandler(d, wh)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceWorkflowHandlerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceWorkflowHandlerUpsert(ctx, d, m)
}

func resourceWorkflowHandlerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("workflow handler delete starts")
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

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		log.Println("error while expanding workflow handler during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().WorkflowHandler().Delete(ctx, options.DeleteOptions{
		Name:    cc.Metadata.Name,
		Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandWorkflowHandler(in *schema.ResourceData) (*eaaspb.WorkflowHandler, error) {
	log.Println("expand workflow handler resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand workflow handler empty input")
	}
	obj := &eaaspb.WorkflowHandler{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandWorkflowHandlerSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "WorkflowHandler"
	return obj, nil
}

func expandWorkflowHandlerSpec(p []interface{}) (*eaaspb.WorkflowHandlerSpec, error) {
	log.Println("expand workflow handler spec")
	spec := &eaaspb.WorkflowHandlerSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand workflow handler spec empty input")
	}

	in := p[0].(map[string]interface{})

	if c, ok := in["config"].([]interface{}); ok && len(c) > 0 {
		spec.Config = expandWorkflowHandlerConfig(c)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["inputs"].([]interface{}); ok && len(v) > 0 {
		spec.Inputs = expandConfigContextCompoundRefs(v)
	}

	var err error
	if v, ok := in["outputs"].(string); ok && len(v) > 0 {
		spec.Outputs, err = expandWorkflowHandlerOutputs(v)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandWorkflowHandlerConfig(p []interface{}) *eaaspb.WorkflowHandlerConfig {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	workflowHandlerConfig := eaaspb.WorkflowHandlerConfig{}
	in := p[0].(map[string]interface{})

	if typ, ok := in["type"].(string); ok && len(typ) > 0 {
		workflowHandlerConfig.Type = typ
	}

	if ts, ok := in["timeout_seconds"].(int); ok {
		workflowHandlerConfig.TimeoutSeconds = int64(ts)
	}

	if sc, ok := in["success_condition"].(string); ok && len(sc) > 0 {
		workflowHandlerConfig.SuccessCondition = sc
	}

	if ts, ok := in["max_retry_count"].(int); ok {
		workflowHandlerConfig.MaxRetryCount = int32(ts)
	}

	if v, ok := in["container"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.Container = expandWorkflowHandlerContainerConfig(v)
	}

	if v, ok := in["http"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.Http = expandWorkflowHandlerHttpConfig(v)
	}

	if v, ok := in["polling_config"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.PollingConfig = expandPollingConfig(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		workflowHandlerConfig.TimeoutSeconds = int64(h)
	}

	return &workflowHandlerConfig
}

func expandWorkflowHandlerContainerConfig(p []interface{}) *eaaspb.ContainerDriverConfig {
	cc := eaaspb.ContainerDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &cc
	}

	in := p[0].(map[string]interface{})

	if img, ok := in["image"].(string); ok && len(img) > 0 {
		cc.Image = img
	}

	if args, ok := in["arguments"].([]interface{}); ok && len(args) > 0 {
		cc.Arguments = toArrayString(args)
	}

	if cmds, ok := in["commands"].([]interface{}); ok && len(cmds) > 0 {
		cc.Commands = toArrayString(cmds)
	}

	if clm, ok := in["cpu_limit_milli"].(string); ok && len(clm) > 0 {
		cc.CpuLimitMilli = clm
	}

	if ev, ok := in["env_vars"].(map[string]interface{}); ok && len(ev) > 0 {
		cc.EnvVars = toMapString(ev)
	}

	if f, ok := in["files"].(map[string]interface{}); ok && len(f) > 0 {
		cc.Files = toMapByte(f)
	}

	if v, ok := in["image_pull_credentials"].([]interface{}); ok && len(v) > 0 {
		cc.ImagePullCredentials = expandImagePullCredentials(v)
	}

	if v, ok := in["kube_config_options"].([]interface{}); ok && len(v) > 0 {
		cc.KubeConfigOptions = expandKubeConfigOptions(v)
	}

	if v, ok := in["kube_options"].([]interface{}); ok && len(v) > 0 {
		cc.KubeOptions = expandContainerKubeOptions(v)
	}

	if mlb, ok := in["memory_limit_mb"].(string); ok && len(mlb) > 0 {
		cc.MemoryLimitMb = mlb
	}

	if v, ok := in["volume_options"].([]interface{}); ok && len(v) > 0 {
		volumes := expandContainerWorkflowHandlerVolumeOptions(v)
		if len(volumes) > 0 {
			cc.VolumeOptions = volumes[0]
		}
	}

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		cc.Volumes = expandContainerWorkflowHandlerVolumeOptions(v)
	}

	if wdp, ok := in["working_dir_path"].(string); ok && len(wdp) > 0 {
		cc.WorkingDirPath = wdp
	}

	return &cc
}

func expandImagePullCredentials(p []interface{}) *eaaspb.ContainerImagePullCredentials {
	hc := eaaspb.ContainerImagePullCredentials{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if pass, ok := in["password"].(string); ok && len(pass) > 0 {
		hc.Password = pass
	}

	if registry, ok := in["registry"].(string); ok && len(registry) > 0 {
		hc.Registry = registry
	}

	if username, ok := in["username"].(string); ok && len(username) > 0 {
		hc.Username = username
	}

	return &hc
}

func expandKubeConfigOptions(p []interface{}) *eaaspb.ContainerKubeConfigOptions {
	hc := eaaspb.ContainerKubeConfigOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if kc, ok := in["kube_config"].(string); ok && len(kc) > 0 {
		hc.KubeConfig = kc
	}

	if ofc, ok := in["out_of_cluster"].(bool); ok {
		hc.OutOfCluster = ofc
	}

	return &hc
}

func expandContainerKubeOptions(p []interface{}) *eaaspb.ContainerKubeOptions {
	hc := eaaspb.ContainerKubeOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if lbls, ok := in["labels"].(map[string]interface{}); ok && len(lbls) > 0 {
		hc.Labels = toMapString(lbls)
	}

	if ns, ok := in["namespace"].(string); ok && len(ns) > 0 {
		hc.Namespace = ns
	}

	if ns, ok := in["node_selector"].(map[string]interface{}); ok && len(ns) > 0 {
		hc.NodeSelector = toMapString(ns)
	}

	if r, ok := in["resources"].([]interface{}); ok && len(r) > 0 {
		hc.Resources = toArrayString(r)
	}

	if sc, ok := in["security_context"].([]interface{}); ok && len(sc) > 0 {
		hc.SecurityContext = expandSecurityContext(sc)
	}

	if san, ok := in["service_account_name"].(string); ok && len(san) > 0 {
		hc.ServiceAccountName = san
	}

	if tolerations, ok := in["tolerations"].([]interface{}); ok {
		hc.Tolerations = expandV3Tolerations(tolerations)
	}

	return &hc
}

func expandSecurityContext(p []interface{}) *eaaspb.KubeSecurityContext {
	ksc := eaaspb.KubeSecurityContext{}
	if len(p) == 0 || p[0] == nil {
		return &ksc
	}

	in := p[0].(map[string]interface{})

	if privileged, ok := in["privileged"].([]interface{}); ok && len(privileged) > 0 {
		ksc.Privileged = expandBoolValue(privileged)
	}

	if ro, ok := in["read_only_root_file_system"].([]interface{}); ok && len(ro) > 0 {
		ksc.ReadOnlyRootFileSystem = expandBoolValue(ro)
	}

	return &ksc
}

func expandWorkflowHandlerHttpConfig(p []interface{}) *eaaspb.HTTPDriverConfig {
	hc := eaaspb.HTTPDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if body, ok := in["body"].(string); ok && len(body) > 0 {
		hc.Body = body
	}

	if endpoint, ok := in["endpoint"].(string); ok && len(endpoint) > 0 {
		hc.Endpoint = endpoint
	}

	if headers, ok := in["headers"].(map[string]interface{}); ok && len(headers) > 0 {
		hc.Headers = toMapString(headers)
	}

	if method, ok := in["method"].(string); ok && len(method) > 0 {
		hc.Method = method
	}

	return &hc
}

func expandWorkflowHandlerOutputs(p string) (*structpb.Struct, error) {
	if len(p) == 0 {
		return nil, nil
	}

	var s structpb.Struct
	if err := s.UnmarshalJSON([]byte(p)); err != nil {
		return nil, err
	}
	return &s, nil
}

// Flatteners

func flattenWorkflowHandler(d *schema.ResourceData, in *eaaspb.WorkflowHandler) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenV3MetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenWorkflowHandlerSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten workflow handler spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenWorkflowHandlerSpec(in *eaaspb.WorkflowHandlerSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten workflow handler spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Config != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["config"] = flattenWorkflowHandlerConfig(in.Config, v)
	}
	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["inputs"] = flattenConfigContextCompoundRefs(in.Inputs)
	obj["outputs"] = flattenWorkflowHandlerOutputs(in.Outputs)
	return []interface{}{obj}, nil
}

func flattenWorkflowHandlerConfig(input *eaaspb.WorkflowHandlerConfig, p []interface{}) []interface{} {
	log.Println("flatten workflow handler config start", input)
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(input.Type) > 0 {
		obj["type"] = input.Type
	}

	obj["timeout_seconds"] = input.TimeoutSeconds

	if len(input.SuccessCondition) > 0 {
		obj["success_condition"] = input.SuccessCondition
	}

	obj["max_retry_count"] = input.MaxRetryCount

	if input.Container != nil {
		v, ok := obj["container"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["container"] = flattenWorkflowHandlerContainerConfig(input.Container, v)
	}

	if input.Http != nil {
		v, ok := obj["http"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["http"] = flattenWorkflowHandlerHttpConfig(input.Http, v)
	}

	if input.PollingConfig != nil {
		v, ok := obj["polling_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["polling_config"] = flattenPollingConfig(input.PollingConfig, v)
	}

	return []interface{}{obj}
}

func flattenWorkflowHandlerContainerConfig(in *eaaspb.ContainerDriverConfig, p []interface{}) []interface{} {
	log.Println("flatten container workflow handler config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["arguments"] = toArrayInterface(in.Arguments)
	obj["commands"] = toArrayInterface(in.Commands)

	if len(in.CpuLimitMilli) > 0 {
		obj["cpu_limit_milli"] = in.CpuLimitMilli
	}

	obj["env_vars"] = toMapInterface(in.EnvVars)
	obj["files"] = toMapByteInterface(in.Files)

	if len(in.Image) > 0 {
		obj["image"] = in.Image
	}

	if in.ImagePullCredentials != nil {
		v, ok := obj["image_pull_credentials"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["image_pull_credentials"] = flattenImagePullCredentials(in.ImagePullCredentials, v)
	}

	if in.KubeConfigOptions != nil {
		v, ok := obj["kube_config_options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["kube_config_options"] = flattenContainerKubeConfig(in.KubeConfigOptions, v)
	}

	if in.KubeOptions != nil {
		v, ok := obj["kube_options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["kube_options"] = flattenContainerKubeOptions(in.KubeOptions, v)
	}

	if len(in.MemoryLimitMb) > 0 {
		obj["memory_limit_mb"] = in.MemoryLimitMb
	}

	if in.VolumeOptions != nil {
		v, ok := obj["volume_options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volume_options"] = flattenContainerWorkflowHandlerVolumeOptions([]*eaaspb.ContainerDriverVolumeOptions{
			in.VolumeOptions,
		}, v)
	}

	if len(in.Volumes) > 0 {
		v, ok := obj["volumes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volumes"] = flattenContainerWorkflowHandlerVolumeOptions(in.Volumes, v)
	}

	if len(in.WorkingDirPath) > 0 {
		obj["working_dir_path"] = in.WorkingDirPath
	}

	return []interface{}{obj}
}

func flattenImagePullCredentials(in *eaaspb.ContainerImagePullCredentials, p []interface{}) []interface{} {
	log.Println("flatten container image pull credentials start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Registry) > 0 {
		obj["registry"] = in.Registry
	}

	if len(in.Username) > 0 {
		obj["username"] = in.Username
	}

	if len(in.Password) > 0 {
		obj["password"] = in.Password
	}

	return []interface{}{obj}
}

func flattenContainerKubeConfig(in *eaaspb.ContainerKubeConfigOptions, p []interface{}) []interface{} {
	log.Println("flatten container kube config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.KubeConfig) > 0 {
		obj["kube_config"] = in.KubeConfig
	}

	obj["out_of_cluster"] = in.OutOfCluster

	return []interface{}{obj}
}

func flattenContainerKubeOptions(in *eaaspb.ContainerKubeOptions, p []interface{}) []interface{} {
	log.Println("flatten container kube options start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["labels"] = toMapInterface(in.Labels)

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	obj["node_selector"] = toMapInterface(in.NodeSelector)
	obj["resources"] = toArrayInterface(in.Resources)

	if in.SecurityContext != nil {
		v, ok := obj["security_context"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["security_context"] = flattenSecurityContext(in.SecurityContext, v)
	}

	if len(in.ServiceAccountName) > 0 {
		obj["service_account_name"] = in.ServiceAccountName
	}

	if len(in.Tolerations) > 0 {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tolerations"] = flattenV3Tolerations(in.Tolerations, v)
	} else {
		delete(obj, "tolerations")
	}

	return []interface{}{obj}
}

func flattenSecurityContext(in *eaaspb.KubeSecurityContext, p []interface{}) []interface{} {
	log.Println("flatten kube security context start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["privileged"] = flattenBoolValue(in.Privileged)
	obj["read_only_root_file_system"] = flattenBoolValue(in.ReadOnlyRootFileSystem)

	return []interface{}{obj}
}

func flattenWorkflowHandlerHttpConfig(in *eaaspb.HTTPDriverConfig, p []interface{}) []interface{} {
	log.Println("flatten http config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Body) > 0 {
		obj["body"] = in.Body
	}

	if len(in.Endpoint) > 0 {
		obj["endpoint"] = in.Endpoint
	}

	obj["headers"] = toMapInterface(in.Headers)

	if len(in.Method) > 0 {
		obj["method"] = in.Method
	}

	return []interface{}{obj}
}

func expandContainerWorkflowHandlerVolumeOptions(p []interface{}) []*eaaspb.ContainerDriverVolumeOptions {
	volumes := make([]*eaaspb.ContainerDriverVolumeOptions, 0)
	if len(p) == 0 {
		return volumes
	}

	for indx := range p {
		volume := &eaaspb.ContainerDriverVolumeOptions{}
		if p[indx] == nil {
			return volumes
		}
		in := p[indx].(map[string]interface{})

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			volume.MountPath = mp
		}

		if pvcsz, ok := in["pvc_size_gb"].(string); ok && len(pvcsz) > 0 {
			volume.PvcSizeGB = pvcsz
		}

		if pvcsc, ok := in["pvc_storage_class"].(string); ok && len(pvcsc) > 0 {
			volume.PvcStorageClass = pvcsc
		}

		if usepvc, ok := in["use_pvc"].([]interface{}); ok && len(usepvc) > 0 {
			volume.UsePVC = expandBoolValue(usepvc)
		}

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
		}

		volumes = append(volumes, volume)

	}

	return volumes
}

func flattenContainerWorkflowHandlerVolumeOptions(input []*eaaspb.ContainerDriverVolumeOptions, p []interface{}) []interface{} {
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten container workflow handler volume options", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["use_pvc"] = flattenBoolValue(in.UsePVC)

		if len(in.MountPath) > 0 {
			obj["mount_path"] = in.MountPath
		}

		if len(in.PvcSizeGB) > 0 {
			obj["pvc_size_gb"] = in.PvcSizeGB
		}

		if len(in.PvcStorageClass) > 0 {
			obj["pvc_storage_class"] = in.PvcStorageClass
		}

		obj["enable_backup_and_restore"] = in.EnableBackupAndRestore

		out[i] = &obj
	}

	return out
}

func flattenWorkflowHandlerOutputs(in *structpb.Struct) string {
	if in == nil {
		return ""
	}
	b, _ := in.MarshalJSON()
	return string(b)
}

func resourceWorkflowHandlerImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	log.Printf("WorkflowHandler Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceWorkflowHandlerImport idParts:", idParts)

	log.Println("resourceWorkflowHandlerImport Invoking expandWorkflowHandler")
	cc, err := expandWorkflowHandler(d)
	if err != nil {
		log.Printf("resourceWorkflowHandlerImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	cc.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(cc.Metadata.Name)
	return []*schema.ResourceData{d}, nil
}
