package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDriver() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDriverCreate,
		ReadContext:   resourceDriverRead,
		UpdateContext: resourceDriverUpdate,
		DeleteContext: resourceDriverDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDriverImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.DriverSchema.Schema,
	}
}

func resourceDriverCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("driver create")
	diags := resourceDriverUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandDriver(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().Driver().Delete(ctx, options.DeleteOptions{
			Name:    cc.Metadata.Name,
			Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceDriverUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("driver upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandDriver(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().Driver().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceDriverRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("driver read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	cc, err := expandDriver(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	driver, err := client.EaasV1().Driver().Get(ctx, options.GetOptions{
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

	err = flattenDriver(d, driver)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceDriverUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceDriverUpsert(ctx, d, m)
}

func resourceDriverDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("driver delete starts")
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

	cc, err := expandDriver(d)
	if err != nil {
		log.Println("error while expanding driver during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().Driver().Delete(ctx, options.DeleteOptions{
		Name:    cc.Metadata.Name,
		Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandDriver(in *schema.ResourceData) (*eaaspb.Driver, error) {
	log.Println("expand driver resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand driver empty input")
	}
	obj := &eaaspb.Driver{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandDriverSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "Driver"
	return obj, nil
}

func expandDriverSpec(p []interface{}) (*eaaspb.DriverSpec, error) {
	log.Println("expand driver spec")
	spec := &eaaspb.DriverSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand driver spec empty input")
	}

	in := p[0].(map[string]interface{})

	if c, ok := in["config"].([]interface{}); ok && len(c) > 0 {
		spec.Config = expandDriverConfig(c)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	return spec, nil
}

func expandDriverConfig(p []interface{}) *eaaspb.DriverConfig {
	config := eaaspb.DriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &config
	}

	in := p[0].(map[string]interface{})

	if typ, ok := in["type"].(string); ok && len(typ) > 0 {
		config.Type = typ
	}

	if ts, ok := in["timeout_seconds"].(int); ok {
		config.TimeoutSeconds = int64(ts)
	}

	if sc, ok := in["success_condition"].(string); ok && len(sc) > 0 {
		config.SuccessCondition = sc
	}

	if ts, ok := in["max_retry_count"].(int); ok {
		config.MaxRetryCount = int32(ts)
	}

	if v, ok := in["container"].([]interface{}); ok && len(v) > 0 {
		config.Container = expandDriverContainerConfig(v)
	}

	if v, ok := in["http"].([]interface{}); ok && len(v) > 0 {
		config.Http = expandDriverHttpConfig(v)
	}

	return &config
}

func expandDriverContainerConfig(p []interface{}) *eaaspb.ContainerDriverConfig {
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
		volumes := expandContainerDriverVolumeOptions(v)
		if len(volumes) > 0 {
			cc.VolumeOptions = volumes[0]
		}
	}

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		cc.Volumes = expandContainerDriverVolumeOptions(v)
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

func expandDriverHttpConfig(p []interface{}) *eaaspb.HTTPDriverConfig {
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

func expandDriverCompoundRef(p []interface{}) *eaaspb.DriverCompoundRef {
	driver := &eaaspb.DriverCompoundRef{}
	if len(p) == 0 || p[0] == nil {
		return driver
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		driver.Name = v
	}

	if v, ok := in["data"].([]interface{}); ok && len(v) > 0 {
		driver.Data = expandDriverInline(v)
	}

	return driver
}

func expandDriverInline(p []interface{}) *eaaspb.DriverInline {
	driver := &eaaspb.DriverInline{}
	if len(p) == 0 || p[0] == nil {
		return driver
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["config"].([]interface{}); ok && len(v) > 0 {
		driver.Config = expandDriverConfig(v)
	}

	if v, ok := in["contexts"].([]interface{}); ok && len(v) > 0 {
		driver.Contexts = expandConfigContextCompoundRefs(v)
	}

	return driver
}

// Flatteners

func flattenDriver(d *schema.ResourceData, in *eaaspb.Driver) error {
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
	ret, err = flattenDriverSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten driver spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenDriverSpec(in *eaaspb.DriverSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten driver spec empty input")
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

		obj["config"] = flattenDriverConfig(in.Config, v)
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}

func flattenDriverConfig(input *eaaspb.DriverConfig, p []interface{}) []interface{} {
	log.Println("flatten driver config start", input)
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

		obj["container"] = flattenDriverContainerConfig(input.Container, v)
	}

	if input.Http != nil {
		v, ok := obj["http"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["http"] = flattenDriverHttpConfig(input.Http, v)
	}

	return []interface{}{obj}
}

func flattenDriverContainerConfig(in *eaaspb.ContainerDriverConfig, p []interface{}) []interface{} {
	log.Println("flatten container driver config start")
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

		obj["volume_options"] = flattenContainerDriverVolumeOptions([]*eaaspb.ContainerDriverVolumeOptions{
			in.VolumeOptions,
		}, v)
	}

	if len(in.Volumes) > 0 {
		v, ok := obj["volumes"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["volumes"] = flattenContainerDriverVolumeOptions(in.Volumes, v)
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

func flattenDriverHttpConfig(in *eaaspb.HTTPDriverConfig, p []interface{}) []interface{} {
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

func expandContainerDriverVolumeOptions(p []interface{}) []*eaaspb.ContainerDriverVolumeOptions {
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

		volumes = append(volumes, volume)

	}

	return volumes
}

func flattenContainerDriverVolumeOptions(input []*eaaspb.ContainerDriverVolumeOptions, p []interface{}) []interface{} {
	if len(input) == 0 {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flatten container driver volume options", in)
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

		out[i] = &obj
	}

	return out
}

func resourceDriverImport(d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
	log.Printf("Driver Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceDriverImport idParts:", idParts)

	log.Println("resourceDriverImport Invoking expandDriver")
	cc, err := expandDriver(d)
	if err != nil {
		log.Printf("resourceDriverImport  expand error %s", err.Error())
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

func flattenDriverCompoundRef(input *eaaspb.DriverCompoundRef) []interface{} {
	log.Println("flatten driver compound ref start")
	if input == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if len(input.Name) > 0 {
		obj["name"] = input.Name
	}
	if input.Data != nil {
		obj["data"] = flattenDriverInline(input.Data)
	}
	return []interface{}{obj}
}

func flattenDriverInline(input *eaaspb.DriverInline) []interface{} {
	log.Println("flatten driver inline start")
	if input == nil {
		return nil
	}
	obj := map[string]interface{}{}
	if input.Config != nil {
		obj["config"] = flattenDriverConfig(input.Config, []interface{}{})
	}
	if len(input.Contexts) > 0 {
		obj["contexts"] = flattenConfigContextCompoundRefs(input.Contexts)
	}
	return []interface{}{obj}
}
