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
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	commonpb "github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
)

type ContainerRegistrySpecTranspose struct {
	Provider string         `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
	Endpoint string         `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Secret   *commonpb.File `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
	// Types that are assignable to Credentials:
	//	*ContainerRegistrySpec_UserPass
	//	*ContainerRegistrySpec_Aws
	//	*ContainerRegistrySpec_Gcp
	//	*ContainerRegistrySpec_Azure
	Credentials struct {
		//	*ContainerRegistrySpec_UserPass
		Username string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
		Password string `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
		//	*ContainerRegistrySpec_Aws
		Region          string `protobuf:"bytes,1,opt,name=region,proto3" json:"region,omitempty"`
		AccessKeyID     string `protobuf:"bytes,2,opt,name=accessKeyID,proto3" json:"accessKeyID,omitempty"`
		AccessSecretKey string `protobuf:"bytes,3,opt,name=accessSecretKey,proto3" json:"accessSecretKey,omitempty"`
		//	*ContainerRegistrySpec_Gcp
		JsonKeyData string `protobuf:"bytes,1,opt,name=jsonKeyData,proto3" json:"jsonKeyData,omitempty"`
		//	*ContainerRegistrySpec_Azure
		ServicePrincipalID       string `protobuf:"bytes,1,opt,name=servicePrincipalID,proto3" json:"servicePrincipalID,omitempty"`
		ServicePrincipalPassword string `protobuf:"bytes,2,opt,name=servicePrincipalPassword,proto3" json:"servicePrincipalPassword,omitempty"`
	} `json:"credentials,omitempty"`
}

func resourceContainerRegistry() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceContainerRegistryCreate,
		ReadContext:   resourceContainerRegistryRead,
		UpdateContext: resourceContainerRegistryUpdate,
		DeleteContext: resourceContainerRegistryDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ContainerRegistrySchema.Schema,
	}
}

func resourceContainerRegistryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ContainerRegistry create starts")
	call := "create"
	diags := resourceContainerRegistryUpsert(ctx, d, m, call)
	log.Println("Upsert err creation failed:", diags)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("ContainerRegistry create got error, perform cleanup")
		cr, err := expandContainerRegistry(d, call)
		if err != nil {
			log.Printf("ContainerRegistry resourceContainerRegistryCreate error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.IntegrationsV3().ContainerRegistry().Delete(ctx, options.DeleteOptions{
			Name:    cr.Metadata.Name,
			Project: cr.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceContainerRegistryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ContainerRegistry update starts")
	call := "update"
	return resourceContainerRegistryUpsert(ctx, d, m, call)
}

func resourceContainerRegistryUpsert(ctx context.Context, d *schema.ResourceData, m interface{}, call string) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("ContainerRegistryUpsert starts")
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

	containerRegistry, err := expandContainerRegistry(d, call)
	if err != nil {
		log.Printf("container regsitry expandContainerRegistry error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	n1 := spew.Sprintf("%+v", containerRegistry)
	log.Println("ContainerRegistry bfr apply:", n1)

	err = client.IntegrationsV3().ContainerRegistry().Apply(ctx, containerRegistry, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		log.Printf("ContainerRegistry apply error")
		n1 := spew.Sprintf("%+v", containerRegistry)
		log.Println("ContainerRegistry w/err:", n1)
		return diag.FromErr(err)
	}

	d.SetId(containerRegistry.Metadata.Name)
	return diags

}

func resourceContainerRegistryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceContainerRegistryRead ")
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

	call := "read"
	containerRegistry, err := expandContainerRegistry(d, call)
	if err != nil {
		return diag.FromErr(err)
	}
	// if containerRegistry.Metadata.Project == "" {
	// 	return diag.FromErr(fmt.Errorf("project empty"))
	// }

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.IntegrationsV3().ContainerRegistry().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: containerRegistry.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenContainerRegistry(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceContainerRegistryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("resourceContainerRegistryDelete ")
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	call := "delete"
	cr, err := expandContainerRegistry(d, call)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().ContainerRegistry().Delete(ctx, options.DeleteOptions{
		Name:    cr.Metadata.Name,
		Project: cr.Metadata.Project,
	})
	if err != nil {
		log.Println("hub delete call err: ", err)
		return diag.FromErr(err)
	}

	return diags
}

func expandContainerRegistry(in *schema.ResourceData, call string) (*integrationspb.ContainerRegistry, error) {
	log.Println("expandContainerRegistry")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand container registry empty input")
	}
	obj := &integrationspb.ContainerRegistry{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		if len(v) == 0 || v[0] == nil {
			return nil, fmt.Errorf("%s", "expandRepoCredential empty ")
		} else {
			inp := v[0].(map[string]interface{})
			if vp, ok := inp["project"].(string); ok && len(vp) > 0 {
				log.Println("got project", ok, len(vp), vp)
				obj.Metadata = expandMetaData(v)
			} else if call == "update" {
				log.Println("wth: ", ok, len(vp), vp)
				return nil, fmt.Errorf("%s", "403 Forbidden: Project Input Field can not be empty")
			}

		}
		//obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandContainerRegistrySpec(v, call)
		if err != nil {
			return nil, err
		}
		log.Println("expandContainerRegistry got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "integrations.k8smgmt.io/v3"
	obj.Kind = "ContainerRegistry"
	return obj, nil
}

func expandContainerRegistrySpec(p []interface{}, call string) (*integrationspb.ContainerRegistrySpec, error) {
	log.Println("expandContainerRegistrySpec")
	obj := &integrationspb.ContainerRegistrySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandContainerRegistrySpec empty input")
	}

	in := p[0].(map[string]interface{})
	//log.Println("spec input", in)

	crt := ContainerRegistrySpecTranspose{}
	providersList := [11]string{"Custom", "JFrog", "System", "ECR", "DockerHub", "GCR", "Quay", "Nexus", "Harbor", "MCR", "ACR"}
	nilSpace := [4]string{"", " ", "null", "nill"}

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		for _, x := range providersList {
			//log.Println(v, x)
			//log.Println(v == x)
			if v == x {
				crt.Provider = v
			}
		}
		if crt.Provider == "" {
			return obj, fmt.Errorf("Invalid provider")
		}
	}
	if v, ok := in["endpoint"].(string); ok && len(v) > 0 {
		for _, x := range nilSpace {
			//log.Println(v, x)
			//log.Println(v == x)
			if v == x {
				return obj, fmt.Errorf("Empty Endpoint")
			}
		}
		crt.Endpoint = v
	} else if call == "update" {
		return obj, fmt.Errorf("Empty Endpoint")
	}

	if v, ok := in["secret"].([]interface{}); ok {
		crt.Secret = expandCommonpbFile(v)
	}

	if vp, ok := in["credentials"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			return nil, fmt.Errorf("%s", "expandRepoCredential empty ")
		} else {
			inp := vp[0].(map[string]interface{})
			//user pass
			if v, ok := inp["username"].(string); ok && len(v) > 0 {
				crt.Credentials.Username = v
			}

			if v, ok := inp["password"].(string); ok && len(v) > 0 {
				crt.Credentials.Password = v
			}
			//aws
			if v, ok := inp["region"].(string); ok && len(v) > 0 {
				crt.Credentials.Region = v
			}

			if v, ok := inp["access_key_id"].(string); ok && len(v) > 0 {
				crt.Credentials.AccessKeyID = v
			}

			if v, ok := inp["access_secret_key"].(string); ok && len(v) > 0 {
				crt.Credentials.AccessSecretKey = v
			}
			//gcp
			if v, ok := inp["json_key_data"].(string); ok && len(v) > 0 {
				crt.Credentials.JsonKeyData = v
			}
			//azure
			if v, ok := inp["service_principal_id"].(string); ok && len(v) > 0 {
				crt.Credentials.ServicePrincipalID = v
			}

			if v, ok := inp["service_principal_password"].(string); ok && len(v) > 0 {
				crt.Credentials.ServicePrincipalPassword = v
			}
		}
	}
	// XXX Debug
	s := spew.Sprintf("%+v", crt)
	log.Println("expandContainerRegistrySpec crt", s)

	jsonSpec, err := json.Marshal(crt)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("expandContainerRegistrySpec jsonSpec ", string(jsonSpec))

	err = json.Unmarshal(jsonSpec, obj)
	//err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandContainerRegistrySpec artifact Unmarshal error ", err)
		return nil, err
	}

	x := spew.Sprintf("%+v", obj)
	log.Println("expandContainerRegistrySPEC entire OBJ", x)

	return obj, nil
}

// Flatten

func flattenContainerRegistry(d *schema.ResourceData, in *integrationspb.ContainerRegistry) error {
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
	ret, err = flattenContainerRegistrySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenContainerRegistrySpec(in *integrationspb.ContainerRegistrySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenContainerRegistrySpec empty input")
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenContainerRegistrySpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenContainerRegistrySpec MarshalJSON error", err)
	}

	log.Println("flattenContainerRegistrySpec jsonBytes ", string(jsonBytes))

	crt := ContainerRegistrySpecTranspose{}
	err = json.Unmarshal(jsonBytes, &crt)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Provider) > 0 {
		obj["provider"] = in.Provider
	}

	if len(in.Endpoint) > 0 {
		obj["endpoint"] = in.Endpoint
	}

	if in.Secret != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	v, ok := obj["credentials"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenContainerRegistryCredentials before ", w1)

	var ret []interface{}
	ret, err = flattenContainerRegistryCredentials(&crt, v)
	if err != nil {
		log.Println("flattenContainerRegistryCredentials error ", err)
		return nil, err
	}
	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenContainerRegistryCredentials after ", w1)
	obj["credentials"] = ret

	return []interface{}{obj}, nil
}

func flattenContainerRegistryCredentials(cst *ContainerRegistrySpecTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true
	//user pass
	if len(cst.Credentials.Username) > 0 {
		obj["username"] = cst.Credentials.Username
		retNil = false
	}

	if len(cst.Credentials.Password) > 0 {
		obj["password"] = cst.Credentials.Password
		retNil = false
	}
	//aws
	if len(cst.Credentials.Region) > 0 {
		obj["region"] = cst.Credentials.Region
		retNil = false
	}

	if len(cst.Credentials.AccessKeyID) > 0 {
		obj["access_key_id"] = cst.Credentials.AccessKeyID
		retNil = false
	}

	if len(cst.Credentials.AccessSecretKey) > 0 {
		obj["access_secret_key"] = cst.Credentials.AccessSecretKey
		retNil = false
	}
	//GCP
	if len(cst.Credentials.JsonKeyData) > 0 {
		obj["json_key_data"] = cst.Credentials.JsonKeyData
		retNil = false
	}
	//Azure
	if len(cst.Credentials.ServicePrincipalID) > 0 {
		obj["service_principal_id"] = cst.Credentials.ServicePrincipalID
		retNil = false
	}

	if len(cst.Credentials.ServicePrincipalPassword) > 0 {
		obj["service_principal_password"] = cst.Credentials.ServicePrincipalPassword
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}
