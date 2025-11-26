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

func resourceOPAPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOPAPolicyCreate,
		ReadContext:   resourceOPAPolicyRead,
		UpdateContext: resourceOPAPolicyUpdate,
		DeleteContext: resourceOPAPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: resourceOPAPolicyImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.OPAPolicySchema.Schema,
	}
}

func resourceOPAPolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("resourceOPAConstraintCreate create starts")
	diags := resourceOPAPolicyUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("Opa policy create got error, perform cleanup")
		ss, err := expandOPAPolicy(d)
		if err != nil {
			log.Printf("Opa policy expandOPAPolicy error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.OpaV3().OPAPolicy().Delete(ctx, options.DeleteOptions{
			Name:    ss.Metadata.Name,
			Project: ss.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceOPAPolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("Opa policy update starts")
	return resourceOPAPolicyUpsert(ctx, d, m)
}

func resourceOPAPolicyUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("Opa policy upsert starts")
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

	opaPolicy, err := expandOPAPolicy(d)
	if err != nil {
		log.Printf("Opa policy expandOPAPolicy error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAPolicy().Apply(ctx, opaPolicy, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", opaPolicy)
		log.Println("Opa policy apply Opa policy:", n1)
		log.Printf("Opa policy apply error")
		return diag.FromErr(err)
	}

	d.SetId(opaPolicy.Metadata.Name)
	return diags

}

func resourceOPAPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceOPAPolicyRead ")
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

	tfOPAPolicyState, err := expandOPAPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	ag, err := client.OpaV3().OPAPolicy().Get(ctx, options.GetOptions{
		//Name:    tfOPAPolicyState.Metadata.Name,
		Name:    meta.Name,
		Project: tfOPAPolicyState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenOPAPolicy(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceOPAPolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandOPAPolicy(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent(), options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.OpaV3().OPAPolicy().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandOPAPolicy(in *schema.ResourceData) (*opapb.OPAPolicy, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand Opa policy empty input")
	}
	obj := &opapb.OPAPolicy{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandOPAPolicySpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandOPAPolicySpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "opa.k8smgmt.io/v3"
	obj.Kind = "OPAPolicy"
	return obj, nil
}

func expandOPAPolicySpec(p []interface{}) (*opapb.OPAPolicySpec, error) {
	obj := &opapb.OPAPolicySpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandOPAPolicySpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["version"].(string); ok && len(v) > 0 {
		obj.Version = v
	}

	// if v, ok := in["installation_params"].([]interface{}); ok && len(v) > 0 {
	// 	obj.InstallationParams = expandOpaPolicyInstallationParams(v)
	// }

	// if v, ok := in["sync_objects"].([]interface{}); ok && len(v) > 0 {
	// 	obj.SyncObjects = expandOpaPolicySyncObjects(v)
	// }

	// if v, ok := in["excluded_namespaces"].([]interface{}); ok && len(v) > 0 {
	// 	obj.ExcludedNamespaces = expandOpaPolicyExcludedNamespaces(v)
	// }

	if v, ok := in["constraint_list"].([]interface{}); ok && len(v) > 0 {
		obj.ConstraintList = expandOpaPolicyConstraintList(v)
	}

	return obj, nil
}

func expandOpaPolicySyncObjects(p []interface{}) []*opapb.SyncObject {
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

func expandOpaPolicyInstallationParams(p []interface{}) *opapb.InstallationParams {
	obj := &opapb.InstallationParams{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["audit_interval"].(int64); ok {
		obj.AuditInterval = v
	}

	if v, ok := in["constraint_violations_limit"].(int64); ok {
		obj.ConstraintViolationsLimit = v
	}

	if v, ok := in["audit_interval"].(bool); ok {
		obj.AuditFromCache = v
	}

	if v, ok := in["audit_match_kind_only"].(bool); ok {
		obj.AuditMatchKindOnly = v
	}

	if v, ok := in["enable_delete_operations"].(bool); ok {
		obj.EnableDeleteOperations = v
	}

	if v, ok := in["audit_chunk_size"].(int64); ok {
		obj.AuditChunkSize = v
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

func expandOpaPolicyExcludedNamespaces(p []interface{}) []*opapb.ExcludedNamespaces {
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
			obj.Namespaces = expandOpaPolicyExcludedNamespacesList(v)
		}

		out[i] = &obj

	}

	return out
}

func expandOpaPolicyExcludedNamespacesList(p []interface{}) []*commonpb.ResourceRef {
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

// func expandOpaPolicyConstraintList(p []interface{}) []*commonpb.ResourceRef {
// 	if len(p) == 0 || p[0] == nil {
// 		return []*commonpb.ResourceRef{}
// 	}

// 	out := make([]*commonpb.ResourceRef, len(p))

// 	for i := range p {
// 		obj := commonpb.ResourceRef{}
// 		in := p[i].(map[string]interface{})

// 		if v, ok := in["name"].(string); ok && len(v) > 0 {
// 			obj.Name = v
// 		}

// 		out[i] = &obj
// 	}
// 	return out
// }

func expandOpaPolicyConstraintList(p []interface{}) []*opapb.OPAPolicyConstraint {
	if len(p) == 0 || p[0] == nil {
		return []*opapb.OPAPolicyConstraint{}
	}

	out := make([]*opapb.OPAPolicyConstraint, len(p))

	for i := range p {
		obj := opapb.OPAPolicyConstraint{}
		in := p[i].(map[string]interface{})

		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}

		if v, ok := in["version"].(string); ok && len(v) > 0 {
			obj.Version = v
		}

		out[i] = &obj
	}
	return out
}

// Flatten

func flattenOPAPolicy(d *schema.ResourceData, in *opapb.OPAPolicy) error {
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
	ret, err = flattenOPAPolicySpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenOPAPolicySpec(in *opapb.OPAPolicySpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenOPAPolicy empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Version) > 0 {
		obj["version"] = in.Version
	}

	// if in.InstallationParams != nil {
	// 	v, ok := obj["installation_params"].([]interface{})
	// 	if !ok {
	// 		v = []interface{}{}
	// 	}
	// 	obj["installation_params"] = flattenInstallationParams(in.InstallationParams, v)
	// }

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	// if in.SyncObjects != nil && len(in.SyncObjects) > 0 {
	// 	v, ok := obj["sync_objects"].([]interface{})
	// 	if !ok {
	// 		v = []interface{}{}
	// 	}
	// 	obj["sync_objects"] = flattenSyncObjects(in.SyncObjects, v)
	// }

	// if in.ExcludedNamespaces != nil && len(in.ExcludedNamespaces) > 0 {
	// 	v, ok := obj["excluded_namespaces"].([]interface{})
	// 	if !ok {
	// 		v = []interface{}{}
	// 	}
	// 	obj["excluded_namespaces"] = flattenExcludedNamespaces(in.ExcludedNamespaces, v)
	// }

	if in.ConstraintList != nil && len(in.ConstraintList) > 0 {
		v, ok := obj["constraintList"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["constraint_list"] = flattenConstraintList(in.ConstraintList, v)
	}

	return []interface{}{obj}, nil
}

func flattenInstallationParams(in *opapb.InstallationParams, p []interface{}) []interface{} {
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

	obj["audit_chunk_size"] = in.AuditChunkSize

	obj["audit_interval"] = in.AuditInterval

	obj["constraint_violations_limit"] = in.ConstraintViolationsLimit

	return []interface{}{obj}
}

func flattenExcludedNamespaces(input []*opapb.ExcludedNamespaces, p []interface{}) []interface{} {
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

func flattenNamespacesRef(input []*commonpb.ResourceRef, p []interface{}) []interface{} {
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

func flattenConstraintList(input []*opapb.OPAPolicyConstraint, p []interface{}) []interface{} {
	log.Println("flattenConstraintList")
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

		if len(in.Name) > 0 {
			obj["version"] = in.Version
		}
		out[i] = &obj
	}

	return out
}

func flattenSyncObjects(input []*opapb.SyncObject, p []interface{}) []interface{} {
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

func resourceOPAPolicyImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceOPAPolicy idParts:", idParts)
	d_debug := spew.Sprintf("%+v", d)
	log.Println("resourceOPAPolicy d.Id:", d.Id())
	log.Println("resourceOPAPolicy d_debug", d_debug)

	opaConstraint, err := expandOPAConstraint(d)
	if err != nil {
		log.Printf("resourceOPAPolicy expand error")
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
