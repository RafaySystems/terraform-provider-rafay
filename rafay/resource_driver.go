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

	if v, ok := in["inputs"].([]interface{}); ok && len(v) > 0 {
		spec.Inputs = expandConfigContextCompoundRefs(v)
	}

	var err error
	if v, ok := in["outputs"].(string); ok && len(v) > 0 {
		spec.Outputs, err = expandDriverOutputs(v)
		if err != nil {
			return nil, err
		}
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

func expandDriverOutputs(p string) (*structpb.Struct, error) {
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
	obj["inputs"] = flattenConfigContextCompoundRefs(in.Inputs)
	obj["outputs"] = flattenDriverOutputs(in.Outputs)
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

		obj["container"] = flattenWorkflowHandlerContainerConfig(input.Container, v)
	}

	if input.Http != nil {
		v, ok := obj["http"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["http"] = flattenWorkflowHandlerHttpConfig(input.Http, v)
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

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
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

		obj["enable_backup_and_restore"] = in.EnableBackupAndRestore

		out[i] = &obj
	}

	return out
}

func flattenDriverOutputs(in *structpb.Struct) string {
	if in == nil {
		return ""
	}
	b, _ := in.MarshalJSON()
	return string(b)
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
