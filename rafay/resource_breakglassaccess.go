package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
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

func resourceBreakGlassAccess() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBreakGlassAccessCreate,
		ReadContext:   resourceBreakGlassAccessRead,
		UpdateContext: resourceBreakGlassAccessUpdate,
		DeleteContext: resourceBreakGlassAccessDelete,
		Importer: &schema.ResourceImporter{
			State: resourceBreakGlassAccessImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.BreakGlassAccessSchema.Schema,
	}
}

func resourceBreakGlassAccessImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if d.Id() == "" {
		return nil, fmt.Errorf("username not provided, usage e.g terraform import rafay_breakglassaccess.resource <breakglassaccess-username>")
	}

	username := d.Id()

	log.Println("Importing break glass access for user: ", username)

	breakGlassAccess, err := expandBreakGlassAccess(d)
	if err != nil {
		log.Printf("breakGlassAccess expandBreakGlassAccess error")
		return nil, err
	}

	var metaD commonpb.Metadata
	metaD.Name = username
	breakGlassAccess.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(breakGlassAccess.Metadata))
	if err != nil {
		return nil, err
	}

	d.SetId(username)

	return []*schema.ResourceData{d}, nil
}

func resourceBreakGlassAccessCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("break glass access create")
	var alreadyExists bool
	alreadyExists = breakGlassAccessExists(ctx, d)

	diags := resourceBreakGlassAccessUpsert(ctx, d, m)
	if diags.HasError() && !alreadyExists {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		zr, err := expandBreakGlassAccess(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.SystemV3().BreakGlassAccess().Delete(ctx, options.DeleteOptions{
			Name: zr.Metadata.Name,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func breakGlassAccessExists(ctx context.Context, d *schema.ResourceData) bool {
	bga, err := expandBreakGlassAccess(d)
	if err != nil {
		log.Printf("breakGlassAccessExists: breakglassaccess expandBreakGlassAccess error")
		return false
	}
	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return false
	}

	bgaFromDb, err := client.SystemV3().BreakGlassAccess().Get(ctx, options.GetOptions{
		Name: bga.Metadata.Name,
	})
	if err != nil {
		return false
	}
	//Since we return empty object for each username/name, we need to check if username has group attached
	return bgaFromDb != nil && bgaFromDb.Spec != nil && bgaFromDb.Spec.Groups != nil && len(bgaFromDb.Spec.Groups) > 0
}

func resourceBreakGlassAccessUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("break glass access upsert starts")
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

	tus, err := expandBreakGlassAccess(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().BreakGlassAccess().Apply(ctx, tus, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tus.Metadata.Name)
	return diags
}

func resourceBreakGlassAccessRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resource break glass access ")

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

	ac, err := client.SystemV3().BreakGlassAccess().Get(ctx, options.GetOptions{
		Name: meta.Name,
	})
	if err != nil {
		log.Println("read get err")
		return diag.FromErr(err)
	}

	err = flattenBreakGlassAccess(d, ac)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags
}

func resourceBreakGlassAccessUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceBreakGlassAccessUpsert(ctx, d, m)
}

func resourceBreakGlassAccessDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	tus, err := expandBreakGlassAccess(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.SystemV3().BreakGlassAccess().Delete(ctx, options.DeleteOptions{
		Name: tus.Metadata.Name,
	})

	if err != nil {
		log.Println("break glass access delete error")
		return diag.FromErr(err)
	}

	return diags
}

func expandBreakGlassAccess(in *schema.ResourceData) (*systempb.BreakGlassAccess, error) {
	log.Println("expand break glass access")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand BreakGlassAccess empty input")
	}
	obj := &systempb.BreakGlassAccess{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandBreakGlassAccessSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "system.k8smgmt.io/v3"
	obj.Kind = "BreakGlassAccess"

	return obj, nil
}

func expandBreakGlassAccessSpec(p []interface{}) (*systempb.BreakGlassAccessSpec, error) {
	obj := &systempb.BreakGlassAccessSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("expandBreakGlassAccessSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["groups"].([]interface{}); ok && len(v) > 0 {
		obj.Groups = expandGroups(v)
	}

	return obj, nil
}

func expandGroups(p []interface{}) []*systempb.GroupSpec {
	groups := make([]*systempb.GroupSpec, len(p))
	for i, group := range p {
		groupMap := group.(map[string]interface{})
		g := &systempb.GroupSpec{}

		if v, ok := groupMap["user_type"].(string); ok && len(v) > 0 {
			g.UserType = v
		}

		if v, ok := groupMap["group_expiry"].([]interface{}); ok && len(v) > 0 {
			g.GroupExpiry = expandGroupExpiry(v)
		} else if v, ok := groupMap["group_expiry"].(*schema.Set); ok && v != nil && v.Len() > 0 {
			g.GroupExpiry = expandGroupExpiry(v.List())
		}

		groups[i] = g
	}
	return groups
}

func expandGroupExpiry(p []interface{}) []*systempb.GroupExpiryDetails {
	groupExpiries := make([]*systempb.GroupExpiryDetails, len(p))
	for i, expiry := range p {
		expiryMap := expiry.(map[string]interface{})
		ge := &systempb.GroupExpiryDetails{}

		if v, ok := expiryMap["expiry"].(float64); ok {
			ge.Expiry = v
		}

		if v, ok := expiryMap["timezone"].(string); ok && len(v) > 0 {
			ge.Timezone = v
		}

		if v, ok := expiryMap["name"].(string); ok && len(v) > 0 {
			ge.Name = v
		}

		if v, ok := expiryMap["start_time"].(string); ok && len(v) > 0 {
			ge.StartTime = v
		}
		groupExpiries[i] = ge
	}
	return groupExpiries
}

// Flatteners
func flattenBreakGlassAccess(d *schema.ResourceData, in *systempb.BreakGlassAccess) error {
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
	ret, err = flattenBreakGlassAccessSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten break glass access spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenBreakGlassAccessSpec(in *systempb.BreakGlassAccessSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("flattenBreakGlassAccessSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Groups != nil {
		obj["groups"] = flattenGroups(in.Groups)
	}

	return []interface{}{obj}, nil
}

func flattenGroups(groups []*systempb.GroupSpec) []interface{} {
	flattenedGroups := make([]interface{}, len(groups))
	for i, group := range groups {
		groupMap := map[string]interface{}{}

		if len(group.UserType) > 0 {
			groupMap["user_type"] = group.UserType
		}

		if group.GroupExpiry != nil {
			groupMap["group_expiry"] = flattenGroupExpiry(group.GroupExpiry)
		}

		flattenedGroups[i] = groupMap
	}
	return flattenedGroups
}

func flattenGroupExpiry(groupExpiries []*systempb.GroupExpiryDetails) []interface{} {
	flattenedGroupExpiries := make([]interface{}, len(groupExpiries))
	for i, expiry := range groupExpiries {
		expiryMap := map[string]interface{}{}

		expiryMap["expiry"] = expiry.Expiry

		if len(expiry.Name) > 0 {
			expiryMap["name"] = expiry.Name
		}
		if len(expiry.Timezone) > 0 {
			expiryMap["timezone"] = expiry.Timezone
		}

		expiryMap["start_time"] = expiry.StartTime

		flattenedGroupExpiries[i] = expiryMap
	}
	return flattenedGroupExpiries
}
