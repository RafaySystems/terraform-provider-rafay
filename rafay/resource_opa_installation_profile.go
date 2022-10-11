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
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/opapb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOPAInstallationProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOPAInstallationProfileCreate,
		ReadContext:   resourceOPAInstallationProfileRead,
		UpdateContext: resourceOPAInstallationProfileUpdate,
		DeleteContext: resourceOPAInstallationProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceOPAInstallationProfileImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.OPAProfileSchema.Schema,
	}
}

func resourceOPAInstallationProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceOPAInstallationProfileCreate reate starts")
	diags := resourceOPAInstallationProfileUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Opa profile create got error, perform cleanup")
		ss, err := expandOPAInstallationProfile(d)
		if err != nil {
			log.Printf("Opa profile expandOPAInstallationProfile error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.OpaV3().OPAProfile().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceOPAInstallationProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Opa profile update starts")
	return resourceOPAInstallationProfileUpsert(ctx, d, m)
}

func resourceOPAInstallationProfileUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Opa profile upsert starts")
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

	opaInstallationProfile, err := expandOPAInstallationProfile(d)
	if err != nil {
		log.Printf("Opa profile expandOPAInstallationProfile error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAProfile().Apply(ctx, opaInstallationProfile, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", opaInstallationProfile)
		log.Println("Opa profile apply Opa opa:", n1)
		log.Printf("Opa profile apply error")
		return diag.FromErr(err)
	}

	d.SetId(opaInstallationProfile.Metadata.Name)
	return diags

}

func resourceOPAInstallationProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceOPAInstallationProfileRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfOPAInstallationProfileState, err := expandOPAInstallationProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.OpaV3().OPAProfile().Get(ctx, options.GetOptions{
		Name:    tfOPAInstallationProfileState.Metadata.Name,
		Project: tfOPAInstallationProfileState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenOPAInstallationProfile(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceOPAInstallationProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandOPAInstallationProfile(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAProfile().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	return diags
}

func expandOPAInstallationProfile(in *schema.ResourceData) (*opapb.OPAProfile, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand Opa profile empty input")
	}
	obj := &opapb.OPAProfile{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandOPAInstallationProfileSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandOPAInstallationProfileSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "opa.k8smgmt.io/v3"
	obj.Kind = "OPAProfile"
	return obj, nil
}

func expandOPAInstallationProfileSpec(p []interface{}) (*opapb.OPAProfileSpec, error) {
	obj := &opapb.OPAProfileSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOPAInstallationProfileSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	if v, ok := in["installation_params"].([]interface{}); ok && len(v) > 0 {
		obj.InstallationParams = expandOpaInstallationsProfileInstallationParams(v)
	}

	if v, ok := in["sync_objects"].([]interface{}); ok && len(v) > 0 {
		obj.SyncObjects = expandOpaInstallationProfileSyncObjects(v)
	}

	if v, ok := in["excluded_namespaces"].([]interface{}); ok && len(v) > 0 {
		obj.ExcludedNamespaces = expandOpaInstallationProfileExcludedNamespaces(v)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	return obj, nil
}

func expandOpaInstallationProfileSyncObjects(p []interface{}) []*opapb.SyncObject {
	if len(p) == 0 || p[0] == nil {
		return []*opapb.SyncObject{}
	}

	out := make([]*opapb.SyncObject, len(p))

	for i := range p {
		obj := opapb.SyncObject{}
		in := p[i].(map[string]interface{})

		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		if v, ok := in["group"].(string); ok && len(v) > 0 {
			obj.Group = v
		}

		if v, ok := in["kind"].(string); ok && len(v) > 0 {
			obj.Kind = v
		}

		out[i] = &obj

	}

	return out
}

func expandOpaInstallationsProfileInstallationParams(p []interface{}) *opapb.InstallationParams {
	obj := &opapb.InstallationParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["audit_interval"].(int); ok {
		obj.AuditInterval = int64(v)
	}

	if v, ok := in["constraint_violations_limit"].(int); ok {
		obj.ConstraintViolationsLimit = int64(v)
	}

	if v, ok := in["audit_from_cache"].(bool); ok {
		obj.AuditFromCache = v
	}

	if v, ok := in["audit_match_kind_only"].(bool); ok {
		obj.AuditMatchKindOnly = v
	}

	if v, ok := in["enable_delete_operations"].(bool); ok {
		obj.EnableDeleteOperations = v
	}

	if v, ok := in["audit_chunk_size"].(int); ok {
		obj.AuditChunkSize = int64(v)
	}

	if v, ok := in["experimental_enable_mutation"].(bool); ok {
		obj.ExperimentalEnableMutation = v
	}

	if v, ok := in["log_denies"].(bool); ok {
		obj.LogDenies = v
	}

	if v, ok := in["emit_admission_events"].(bool); ok {
		obj.EmitAdmissionEvents = v
	}

	if v, ok := in["emit_audit_events"].(bool); ok {
		obj.EmitAuditEvents = v
	}

	return obj
}

func expandOpaInstallationProfileExcludedNamespaces(p []interface{}) []*opapb.ExcludedNamespaces {
	if len(p) == 0 || p[0] == nil {
		return []*opapb.ExcludedNamespaces{}
	}

	out := make([]*opapb.ExcludedNamespaces, len(p))

	for i := range p {
		obj := opapb.ExcludedNamespaces{}
		in := p[i].(map[string]interface{})

		if v, ok := in["processes"].([]interface{}); ok && len(v) > 0 {
			obj.Processes = toArrayString(v)
		}

		if v, ok := in["namespaces"].([]interface{}); ok && len(v) > 0 {
			obj.Namespaces = expandOpaInstallationProfileExcludedNamespacesList(v)
		}

		out[i] = &obj

	}

	return out
}

func expandOpaInstallationProfileExcludedNamespacesList(p []interface{}) []*commonpb.ResourceRef {
	if len(p) == 0 || p[0] == nil {
		return []*commonpb.ResourceRef{}
	}

	out := make([]*commonpb.ResourceRef, len(p))

	for i := range p {
		obj := commonpb.ResourceRef{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		out[i] = &obj
	}
	return out
}

// Flatten

func flattenOPAInstallationProfile(d *schema.ResourceData, in *opapb.OPAProfile) error {
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
	ret, err = flattenOPAInstallationProfileSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenOPAInstallationProfileSpec(in *opapb.OPAProfileSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenOPAInstallationProfile empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	if in.InstallationParams != nil {
		v, ok := obj["installation_params"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["installation_params"] = flattenOpaProfileInstallationParams(in.InstallationParams, v)
	}

	if in.SyncObjects != nil && len(in.SyncObjects) > 0 {
		v, ok := obj["sync_objects"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["sync_objects"] = flattenOpaProfileSyncObjects(in.SyncObjects, v)
	}

	if in.ExcludedNamespaces != nil && len(in.ExcludedNamespaces) > 0 {
		v, ok := obj["excluded_namespaces"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["excluded_namespaces"] = flattenOpaProfileExcludedNamespaces(in.ExcludedNamespaces, v)
	}

	if in.Sharing != nil {
		obj["sharing"] = flattenSharingSpec(in.Sharing)
	}

	return []interface{}{obj}, nil
}

func flattenOpaProfileInstallationParams(in *opapb.InstallationParams, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.AuditFromCache {
		obj["audit_from_cache"] = in.AuditFromCache
	}

	if in.AuditMatchKindOnly {
		obj["audit_match_kind_only"] = in.AuditMatchKindOnly
	}

	if in.EnableDeleteOperations {
		obj["enable_delete_operations"] = in.EnableDeleteOperations
	}

	if in.ExperimentalEnableMutation {
		obj["experimental_enable_mutation"] = in.ExperimentalEnableMutation
	}

	if in.AuditFromCache {
		obj["audit_from_cache"] = in.AuditFromCache
	}

	if in.AuditChunkSize != 0 {
		obj["audit_chunk_size"] = in.AuditChunkSize
	}

	if in.ConstraintViolationsLimit != 0 {
		obj["constraint_violations_limit"] = in.ConstraintViolationsLimit
	}

	if in.AuditInterval != 0 {
		obj["audit_interval"] = in.AuditInterval
	}

	return []interface{}{obj}
}

func flattenOpaProfileSyncObjects(input []*opapb.SyncObject, p []interface{}) []interface{} {
	log.Println("flattenSyncObjects")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Group) > 0 {
			obj["group"] = in.Group
		}

		if len(in.Version) > 0 {
			obj["version"] = in.Version
		}

		if len(in.Kind) > 0 {
			obj["kind"] = in.Kind
		}

		out[i] = &obj
	}

	return out
}

func flattenOpaProfileExcludedNamespaces(input []*opapb.ExcludedNamespaces, p []interface{}) []interface{} {
	log.Println("flattenExcludedNamespaces")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if in.Namespaces != nil && len(in.Namespaces) > 0 {
			v, ok := obj["namespaces"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["namespaces"] = flattenNamespacesRef(in.Namespaces, v)
		}

		if len(in.Processes) > 0 {
			obj["processes"] = toArrayInterface(in.Processes)
		}

		out[i] = &obj
	}

	return out
}

func flattenOpaProfileNamespacesRef(input []*commonpb.ResourceRef, p []interface{}) []interface{} {
	log.Println("flattenNamespacesRef")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}
		out[i] = &obj
	}

	return out
}

func resourceOPAInstallationProfileImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceOPAInstallationProfile idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceOPAInstallationProfile d.Id:", d.Id())
	log.Println("resourceOPAInstallationProfile d_debug", d_debug)

	opaConstraint, err := expandOPAConstraint(d)
	if err != nil {
		log.Printf("resourceOPAInstallationProfile expand error")
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	opaConstraint.Metadata = &metaD

	err = d.Set("metadata", flattenMetaData(opaConstraint.Metadata))
	if err != nil {
		log.Println("import set err")
		return nil, err
	}
	d.SetId(opaConstraint.Metadata.Name)
	return []*schema.ResourceData{d}, nil
}
