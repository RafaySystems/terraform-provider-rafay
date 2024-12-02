package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/systempb"
	"github.com/RafaySystems/rctl/pkg/versioninfo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var showZTKAArtifactFlag bool = false

func resourceZTKARule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZTKARuleCreate,
		ReadContext:   resourceZTKARuleRead,
		UpdateContext: resourceZTKARuleUpdate,
		DeleteContext: resourceZTKARuleDelete,
		Importer: &schema.ResourceImporter{
			State: resourceZTKARuleImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.ZTKARuleSchema.Schema,
	}
}

func resourceZTKARuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() == "" {
		return nil, fmt.Errorf("ztkarule name not provided, usage e.g terraform import rafay_ztkarule.resource <ztkarule-name>")
	}
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceZTKARuleImport idParts:", idParts)

	if len(idParts) > 1 && idParts[1] == "show_artifact" {
		showZTKAArtifactFlag = true
	}

	ztkaRule, err := expandZTKARule(d)
	if err != nil {
		log.Printf("ZTKARule expandZTKARule error")
		return nil, err
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	ztkaRule.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(ztkaRule.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(ztkaRule.Metadata.Name)

	return []*schema.ResourceData{d}, nil
}

func resourceZTKARuleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("ztka rule create")
	diags := resourceZTKARuleUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		zr, err := expandZTKARule(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().ZTKARule().Delete(ctx, options.DeleteOptions{
			Name: zr.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceZTKARuleUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("ztka rule upsert starts")
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

	ac, err := expandZTKARule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ZTKARule().Apply(ctx, ac, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ac.Metadata.Name)
	return diags
}

func resourceZTKARuleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resource ztka rule ")

	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	ac, err := client.SystemV3().ZTKARule().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenZTKARule(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags
}

func resourceZTKARuleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceZTKARuleUpsert(ctx, d, m)
}

func resourceZTKARuleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	addon, err := expandZTKARule(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().ZTKARule().Delete(ctx, options.DeleteOptions{
		Name: addon.Metadata.Name,
	})

	if err != nil {
		log.Println("ztka rule delete error")
		return diag.FromErr(err)
	}

	return diags
}

func expandZTKARule(in *schema.ResourceData) (*systempb.ZTKARule, error) {
	log.Println("expand ztka rule")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand ZTKARule empty input")
	}
	obj := &systempb.ZTKARule{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandZTKARuleSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "ZTKARule"

	return obj, nil
}

func expandClusterSelectors(p []interface{}) *systempb.ZTKAClusters {
	obj := &systempb.ZTKAClusters{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	if v, ok := in["select_all"].(bool); ok {
		obj.SelectAll = v
	}

	if v, ok := in["match_names"].([]interface{}); ok && len(v) > 0 {
		obj.MatchNames = toArrayString(v)
	}

	if v, ok := in["match_labels"].(map[string]interface{}); ok && len(v) > 0 {
		obj.MatchLabels = toMapString(v)
	}

	return obj
}

func expandProjectSelector(p []interface{}) *systempb.ZTKAProjects {
	obj := &systempb.ZTKAProjects{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	if v, ok := in["select_all"].(bool); ok {
		obj.SelectAll = v
	}

	if v, ok := in["match_names"].([]interface{}); ok && len(v) > 0 {
		obj.MatchNames = toArrayStringSorted(v)
	}

	return obj
}

func expandZTKARuleSpec(p []interface{}) (*systempb.ZTKARuleSpec, error) {

	obj := &systempb.ZTKARuleSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandZTKARuleSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cluster_selector"].([]interface{}); ok && len(v) > 0 {
		obj.ClusterSelector = expandClusterSelectors(v)
	}

	if v, ok := in["project_selector"].([]interface{}); ok && len(v) > 0 {
		obj.ProjectSelector = expandProjectSelector(v)
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

func flattenZTKARule(d *schema.ResourceData, in *systempb.ZTKARule) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenZTKARuleSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten ztka rule spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenZTKARuleSpec(in *systempb.ZTKARuleSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenZTKARuleSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}
	obj["published"] = in.Published

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	var flattenArtifact []interface{}
	var err error
	flattenArtifact, err = FlattenArtifactSpec(showZTKAArtifactFlag, in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}
	obj["artifact"] = flattenArtifact

	if in.ClusterSelector != nil {
		v, ok := obj["cluster_selector"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cluster_selector"] = flattenClusterSelector(in.ClusterSelector, v)
	}

	if in.ProjectSelector != nil {
		v, ok := obj["project_selector"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["project_selector"] = flattenProjectSelector(in.ProjectSelector, v)
	}

	return []interface{}{obj}, nil
}

func flattenProjectSelector(in *systempb.ZTKAProjects, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in.SelectAll {
		obj["select_all"] = in.SelectAll
	}

	if in.MatchNames != nil && len(in.MatchNames) > 0 {
		obj["match_names"] = toArrayInterface(in.MatchNames)
	}
	return []interface{}{obj}
}

func flattenClusterSelector(in *systempb.ZTKAClusters, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	if in.SelectAll {
		obj["select_all"] = in.SelectAll
	}

	if in.MatchNames != nil && len(in.MatchNames) > 0 {
		obj["match_names"] = toArrayInterface(in.MatchNames)
	}
	if in.MatchLabels != nil && len(in.MatchLabels) > 0 {
		obj["match_labels"] = toMapInterface(in.MatchLabels)
	}
	return []interface{}{obj}
}
