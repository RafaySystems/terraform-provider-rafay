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
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceClusterNetworkPolicyRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceClusterNetworkPolicyRuleCreate,
		ReadContext:   resourceClusterNetworkPolicyRuleRead,
		UpdateContext: resourceClusterNetworkPolicyRuleUpdate,
		DeleteContext: resourceClusterNetworkPolicyRuleDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ClusterNetworkPolicyRuleSchema.Schema,
	}
}

func resourceClusterNetworkPolicyRuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("ClusterNetworkPolicyRule create starts")
	diags := resourceClusterNetworkPolicyRuleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("clusterNetworkPolicyRule create got error, perform cleanup")
		cnpr, err := expandClusterNetworkPolicyRule(d)
		if err != nil {
			log.Printf("clusterNetworkPolicyRule expandClusterNetworkPolicyRule error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SecurityV3().ClusterNetworkPolicyRule().Delete(ctx, options.DeleteOptions{
			Name:    cnpr.Metadata.Name,
			Project: cnpr.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceClusterNetworkPolicyRuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("clusterNetworkPolicyRule update starts")
	return resourceClusterNetworkPolicyRuleUpsert(ctx, d, m)
}

func resourceClusterNetworkPolicyRuleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("clusterNetworkPolicyRule upsert starts")
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

	clusterNetworkPolicyRule, err := expandClusterNetworkPolicyRule(d)
	if err != nil {
		log.Printf("clusterNetworkPolicyRule expandClusterNetworkPolicyRule error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().ClusterNetworkPolicyRule().Apply(ctx, clusterNetworkPolicyRule, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", clusterNetworkPolicyRule)
		log.Println("clusterNetworkPolicyRule apply clusterNetworkPolicyRule:", n1)
		log.Printf("clusterNetworkPolicyRule apply error")
		return diag.FromErr(err)
	}

	d.SetId(clusterNetworkPolicyRule.Metadata.Name)
	return diags

}

func resourceClusterNetworkPolicyRuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceClusterNetworkPolicyRuleRead ")
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

	tfClusterNetworkPolicyRuleState, err := expandClusterNetworkPolicyRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	cnp, err := client.SecurityV3().ClusterNetworkPolicyRule().Get(ctx, options.GetOptions{
		//Name:    tfClusterNetworkPolicyRuleState.Metadata.Name,
		Name:    meta.Name,
		Project: tfClusterNetworkPolicyRuleState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenClusterNetworkPolicyRule(d, cnp)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceClusterNetworkPolicyRuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cnpr, err := expandClusterNetworkPolicyRule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, getUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SecurityV3().ClusterNetworkPolicyRule().Delete(ctx, options.DeleteOptions{
		Name:    cnpr.Metadata.Name,
		Project: cnpr.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandClusterNetworkPolicyRule(in *schema.ResourceData) (*securitypb.ClusterNetworkPolicyRule, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand clusterNetworkPolicyRule empty input")
	}
	obj := &securitypb.ClusterNetworkPolicyRule{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandClusterNetworkPolicyRuleSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandClusterNetworkPolicyRuleSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "security.k8smgmt.io/v3"
	obj.Kind = "ClusterNetworkPolicyRule"
	return obj, nil
}

func expandClusterNetworkPolicyRuleSpec(p []interface{}) (*securitypb.ClusterNetworkPolicyRuleSpec, error) {
	obj := &securitypb.ClusterNetworkPolicyRuleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandClusterNetworkPolicySpec empty input")
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

func flattenClusterNetworkPolicyRule(d *schema.ResourceData, in *securitypb.ClusterNetworkPolicyRule) error {
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
	ret, err = flattenClusterNetworkPolicyRuleSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenClusterNetworkPolicyRuleSpec(in *securitypb.ClusterNetworkPolicyRuleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenClusterNetworkPolicyRuleSpec empty input")
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
