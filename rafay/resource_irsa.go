package rafay

import (
	"context"
	"fmt"

	"log"
	"os"
	"time"

	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/iamserviceaccount"

	log2 "github.com/RafaySystems/rctl/pkg/log"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceIRSA() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIRSACreate,
		ReadContext:   resourceIRSARead,
		UpdateContext: resourceIRSAUpdate,
		DeleteContext: resourceIRSADelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"metadata": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "cluster yaml file",
				Elem: &schema.Resource{
					Schema: irsaMetadataField(),
				},
			},
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "cluster config yaml file",
				Elem: &schema.Resource{
					Schema: irsaSpecField(),
				},
			},
		},
	}
}

func irsaMetadataField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "name of iam service accoount",
		},
		"labels": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "labels of iam service account",
		},
		"annotations": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "annotations of iam service account",
		},
	}
	return s
}
func irsaSpecField() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"cluster_name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "name of cluster",
		},
		"namespace": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "namespace",
		},
		"permissions_boundary": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "",
		},
		"role_only": {
			Type:     schema.TypeBool,
			Optional: true,
			//what is the default supposed to be?
			Default:     false,
			Description: "Specify if only the IAM Service Account role should be created without creating/annotating the service account",
		},
		"policy_arns": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "attach polciy ARN",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
		"tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "tags of iam service account",
		},
		"policy_document": { //how do we deal with this map[string]interface
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom address used for DNS lookups",
			Computed:    true,
			/*Elem: &schema.Resource{
				Schema: DataSourcePolicyDocument(),
			},*/
		},
	}
	return s
}

/*
func DataSourcePolicyDocument() map[string]*schema.Schema {
	setOfString := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
	s := map[string]*schema.Schema{
		"json": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"override_json": {
			Type:       schema.TypeString,
			Optional:   true,
			Deprecated: "Use the attribute \"override_policy_documents\" instead.",
		},
		"override_policy_documents": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"policy_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"source_json": {
			Type:       schema.TypeString,
			Optional:   true,
			Deprecated: "Use the attribute \"source_policy_documents\" instead.",
		},
		"source_policy_documents": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"statement": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"actions": setOfString,
					"condition": {
						Type:     schema.TypeSet,
						Optional: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"test": {
									Type:     schema.TypeString,
									Required: true,
								},
								"values": {
									Type:     schema.TypeList,
									Required: true,
									Elem: &schema.Schema{
										Type: schema.TypeString,
									},
								},
								"variable": {
									Type:     schema.TypeString,
									Required: true,
								},
							},
						},
					},
					"effect": {
						Type:         schema.TypeString,
						Optional:     true,
						Default:      "Allow",
						ValidateFunc: validation.StringInSlice([]string{"Allow", "Deny"}, false),
					},
					"not_actions":    setOfString,
					"not_principals": dataSourcePolicyPrincipalSchema(),
					"not_resources":  setOfString,
					"principals":     dataSourcePolicyPrincipalSchema(),
					"resources":      setOfString,
					"sid": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"version": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "2012-10-17",
			ValidateFunc: validation.StringInSlice([]string{
				"2008-10-17",
				"2012-10-17",
			}, false),
		},
	}
	return s
}
func dataSourcePolicyPrincipalSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"identifiers": {
					Type:     schema.TypeSet,
					Required: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
		},
	}
}*/
func resourceIRSACreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var logger log2.Logger
	var config *config.Config
	log.Printf("IRSA create starts")

	//what is happening here? copied from resource_agent.go
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	irsa, err := expandIRSA(d)
	if err != nil {
		log.Printf("irsa expandIRSA error")
		return diag.FromErr(err)
	}

	rctlERR := iamserviceaccount.Create(logger, config, irsa.Spec.ClusterName, irsa.Metadata.Name, irsa.Spec.Namespace, irsa.Spec.PolicyARNs, irsa.Spec.PolicyDocument, irsa.Spec.PermissionsBoundary, irsa.Metadata.Labels, irsa.Metadata.Annotations, irsa.Spec.Tags, *irsa.Spec.RoleOnly)
	if rctlERR != nil {
		log.Println("rclt create err")
		return diag.FromErr(rctlERR)
	}
	log.Println("create with no errs")
	return diags
}

func resourceIRSAUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var logger log2.Logger
	var config *config.Config
	log.Printf("IRSA update starts")
	//what is happening here? copied from resource_agent.go
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	irsa, err := expandIRSA(d)
	if err != nil {
		log.Printf("irsa expandIRSA error")
		return diag.FromErr(err)
	}

	rctlERR := iamserviceaccount.Update(logger, config, irsa.Spec.ClusterName, irsa.Metadata.Name, irsa.Spec.Namespace, irsa.Spec.PolicyARNs, irsa.Spec.PolicyDocument, irsa.Spec.PermissionsBoundary, irsa.Spec.Tags)
	if rctlERR != nil {
		log.Println("rclt update err")
		return diag.FromErr(rctlERR)
	}
	return diags
}

func resourceIRSARead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var logger log2.Logger
	var config *config.Config

	log.Println("resourceAgentRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	irsa, err := expandIRSA(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfAgentState)
	// log.Println("resourceAgentRead tfAgentState", w1)

	resp, rctlErr := iamserviceaccount.Get(logger, config, irsa.Spec.ClusterName, irsa.Metadata.Name, irsa.Spec.Namespace)
	if rctlErr != nil {
		log.Println("rctl get err")
		return diag.FromErr(err)
	}
	log.Println("resp from get:", resp)
	//what doe we do with resp / how do we unmarshal to get irsa?
	//irsa := json.Unmarshal([]byte(resp))
	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceAgentRead wl", w1)

	err = flattenIRSA(d, irsa)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceIRSADelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var logger log2.Logger
	var config *config.Config
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	irsa, err := expandIRSA(d)
	if err != nil {
		return diag.FromErr(err)
	}
	//delete irsa
	err = iamserviceaccount.Delete(logger, config, irsa.Spec.ClusterName, irsa.Metadata.Name, irsa.Spec.Namespace)
	if err != nil {
		log.Println("error deleting irsa")
	} else {
		log.Println("Deleted irsa: ", irsa.Metadata.Name)
	}
	return diags
}

func expandIRSA(in *schema.ResourceData) (*IRSA, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand agent empty input")
	}
	obj := &IRSA{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandIRSASpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandIRSASpec got spec")
		obj.Spec = objSpec
	}
	return obj, nil
}
func expandIRSASpec(p []interface{}) (*IRSASpec, error) {
	obj := &IRSASpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAgentSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["cluster_name"].(string); ok && len(v) > 0 {
		obj.ClusterName = v
	}
	if v, ok := in["namespace"].(string); ok && len(v) > 0 {
		obj.Namespace = v
	}
	if v, ok := in["permissions_boundary"].(string); ok && len(v) > 0 {
		obj.PermissionsBoundary = v
	}
	if v, ok := in["role_only"].(bool); ok {
		obj.RoleOnly = &v
	}
	if v, ok := in["tags"].(map[string]interface{}); ok && len(v) > 0 {
		obj.Tags = toMapString(v)
	}
	if v, ok := in["policy_arns"].([]interface{}); ok && len(v) > 0 {
		obj.PolicyARNs = toArrayStringSorted(v)
	}

	if v, ok := in["policy_document"].([]interface{}); ok && len(v) > 0 {
		obj.PolicyDocument = expandPolicyDocument(v)
	}

	return obj, nil
}

func expandPolicyDocument(in []interface{}) map[string]interface{} {
	if in == nil {
		return nil
	}
	obj := make(map[string]interface{})
	log.Println("expand policy document:", in)
	//what do for map[string]interface object
	return obj
}

func flattenIRSA(d *schema.ResourceData, in *IRSA) error {
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
	// log.Println("flattenAgent before ", w1)
	//var ret []interface{}
	spec := flattenIRSASpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenAgent after ", w1)

	err = d.Set("spec", spec)
	if err != nil {
		return err
	}
	return nil
}

func flattenIRSASpec(in *IRSASpec, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.ClusterName) > 0 {
		obj["cluster_name"] = in.ClusterName
	}

	if len(in.Namespace) > 0 {
		obj["namespace"] = in.Namespace
	}

	if len(in.PermissionsBoundary) > 0 {
		obj["permissions_boundary"] = in.PermissionsBoundary
	}

	obj["role_only"] = in.RoleOnly

	if in.Tags != nil && len(in.Tags) > 0 {
		obj["tags"] = toMapInterface(in.Tags)
	}

	if len(in.PolicyARNs) > 0 {
		obj["policy_arns"] = toArrayInterface(in.PolicyARNs)
	}

	//what do we do for policy document?

	return []interface{}{obj}
}
