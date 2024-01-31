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
	"github.com/RafaySystems/rafay-common/proto/types/hub/infrapb"
	"github.com/RafaySystems/rctl/pkg/addon"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAddon() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAddonCreate,
		ReadContext:   resourceAddonRead,
		UpdateContext: resourceAddonUpdate,
		DeleteContext: resourceAddonDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.AddonSchema.Schema,
	}
}

func resourceAddonCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("addon create starts")
	diags := resourceAddonUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("addon create got error, perform cleanup")
		ns, err := expandAddon(d)
		if err != nil {
			log.Printf("addon expandAddon error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.InfraV3().Addon().Delete(ctx, options.DeleteOptions{
			Name:    ns.Metadata.Name,
			Project: ns.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceAddonUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("addon update starts")
	return resourceAddonUpsert(ctx, d, m)
}

func resourceAddonUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("addon upsert starts")
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

	addon, err := expandAddon(d)
	if err != nil {
		log.Printf("addon expandAddon error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Addon().Apply(ctx, addon, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", addon)
		log.Println("error addon apply addon:", n1)
		return diag.FromErr(err)
	}

	d.SetId(addon.Metadata.Name)
	return diags

}

func resourceAddonDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	addon, err := expandAddon(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.InfraV3().Addon().Delete(ctx, options.DeleteOptions{
		Name:    addon.Metadata.Name,
		Project: addon.Metadata.Project,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourceAddonV2Delete(ctx, addon)
	}

	return diags
}

func resourceAddonRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceAddonRead ")
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

	//tfAddonState, err := expandAddon(d)
	//if err != nil {
	//	return diag.FromErr(err)
	//}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfAddonState)
	// log.Println("resourceAddonRead tfAddonState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	addon, err := client.InfraV3().Addon().Get(ctx, options.GetOptions{
		Name: meta.Name,
		//Name:    tfAddonState.Metadata.Name,
		Project: meta.Project,
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
	addst := spew.Sprintf("%+v", addon)
	log.Println("resourceAddonRead addst", addst)

	err = flattenAddon(d, addon)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func expandAddon(in *schema.ResourceData) (*infrapb.Addon, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand addon empty input")
	}
	obj := &infrapb.Addon{}

	if v, ok := in.Get("metadata").([]interface{}); ok {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok {
		objSpec, err := expandAddonSpec(v)
		if err != nil {
			return nil, err
		}
		// XXX Debug
		addOn := spew.Sprintf("%+v", objSpec)
		log.Println("expandAddonSpec  ", addOn)
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Addon"
	return obj, nil
}

func expandAddonSpec(p []interface{}) (*infrapb.AddonSpec, error) {
	obj := &infrapb.AddonSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAddonSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["artifact"].([]interface{}); ok && len(v) > 0 {
		// XXX Debug
		artfct := spew.Sprintf("%+v", v)
		log.Println("expandAddonSpec before expand ", artfct)

		objArtifact, err := ExpandArtifactSpec(v)
		if err != nil {
			return nil, err
		}
		// XXX Debug
		artfct = spew.Sprintf("%+v", objArtifact.Artifact)
		log.Println("expandAddonSpec Artifact after expand ", artfct)
		artfct = spew.Sprintf("%+v", objArtifact.Options)
		log.Println("expandAddonSpec Options after expand ", artfct)

		obj.Artifact = objArtifact
	}
	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	return obj, nil
}

// Flatteners

func flattenAddon(d *schema.ResourceData, in *infrapb.Addon) error {
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
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenAddon before ", w1)
	var ret []interface{}
	ret, err = flattenAddonSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenAddon after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenAddonSpec(in *infrapb.AddonSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenAddonSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	v, ok := obj["artifact"].([]interface{})
	if !ok {
		v = []interface{}{}
	}
	// XXX Debug
	// w1 := spew.Sprintf("%+v", v)
	// log.Println("flattenAddonSpec before ", w1)

	var ret []interface{}
	var err error
	ret, err = FlattenArtifactSpec(in.Artifact, v)
	if err != nil {
		log.Println("FlattenArtifactSpec error ", err)
		return nil, err
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenAddonSpec after ", w1)

	obj["artifact"] = ret

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	return []interface{}{obj}, nil
}

func resourceAddonV2Delete(ctx context.Context, addonp *infrapb.Addon) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(addonp.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}

	if addonp.Spec.Artifact != nil && addonp.Spec.Artifact.GetType() == "alertmanager" {
		errDel := addon.DeleteManagedAddon(addonp.Metadata.Name, projectId)
		if errDel != nil {
			log.Printf("delete addon error %s", errDel.Error())
			return diag.FromErr(errDel)
		}
	} else {
		errDel := addon.DeleteAddon(addonp.Metadata.Name, projectId)
		if errDel != nil {
			log.Printf("delete addon error %s", errDel.Error())
			return diag.FromErr(errDel)
		}
	}
	return diags
}
