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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type credentialsSpecTranspose struct {
	Sharing  *commonpb.SharingSpec `yaml:"sharing,omitempty"`
	Type     string                `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Provider string                `protobuf:"bytes,1,opt,name=provider,proto3" json:"provider,omitempty"`
	// Types that are assignable to Credentials:
	//	*CredentialsSpec_AwsRolebased
	//	*CredentialsSpec_AwsAccessbased
	//	*CredentialsSpec_Gcp
	//	*CredentialsSpec_Azure
	//	*CredentialsSpec_Vsphere
	//	*CredentialsSpec_Minio
	//	*CredentialsSpec_SshRemote
	Credentials CredentialsDetailsTranspose `protobuf:"bytes,1,opt,name=credentials,proto3" json:"credentials,omitempty"`
}

type CredentialsDetailsTranspose struct {
	Type           string       `protobuf:"bytes,1,opt,name=type,proto3" json:"type,omitempty"`
	Arn            string       `protobuf:"bytes,1,opt,name=arn,proto3" json:"arn,omitempty"`
	AccessId       string       `protobuf:"bytes,1,opt,name=accessId,proto3" json:"accessId,omitempty"`
	SecretKey      string       `protobuf:"bytes,1,opt,name=secretKey,proto3" json:"secretKey,omitempty"`
	SessionToken   string       `protobuf:"bytes,1,opt,name=sessionToken,proto3" json:"sessionToken,omitempty"`
	File           string       `protobuf:"bytes,1,opt,name=file,proto3" json:"file,omitempty"`
	TenantId       string       `protobuf:"bytes,1,opt,name=tenantId,proto3" json:"tenantId,omitempty"`
	SubscriptionId string       `protobuf:"bytes,1,opt,name=subscriptionId,proto3" json:"subscriptionId,omitempty"`
	ClientId       string       `protobuf:"bytes,1,opt,name=clientId,proto3" json:"clientId,omitempty"`
	ClientSecret   string       `protobuf:"bytes,1,opt,name=clientSecret,proto3" json:"clientSecret,omitempty"`
	GatewayId      string       `protobuf:"bytes,1,opt,name=gatewayId,proto3" json:"gatewayId,omitempty"`
	VsphereServer  string       `protobuf:"bytes,1,opt,name=vsphereServer,proto3" json:"vsphereServer,omitempty"`
	Username       string       `protobuf:"bytes,1,opt,name=username,proto3" json:"username,omitempty"`
	Password       string       `protobuf:"bytes,1,opt,name=password,proto3" json:"password,omitempty"`
	ExternalId     string       `protobuf:"bytes,1,opt,name=externalId,proto3" json:"externalId,omitempty"`
	AccountId      string       `protobuf:"bytes,1,opt,name=accountId,proto3" json:"accountId,omitempty"`
	Agents         []*AgentMeta `protobuf:"bytes,1,rep,name=agents,proto3" json:"agents,omitempty"`
	PrivateKey     string       `protobuf:"bytes,2,opt,name=privateKey,proto3" json:"privateKey,omitempty"`
	Port           string       `protobuf:"bytes,4,opt,name=port,proto3" json:"port,omitempty"`
	Passphrase     string       `protobuf:"bytes,5,opt,name=passphrase,proto3" json:"passphrase,omitempty"`
}

type AgentMeta struct {
	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Id   string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
}

func resourceCredentials() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCredentialsCreate,
		ReadContext:   resourceCredentialsRead,
		UpdateContext: resourceCredentialsUpdate,
		DeleteContext: resourceCredentialsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceCredentialsImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.CredentialsSchema.Schema,
	}
}

func resourceCredentialsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceCredentials idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceCredentials d.Id:", d.Id())
	log.Println("resourceCredentials d_debug", d_debug)

	credentials, err := expandCredentials(d)
	if err != nil {
		log.Printf("resourceCredentials expand error")
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	credentials.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(credentials.Metadata))
	if err != nil {
		log.Println("import set err")
		return nil, err
	}
	d.SetId(credentials.Metadata.Name)
	return []*schema.ResourceData{d}, nil
}

func resourceCredentialsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceCredentialsCreate create starts")
	diags := resourceV3CredentialsUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Credentials create got error, perform cleanup")
		ss, err := expandCredentials(d)
		if err != nil {
			log.Printf("Credentials expandCredentials error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.InfraV3().Credentials().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceCredentialsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Credentials update starts")
	return resourceV3CredentialsUpsert(ctx, d, m)
}

func resourceV3CredentialsUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Credentials upsert starts")
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

	credentials, err := expandCredentials(d)
	if err != nil {
		log.Printf("Credentials expandCredentials error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Credentials().Apply(ctx, credentials, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", credentials)
		log.Println("Credentials apply credentials:", n1)
		log.Printf("Credentials apply error")
		return diag.FromErr(err)
	}

	d.SetId(credentials.Metadata.Name)
	return diags

}

func resourceCredentialsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceCredentialsRead")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	tfCredentialsState, err := expandCredentials(d)
	if err != nil {
		log.Println("expandCredentials failed in Read; clearing ID")
		d.SetId("") // avoid hard failure
		return nil
	}

	stateName := d.Id()
	currentName := tfCredentialsState.Metadata.Name
	log.Printf("Read(): state name = %s, current config name = %s", stateName, currentName)

	// If name in config != ID from state, prefer ID for safe deletion
	if stateName != "" && currentName != "" && stateName != currentName {
		log.Printf("Mismatch between config and state during Read. Using state ID %s", stateName)
		tfCredentialsState.Metadata.Name = stateName
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.InfraV3().Credentials().Get(ctx, options.GetOptions{
		Name:    tfCredentialsState.Metadata.Name,
		Project: tfCredentialsState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			log.Printf("resource not found in Read: %s", tfCredentialsState.Metadata.Name)
			d.SetId("") // remove from state gracefully
			return nil
		}
		return diag.FromErr(err)
	}

	if tfCredentialsState.Spec != nil && tfCredentialsState.Spec.Sharing != nil && !tfCredentialsState.Spec.Sharing.Enabled && ag.Spec.Sharing == nil {
		  ag.Spec.Sharing = &commonpb.SharingSpec{}
		  ag.Spec.Sharing.Enabled = false
		  ag.Spec.Sharing.Projects = tfCredentialsState.Spec.Sharing.Projects

	}

	err = flattenCredentials(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}


func resourceCredentialsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("resourceCredentialsDelete")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cred, err := expandCredentials(d)
	if err != nil {
		log.Println("delete expand err")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.InfraV3().Credentials().Delete(ctx, options.DeleteOptions{
		Name:    cred.Metadata.Name,
		Project: cred.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandCredentials(in *schema.ResourceData) (*infrapb.Credentials, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand credentials empty input")
	}
	obj := &infrapb.Credentials{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandCredentialsSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandCredentialsSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Credentials"
	return obj, nil
}

func expandCredentialsSpec(p []interface{}) (*infrapb.CredentialsSpec, error) {
	obj := &infrapb.CredentialsSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandCredentialsSpec empty input")
	}

	in := p[0].(map[string]interface{})

	cst := credentialsSpecTranspose{}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		cst.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		cst.Type = v
	}

	if v, ok := in["provider"].(string); ok && len(v) > 0 {
		cst.Provider = v
	}

	if vp, ok := in["credentials"].([]interface{}); ok && len(vp) > 0 {
		if len(vp) == 0 || vp[0] == nil {
			log.Println("expandCredentials empty credentials")
		} else {
			ina := vp[0].(map[string]interface{})

			if v, ok := ina["type"].(string); ok && len(v) > 0 {
				cst.Credentials.Type = v
			}

			if v, ok := ina["arn"].(string); ok && len(v) > 0 {
				cst.Credentials.Arn = v
			}

			if v, ok := ina["access_id"].(string); ok && len(v) > 0 {
				cst.Credentials.AccessId = v
			}

			if v, ok := ina["secret_key"].(string); ok && len(v) > 0 {
				cst.Credentials.SecretKey = v
			}

			if v, ok := ina["session_token"].(string); ok && len(v) > 0 {
				cst.Credentials.SessionToken = v
			}

			if v, ok := ina["file"].(string); ok && len(v) > 0 {
				cst.Credentials.File = v
			}

			if v, ok := ina["tenant_id"].(string); ok && len(v) > 0 {
				cst.Credentials.TenantId = v
			}

			if v, ok := ina["subscription_id"].(string); ok && len(v) > 0 {
				cst.Credentials.SubscriptionId = v
			}

			if v, ok := ina["client_id"].(string); ok && len(v) > 0 {
				cst.Credentials.ClientId = v
			}

			if v, ok := ina["client_secret"].(string); ok && len(v) > 0 {
				cst.Credentials.ClientSecret = v
			}

			if v, ok := ina["gateway_id"].(string); ok && len(v) > 0 {
				cst.Credentials.GatewayId = v
			}

			if v, ok := ina["vsphere_server"].(string); ok && len(v) > 0 {
				cst.Credentials.VsphereServer = v
			}

			if v, ok := ina["username"].(string); ok && len(v) > 0 {
				cst.Credentials.Username = v
			}

			if v, ok := ina["password"].(string); ok && len(v) > 0 {
				cst.Credentials.Password = v
			}

			if v, ok := ina["private_key"].(string); ok && len(v) > 0 {
				cst.Credentials.PrivateKey = v
			}

			if v, ok := ina["port"].(string); ok && len(v) > 0 {
				cst.Credentials.Port = v
			}

			if v, ok := ina["passphrase"].(string); ok && len(v) > 0 {
				cst.Credentials.Passphrase = v
			}

			if v, ok := ina["agents"].([]interface{}); ok && len(v) > 0 {
				for _, a := range v {
					am := a.(map[string]interface{})
					agentMeta := AgentMeta{}
					if v, ok := am["name"].(string); ok && len(v) > 0 {
						agentMeta.Name = v
					}
					if v, ok := am["id"].(string); ok && len(v) > 0 {
						agentMeta.Id = v
					}
					cst.Credentials.Agents = append(cst.Credentials.Agents, &agentMeta)
				}
			}

			if v, ok := ina["account_id"].(string); ok && len(v) > 0 {
				cst.Credentials.AccountId = v
			}

			if v, ok := ina["external_id"].(string); ok && len(v) > 0 {
				cst.Credentials.ExternalId = v
			}
		}
	}

	// XXX Debug
	s := spew.Sprintf("%+v", cst)
	log.Println("expandCredentialsSpec cst", s)

	jsonSpec, err := json.Marshal(cst)
	if err != nil {
		return nil, err
	}

	// XXX Debug
	log.Println("expandCredentialsSpec jsonSpec ", string(jsonSpec))

	err = obj.UnmarshalJSON(jsonSpec)
	if err != nil {
		log.Println("expandCredentialsSpec artifact UnmarshalJSON error ", err)
		return nil, err
	}

	return obj, nil
}

// flatten

func flattenCredentials(d *schema.ResourceData, in *infrapb.Credentials) error {
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
	ret, err = flattenCredentialsSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenCredentialsSpec(in *infrapb.CredentialsSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenCredentialsSpec empty input")
	}

	jsonBytes, err := in.MarshalJSON()
	if err != nil {
		log.Println("flattenCredentialsSpec MarshalJSON error", err)
		return nil, fmt.Errorf("%s %+v", "flattenCredentialsSpec MarshalJSON error", err)
	}

	log.Println("flattenCredentialsSpec jsonBytes ", string(jsonBytes))

	cst := credentialsSpecTranspose{}
	err = json.Unmarshal(jsonBytes, &cst)

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	} else {
		obj["sharing"] = []interface{}{}
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if len(in.Provider) > 0 {
		obj["provider"] = in.Provider
	}

	v, ok := obj["credentials"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	w1 := spew.Sprintf("%+v", v)
	log.Println("flattenCredentialsConfig before ", w1)

	var ret []interface{}
	ret, err = flattenCredentialsConfig(&cst, v)
	if err != nil {
		log.Println("flattenCredentialsConfig error ", err)
		return nil, err
	}
	// XXX Debug
	w1 = spew.Sprintf("%+v", ret)
	log.Println("flattenCredentialsConfig after ", w1)
	obj["credentials"] = ret

	return []interface{}{obj}, nil
}

func flattenCredentialsConfig(cst *credentialsSpecTranspose, p []interface{}) ([]interface{}, error) {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	retNil := true
	if len(cst.Credentials.Type) > 0 {
		obj["type"] = cst.Credentials.Type
		retNil = false
	}

	if len(cst.Credentials.Arn) > 0 {
		obj["arn"] = cst.Credentials.Arn
		retNil = false
	}

	if len(cst.Credentials.AccessId) > 0 {
		obj["access_id"] = cst.Credentials.AccessId
		retNil = false
	}

	if len(cst.Credentials.SecretKey) > 0 {
		obj["secret_key"] = cst.Credentials.SecretKey
		retNil = false
	}

	if len(cst.Credentials.SessionToken) > 0 {
		obj["session_token"] = cst.Credentials.SessionToken
		retNil = false
	}

	if len(cst.Credentials.File) > 0 {
		obj["file"] = cst.Credentials.File
		retNil = false
	}

	if len(cst.Credentials.TenantId) > 0 {
		obj["tenant_id"] = cst.Credentials.TenantId
		retNil = false
	}

	if len(cst.Credentials.SubscriptionId) > 0 {
		obj["subscription_id"] = cst.Credentials.SubscriptionId
		retNil = false
	}

	if len(cst.Credentials.ClientId) > 0 {
		obj["client_id"] = cst.Credentials.ClientId
		retNil = false
	}

	if len(cst.Credentials.ClientSecret) > 0 {
		obj["client_secret"] = cst.Credentials.ClientSecret
		retNil = false
	}

	if len(cst.Credentials.GatewayId) > 0 {
		obj["gateway_id"] = cst.Credentials.GatewayId
		retNil = false
	}

	if len(cst.Credentials.VsphereServer) > 0 {
		obj["vsphere_server"] = cst.Credentials.VsphereServer
		retNil = false
	}

	if len(cst.Credentials.Username) > 0 {
		obj["username"] = cst.Credentials.Username
		retNil = false
	}

	if len(cst.Credentials.PrivateKey) > 0 {
		obj["private_key"] = cst.Credentials.PrivateKey
		retNil = false
	}

	if len(cst.Credentials.Port) > 0 {
		obj["port"] = cst.Credentials.Port
		retNil = false
	}

	if len(cst.Credentials.Passphrase) > 0 {
		obj["passphrase"] = cst.Credentials.Passphrase
		retNil = false
	}

	if AgentMetaLen := len(cst.Credentials.Agents); AgentMetaLen > 0 {
		agents := make([]interface{}, AgentMetaLen)
		for i, v := range cst.Credentials.Agents {
			agent := map[string]interface{}{}
			if len(v.Name) > 0 {
				agent["name"] = v.Name
			}
			if len(v.Id) > 0 {
				agent["id"] = v.Id
			}
			agents[i] = agent
		}
		obj["agents"] = agents
		retNil = false
	}

	if len(cst.Credentials.Password) > 0 {
		obj["password"] = cst.Credentials.Password
		retNil = false
	}

	if len(cst.Credentials.ExternalId) > 0 {
		obj["external_id"] = cst.Credentials.ExternalId
		retNil = false
	}

	if len(cst.Credentials.AccountId) > 0 {
		obj["account_id"] = cst.Credentials.AccountId
		retNil = false
	}

	if retNil {
		return nil, nil
	}

	return []interface{}{obj}, nil

}
