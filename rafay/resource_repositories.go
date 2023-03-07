package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/integrationspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/user"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type repositorySpec struct {
	Type        string                            `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Endpoint    string                            `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	Agents      []*integrationspb.AgentMeta       `protobuf:"bytes,3,rep,name=agents,proto3" json:"agents,omitempty"`
	Options     *integrationspb.RepositoryOptions `protobuf:"bytes,4,opt,name=options,proto3" json:"options,omitempty"`
	Secret      *commonpb.File                    `protobuf:"bytes,5,opt,name=secret,proto3" json:"secret,omitempty"`
	Credentials struct {
		Username   string `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
		Password   string `protobuf:"bytes,2,opt,name=password,proto3" json:"password,omitempty"`
		PrivateKey string `protobuf:"bytes,1,opt,name=privateKey,proto3" json:"privateKey,omitempty"`
	} `json:"credentials,omitempty"`
	Sharing *commonpb.SharingSpec `protobuf:"bytes,5,opt,name=sharing,proto3" json:"sharing,omitempty"`
}

func resourceRepositories() *schema.Resource {
	modSchema := resource.RepositorySchema.Schema
	modSchema["impersonate"] = &schema.Schema{
		Description: "impersonate user",
		Optional:    true,
		Type:        schema.TypeString,
	}
	return &schema.Resource{
		CreateContext: resourceRepositoriesCreate,
		ReadContext:   resourceRepositoriesRead,
		UpdateContext: resourceRepositoriesUpdate,
		DeleteContext: resourceRepositoriesDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

func resourceRepositoriesCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("repository create starts")
	return resourceRepositoryUpsert(ctx, d, m)
}

func resourceRepositoryUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("repository upsert starts")
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

	repo, err := expandRepository(d)
	if err != nil {
		log.Printf("repository expandRepository error")
		return diag.FromErr(err)
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
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().Repository().Apply(ctx, repo, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", repo)
		log.Println("repository apply repository:", n1)
		log.Printf("repository apply error")
		return diag.FromErr(err)
	}

	d.SetId(repo.Metadata.Name)
	return diags
}

func resourceRepositoriesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("repository read starts")
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

	repoTFState, err := expandRepository(d)
	if err != nil {
		log.Printf("repository expandRepository error")
		return diag.FromErr(err)
	}

	// XXX Debug
	w1 := spew.Sprintf("%+v", repoTFState)
	log.Println("resourceRepositoriesRead repoTFState", w1)

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
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	repo, err := client.IntegrationsV3().Repository().Get(ctx, options.GetOptions{
		//Name:    repoTFState.Metadata.Name,
		Name:    meta.Name,
		Project: repoTFState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", repo)
	// log.Println("resourceRepositoriesRead wl", w1)

	err = flattenRepository(d, repo)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceRepositoriesUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("repository update starts")
	return resourceRepositoryUpsert(ctx, d, m)
}

func resourceRepositoriesDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("repository upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	repo, err := expandRepository(d)
	if err != nil {
		log.Printf("repository expandRepository error")
		return diag.FromErr(err)
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
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.IntegrationsV3().Repository().Delete(ctx, options.DeleteOptions{
		Name:    repo.Metadata.Name,
		Project: repo.Metadata.Project,
	})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", repo)
		log.Println("repository delete repository:", n1)
		log.Printf("repository delete error")
		return diag.FromErr(err)
	}
	return diags
}

// Expanders

func expandRepository(in *schema.ResourceData) (*integrationspb.Repository, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand blueprint empty input")
	}
	obj := &integrationspb.Repository{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandRepositorySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandRepositorySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "integrations.k8smgmt.io/v3"
	obj.Kind = "Repository"
	return obj, nil
}

func expandRepositorySpec(p []interface{}) (*integrationspb.RepositorySpec, error) {
	repoSpec := repositorySpec{}
	obj := integrationspb.RepositorySpec{}

	if len(p) == 0 || p[0] == nil {
		return &obj, fmt.Errorf("%s", "expandAddonSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		//obj.Type = v
		repoSpec.Type = v
	}

	if v, ok := in["endpoint"].(string); ok && len(v) > 0 {
		//obj.Endpoint = v
		repoSpec.Endpoint = v
	}

	if v, ok := in["agents"].([]interface{}); ok && len(v) > 0 {
		//obj.Agents = expandAgents(v)
		repoSpec.Agents = expandAgents(v)
	}

	if v, ok := in["options"].([]interface{}); ok && len(v) > 0 {
		//obj.Options = expandRepositoryOptions(v)
		repoSpec.Options = expandRepositoryOptions(v)
	} else {
		repoSpec.Options = expandRepositoryOptions(nil)
	}

	if v, ok := in["secret"].([]interface{}); ok {
		//obj.Secret = expandCommonpbFile(v)
		repoSpec.Secret = expandCommonpbFile(v)
	}

	if vp, ok := in["credentials"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			return nil, fmt.Errorf("%s", "expandRepoCredential empty ")
		}
		inp := vp[0].(map[string]interface{})
		if v, ok := inp["username"].(string); ok && len(v) > 0 {
			repoSpec.Credentials.Username = v
		}

		if v, ok := inp["password"].(string); ok && len(v) > 0 {
			repoSpec.Credentials.Password = v
		}

		if v, ok := inp["private_key"].(string); ok && len(v) > 0 {
			repoSpec.Credentials.PrivateKey = v
		}
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		repoSpec.Sharing = expandSharingSpec(v)
	}

	// XXX Debug
	s := spew.Sprintf("%+v", repoSpec)
	log.Println("expandRepositorySpec repoSpec", s)

	jsonSpec, err := json.Marshal(repoSpec)
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

	// XXX Debug
	s1 := spew.Sprintf("%+v", obj)
	log.Println("expandRepositorySpec obj", s1)

	return &obj, nil
}

func expandRepositoryOptions(p []interface{}) *integrationspb.RepositoryOptions {
	obj := &integrationspb.RepositoryOptions{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["insecure"].(bool); ok {
		obj.Insecure = v
	} else {
		obj.Insecure = false
	}

	if v, ok := in["enable_submodules"].(bool); ok {
		obj.EnableSubmodules = v
	} else {
		obj.EnableSubmodules = false
	}

	if v, ok := in["max_retires"].(int); ok {
		obj.MaxRetires = int32(v)
	} else {
		obj.MaxRetires = 0
	}

	if v, ok := in["enable_lfs"].(bool); ok {
		obj.EnableLFS = v
	} else {
		obj.EnableLFS = false
	}

	if v, ok := in["ca_cert"].([]interface{}); ok {
		obj.CaCert = expandCommonpbFile(v)
	}

	return obj
}

// Flatteners

func flattenAgents(input []*integrationspb.AgentMeta, p []interface{}) []interface{} {
	log.Println("flattenAgents")
	if input == nil {
		return nil
	}

	out := make([]interface{}, len(input))
	for i, in := range input {
		log.Println("flattenAgents in ", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		// if len(in.Name) > 0 {
		// 	obj["id"] = in.Id
		// }

		out[i] = &obj
	}

	return out
}

func flattenCredentials(in *repositorySpec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	if len(p) == 0 || p[0] == nil {
		return nil
	}

	obj := p[0].(map[string]interface{})

	if len(in.Credentials.Username) > 0 {
		obj["username"] = in.Credentials.Username
	}

	if len(in.Credentials.Password) > 0 {
		obj["password"] = in.Credentials.Password
	}

	if len(in.Credentials.PrivateKey) > 0 {
		obj["private_key"] = in.Credentials.PrivateKey
	}

	return []interface{}{obj}
}

func flattenRepoOptions(in *integrationspb.RepositoryOptions, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	retNel := true

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Insecure {
		obj["insecure"] = in.Insecure
		retNel = false
	} else {
		if v1, ok := obj["insecure"].(bool); ok && v1 {
			obj["insecure"] = false
			retNel = false
		}
	}

	if in.EnableSubmodules {
		obj["enable_submodules"] = in.EnableSubmodules
		retNel = false
	} else {
		if v1, ok := obj["enable_submodules"].(bool); ok && v1 {
			obj["enable_submodules"] = false
			retNel = false
		}
	}

	if in.MaxRetires > 0 {
		obj["max_retires"] = int(in.MaxRetires)
		retNel = false
	} else {
		if v1, ok := obj["max_retires"].(int); ok && v1 != int(in.MaxRetires) {
			obj["max_retires"] = 0
			retNel = false
		}
	}

	if in.EnableLFS {
		obj["enable_lfs"] = in.EnableLFS
		retNel = false
	} else {
		if v1, ok := obj["enable_lfs"].(bool); ok && v1 {
			obj["enable_lfs"] = false
			retNel = false
		}
	}

	// if in.CaCert != nil {
	// 	obj["ca_cert"] = flattenCommonpbFile(in.CaCert)
	// 	retNel = false
	// }

	if retNel {
		return nil
	}

	return []interface{}{obj}
}

// flattenRepositorySpec RepositorySpec to TF State
func flattenRepositorySpec(in *integrationspb.RepositorySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		log.Println("flattenRepositorySpec empty input")
		return nil, fmt.Errorf("%s", "flattenRepositorySpec empty input")
	}

	// XXX Debug
	in1 := spew.Sprintf("%+v", in)
	log.Println("flattenRepositorySpec in", in1)

	// XXX Debug
	ob := spew.Sprintf("%+v", p)
	log.Println("flattenRepositorySpec p", ob)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	} else {
		obj["type"] = ""
	}

	if len(in.Endpoint) > 0 {
		obj["endpoint"] = in.Endpoint
	} else {
		obj["endpoint"] = ""
	}

	if len(in.Agents) > 0 {
		v, ok := obj["agents"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["agents"] = flattenAgents(in.Agents, v)
	} else {
		obj["agents"] = nil
	}

	if in.Options != nil {
		v, ok := obj["options"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["options"] = flattenRepoOptions(in.Options, v)
	} else {
		obj["options"] = nil
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	if in.Secret != nil {
		obj["secrets"] = flattenCommonpbFile(in.Secret)
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenRepositorySpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenRepositorySpec MarshalJSON error", err)
	}

	rs := repositorySpec{}
	err = json.Unmarshal(jsonBytes, &rs)
	if err != nil {
		return nil, fmt.Errorf("%s %+v", "flattenRepositorySpec json unmarshal error", err)
	}

	// XXX Debug
	log.Println("flattenRepositorySpec jsonBytes:", string(jsonBytes))
	s1 := spew.Sprintf("%+v", rs)
	log.Println("flattenRepositorySpec rs", s1)

	v, ok := obj["credentials"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	obj["credentials"] = flattenCredentials(&rs, v)

	// XXX Debug
	o1 := spew.Sprintf("%+v", obj)
	log.Println("flattenRepositorySpec obj", o1)

	return []interface{}{obj}, nil
}

func flattenRepository(d *schema.ResourceData, in *integrationspb.Repository) error {
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
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenRepository before ", w1)
	var ret []interface{}
	ret, err = flattenRepositorySpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenRepository after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}
