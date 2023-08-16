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
	"github.com/RafaySystems/rafay-common/proto/types/hub/securitypb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceNamespaceNetworkPolicyRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNamespaceNetworkPolicyRuleCreate,
		ReadContext:   resourceNamespaceNetworkPolicyRuleRead,
		UpdateContext: resourceNamespaceNetworkPolicyRuleUpdate,
		DeleteContext: resourceNamespaceNetworkPolicyRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.NamespaceNetworkPolicyRuleSchema.Schema,
	}
}

func resourceNamespaceNetworkPolicyRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("NamespaceNetworkPolicyRule create starts")
	diags := resourceNamespaceNetworkPolicyRuleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("namespaceNetworkPolicyRule create got error, perform cleanup")
		nnpr, err := expandNamespaceNetworkPolicyRule(d)
		if err != nil {
			log.Printf("namespaceNetworkPolicyRule expandNamespaceNetworkPolicyRule error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SecurityV3().NamespaceNetworkPolicyRule().Delete(ctx, options.DeleteOptions{
			Name:    nnpr.Metadata.Name,
			Project: nnpr.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceNamespaceNetworkPolicyRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("namespaceNetworkPolicyRule update starts")
	return resourceNamespaceNetworkPolicyRuleUpsert(ctx, d, m)
}

func resourceNamespaceNetworkPolicyRuleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("namespaceNetworkPolicyRule upsert starts")
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

	namespaceNetworkPolicyRule, err := expandNamespaceNetworkPolicyRule(d)
	if err != nil {
		log.Printf("namespaceNetworkPolicyRule expandNamespaceNetworkPolicyRule error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NamespaceNetworkPolicyRule().Apply(ctx, namespaceNetworkPolicyRule, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", namespaceNetworkPolicyRule)
		log.Println("namespaceNetworkPolicyRule apply namespaceNetworkPolicyRule:", n1)
		log.Printf("namespaceNetworkPolicyRule apply error")
		return diag.FromErr(err)
	}

	d.SetId(namespaceNetworkPolicyRule.Metadata.Name)
	return diags

}

func resourceNamespaceNetworkPolicyRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceNamespaceNetworkPolicyRuleRead ")
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

	tfNamespaceNetworkPolicyRuleState, err := expandNamespaceNetworkPolicyRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	nnpr, err := client.SecurityV3().NamespaceNetworkPolicyRule().Get(ctx, options.GetOptions{
		//Name:    tfNamespaceNetworkPolicyRuleState.Metadata.Name,
		Name:    meta.Name,
		Project: tfNamespaceNetworkPolicyRuleState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenNamespaceNetworkPolicyRule(d, nnpr)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceNamespaceNetworkPolicyRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	nnpr, err := expandNamespaceNetworkPolicyRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().NamespaceNetworkPolicyRule().Delete(ctx, options.DeleteOptions{
		Name:    nnpr.Metadata.Name,
		Project: nnpr.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandNamespaceNetworkPolicyRule(in *schema.ResourceData) (*securitypb.NamespaceNetworkPolicyRule, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand namespaceNetworkPolicyRule empty input")
	}
	obj := &securitypb.NamespaceNetworkPolicyRule{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandNamespaceNetworkPolicyRuleSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandNamespaceNetworkPolicyRuleSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "security.k8smgmt.io/v3"
	obj.Kind = "NamespaceNetworkPolicyRule"
	return obj, nil
}

func expandNamespaceNetworkPolicyRuleSpec(p []interface{}) (*securitypb.NamespaceNetworkPolicyRuleSpec, error) {
	obj := &securitypb.NamespaceNetworkPolicyRuleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandNamespaceNetworkPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["artifact"].([]interface{}); ok {
		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Artifact = objArtifact
	}

	return obj, nil
}

// Flatteners

func flattenNamespaceNetworkPolicyRule(d *schema.ResourceData, in *securitypb.NamespaceNetworkPolicyRule) error {
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
	ret, err = flattenNamespaceNetworkPolicyRuleSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenNamespaceNetworkPolicyRuleSpec(in *securitypb.NamespaceNetworkPolicyRuleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenNamespaceNetworkPolicyRuleSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	var err error
	ret, err = FlattenArtifactSpec(in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}

	obj["artifact"] = ret

	return []interface{}{obj}, nil
}
