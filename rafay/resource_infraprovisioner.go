package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/infraprovisioner"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type infraProvisionerSpec struct {
	Type       string         `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Repository string         `protobuf:"bytes,2,opt,name=repository,proto3" json:"repository,omitempty"`
	Revision   string         `protobuf:"bytes,3,opt,name=revision,proto3" json:"revision,omitempty"`
	FolderPath *commonpb.File `protobuf:"bytes,4,opt,name=folderPath,proto3" json:"folderPath,omitempty"`
	Secret     *commonpb.File `protobuf:"bytes,5,opt,name=secret,proto3" json:"secret,omitempty"`
	Config     struct {
		Version         string               `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
		InputVars       []*gitopspb.KeyValue `protobuf:"bytes,2,rep,name=inputVars,proto3" json:"inputVars,omitempty"`
		TfVarsFilePath  *commonpb.File       `protobuf:"bytes,3,opt,name=tfVarsFilePath,proto3" json:"tfVarsFilePath,omitempty"`
		EnvVars         []*gitopspb.KeyValue `protobuf:"bytes,4,rep,name=envVars,proto3" json:"envVars,omitempty"`
		BackendVars     []*gitopspb.KeyValue `protobuf:"bytes,6,rep,name=backendVars,proto3" json:"backendVars,omitempty"`
		BackendFilePath *commonpb.File       `protobuf:"bytes,7,opt,name=backendFilePath,proto3" json:"backendFilePath,omitempty"`
	} `json:"config,omitempty"`
}

func resourceInfraProvisioner() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInfraProvisionerCreate,
		ReadContext:   resourceInfraProvisionerRead,
		UpdateContext: resourceInfraProvisionerUpdate,
		DeleteContext: resourceInfraProvisionerDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.InfraProvisionerSchema.Schema,
	}
}

func resourceInfraProvisionerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("infra provisioner create starts")
	diags := resourceInfraProvisionerUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("infra provisioner create got error, perform cleanup")
		ag, err := expandInfraProvisioner(d)
		if err != nil {
			log.Printf("infra provisioner expandInfraProvisioner error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.GitopsV3().InfraProvisioner().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceInfraProvisionerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("infra provisioner update starts")
	return resourceInfraProvisionerUpsert(ctx, d, m)
}

func resourceInfraProvisionerUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("infra provisioner upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ip, err := expandInfraProvisioner(d)
	if err != nil {
		log.Printf("ip expandInfraProvisioner error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().InfraProvisioner().Apply(ctx, ip, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", ip)
		log.Println("ip apply:", n1)
		log.Printf("ip apply error")
		return diag.FromErr(err)
	}

	d.SetId(ip.Metadata.Name)
	return diags

}

func resourceInfraProvisionerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceInfraProvisionerRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfIPState, err := expandInfraProvisioner(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ip, err := client.GitopsV3().InfraProvisioner().Get(ctx, options.GetOptions{
		Name:    tfIPState.Metadata.Name,
		Project: tfIPState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenInfraProvisioner(d, ip)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceInfraProvisionerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ip, err := expandInfraProvisioner(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().InfraProvisioner().Delete(ctx, options.DeleteOptions{
		Name:    ip.Metadata.Name,
		Project: ip.Metadata.Project,
	})

	if err != nil {
		return resourceInfraProvisionerV2Delete(ctx, ip)
	}

	return diags
}

func resourceInfraProvisionerV2Delete(ctx context.Context, ip *gitopspb.InfraProvisioner) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(ip.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	//delete agent
	err = infraprovisioner.DeleteInfraProvisioner(ip.Metadata.Name, projectId)
	if err != nil {
		log.Println("error deleting agent")
	} else {
		log.Println("Deleted InfraProvisioner: ", ip.Metadata.Name)
	}
	return diags
}

func expandInfraProvisioner(in *schema.ResourceData) (*gitopspb.InfraProvisioner, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand InfraProvisioner empty input")
	}
	obj := &gitopspb.InfraProvisioner{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandInfraProvisionerSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandInfraProvisionerSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "InfraProvisioner"
	return obj, nil
}

func expandInfraProvisionerSpec(p []interface{}) (*gitopspb.InfraProvisionerSpec, error) {
	ipSpec := infraProvisionerSpec{}
	obj := gitopspb.InfraProvisionerSpec{}
	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandInfraProvisionerSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		//obj.Type = v
		ipSpec.Type = v
	}

	if v, ok := in["repository"].(string); ok && len(v) > 0 {
		//obj.Repository = v
		ipSpec.Repository = v
	}

	if v, ok := in["revision"].(string); ok && len(v) > 0 {
		ipSpec.Revision = v

	}

	if v, ok := in["folderPath"].([]interface{}); ok {
		ipSpec.FolderPath = expandCommonpbFile(v)
	}

	if v, ok := in["secret"].([]interface{}); ok {
		ipSpec.Secret = expandCommonpbFile(v)
	}

	if vp, ok := in["config"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			return nil, fmt.Errorf("%s", "expandRepoCredential empty ")
		}
		inp := vp[0].(map[string]interface{})
		if v, ok := inp["version"].(string); ok && len(v) > 0 {
			ipSpec.Config.Version = v
		}

		if v, ok := inp["input_vars"].([]interface{}); ok && len(v) > 0 {
			ipSpec.Config.InputVars = expandInfraProvisionerConfigKeyValue(v)
		}

		if v, ok := inp["tf_vars_file_path"].([]interface{}); ok {
			ipSpec.Config.TfVarsFilePath = expandCommonpbFile(v)
		}

		if v, ok := inp["env_vars"].([]interface{}); ok && len(v) > 0 {
			ipSpec.Config.EnvVars = expandInfraProvisionerConfigKeyValue(v)
		}

		if v, ok := inp["backend_vars"].([]interface{}); ok && len(v) > 0 {
			ipSpec.Config.BackendVars = expandInfraProvisionerConfigKeyValue(v)
		}

		if v, ok := in["backend_file_path"].([]interface{}); ok {
			ipSpec.Config.BackendFilePath = expandCommonpbFile(v)
		}
	}

	jsonSpec, err := json.Marshal(ipSpec)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("expandRepositorySpec jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandRepositorySpec UnmarshalJSON error ", err)
		return nil, err
	}

	return &obj, nil
}

func expandInfraProvisionerConfigKeyValue(p []interface{}) []*gitopspb.KeyValue {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	out := make([]*gitopspb.KeyValue, len(p))

	for i := range p {
		obj := gitopspb.KeyValue{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}

		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}

		if v, ok := in["type"].(string); ok && len(v) > 0 {
			obj.Type = v
		}
		out[i] = &obj
	}
	return out
}

// Flatten

func flattenInfraProvisioner(d *schema.ResourceData, in *gitopspb.InfraProvisioner) error {
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

	var ret []interface{}
	ret, err = flattenInfraProvisionerSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenInfraProvisionerSpec(in *gitopspb.InfraProvisionerSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenWorkloadSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if len(in.Repository) > 0 {
		obj["repository"] = in.Repository
	}

	if len(in.Revision) > 0 {
		obj["revision"] = in.Revision
	}

	if in.FolderPath != nil {
		obj["folderPath"] = flattenCommonpbFile(in.FolderPath)
	}

	if in.Secret != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenInfraProvisionerSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenInfraProvisionerSpec MarshalJSON error", err)
	}

	ip := infraProvisionerSpec{}
	err = json.Unmarshal(jsonBytes, &ip)
	if err != nil {
		return nil, fmt.Errorf("%s %+v", "flattenInfraProvisionerSpec json unmarshal error", err)
	}

	// XXX Debug
	log.Println("flattenInfraProvisionerSpec jsonBytes:", string(jsonBytes))
	s1 := spew.Sprintf("%+v", ip)
	log.Println("flattenInfraProvisionerSpec ip", s1)

	v, ok := obj["config"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["config"] = flattenIPConfig(&ip, v)

	// XXX Debug
	o1 := spew.Sprintf("%+v", obj)
	log.Println("flattenRepositorySpec obj", o1)

	return []interface{}{obj}, nil
}

func flattenIPConfig(in *infraProvisionerSpec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	if len(p) == 0 || p[0] == nil {
		return nil
	}

	obj := p[0].(map[string]interface{})

	if len(in.Config.Version) > 0 {
		obj["version"] = in.Config.Version
	}

	if in.Config.TfVarsFilePath != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	if len(in.Config.InputVars) > 0 {
		v, ok := obj["input_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["input_vars"] = flattenIPKeyValue(in.Config.InputVars, v)
	} else {
		obj["input_vars"] = nil
	}

	if in.Config.BackendFilePath != nil {
		obj["backend_file_path"] = flattenCommonpbFile(in.Secret)
	}

	if len(in.Config.EnvVars) > 0 {
		v, ok := obj["env_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["env_vars"] = flattenIPKeyValue(in.Config.EnvVars, v)
	} else {
		obj["env_vars"] = nil
	}

	if len(in.Config.BackendVars) > 0 {
		v, ok := obj["backend_vars"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["backend_vars"] = flattenIPKeyValue(in.Config.BackendVars, v)
	} else {
		obj["backend_vars"] = nil
	}

	return []interface{}{obj}
}

func flattenIPKeyValue(input []*gitopspb.KeyValue, p []interface{}) []interface{} {
	log.Println("flattenStageSpecConfigActionKeyValue")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenStageSpecConfigActionKeyValue in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}

		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}

		if len(in.Type) > 0 {
			obj["Type"] = in.Type
		}

		out[i] = &obj
	}

	return out
}
