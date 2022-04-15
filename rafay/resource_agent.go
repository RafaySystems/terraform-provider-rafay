package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	typed "github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rctl/pkg/agent"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/versioninfo"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAgent() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAgentCreate,
		ReadContext:   resourceAgentRead,
		UpdateContext: resourceAgentUpdate,
		DeleteContext: resourceAgentDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.AgentSchema.Schema,
	}
}

func resourceAgentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("agent create starts")
	diags := resourceAgentUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("agent create got error, perform cleanup")
		ag, err := expandAgent(d)
		if err != nil {
			log.Printf("agent expandAgent error")
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
		if err != nil {
			return diags
		}

		err = client.GitopsV3().Agent().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceAgentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("agent update starts")
	return resourceAgentUpsert(ctx, d, m)
}

func resourceAgentUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("agent upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	agent, err := expandAgent(d)
	if err != nil {
		log.Printf("agent expandAgent error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().Agent().Apply(ctx, agent, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", agent)
		log.Println("agent apply agent:", n1)
		log.Printf("agent apply error")
		return diag.FromErr(err)
	}

	d.SetId(agent.Metadata.Name)
	return diags

}

func resourceAgentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceAgentRead ")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}
	tfAgentState, err := expandAgent(d)
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfAgentState)
	// log.Println("resourceAgentRead tfAgentState", w1)

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.GitopsV3().Agent().Get(ctx, options.GetOptions{
		Name:    tfAgentState.Metadata.Name,
		Project: tfAgentState.Metadata.Project,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	// XXX Debug
	// w1 = spew.Sprintf("%+v", wl)
	// log.Println("resourceAgentRead wl", w1)

	err = flattenAgent(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceAgentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandAgent(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, versioninfo.GetUserAgent())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().Agent().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	if err != nil {
		//v3 spec gave error try v2
		return resourceAgentV2Delete(ctx, ag)
	}

	return diags
}

func resourceAgentV2Delete(ctx context.Context, ag *gitopspb.Agent) diag.Diagnostics {
	var diags diag.Diagnostics

	projectId, err := config.GetProjectIdByName(ag.Metadata.Project)
	if err != nil {
		return diag.FromErr(err)
	}
	//delete agent
	err = agent.DeleteAgent(ag.Metadata.Name, projectId)
	if err != nil {
		log.Println("error deleting agent")
	} else {
		log.Println("Deleted agent: ", ag.Metadata.Name)
	}
	return diags
}

func expandAgent(in *schema.ResourceData) (*gitopspb.Agent, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand agent empty input")
	}
	obj := &gitopspb.Agent{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandAgentSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandAgentSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "infra.k8smgmt.io/v3"
	obj.Kind = "Agent"
	return obj, nil
}

func expandAgentSpec(p []interface{}) (*gitopspb.AgentSpec, error) {
	obj := &gitopspb.AgentSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAgentSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["type"].(string); ok && len(v) > 0 {
		obj.Type = v
	}

	if v, ok := in["cluster"].([]interface{}); ok && len(v) > 0 {
		obj.Cluster = expandAgentClusterMeta(v)
	}

	if v, ok := in["active"].(bool); ok {
		obj.Active = v
	}

	return obj, nil
}

func expandAgentClusterMeta(p []interface{}) *gitopspb.ClusterMeta {
	obj := &gitopspb.ClusterMeta{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		obj.Name = v
	}

	if v, ok := in["id"].(string); ok && len(v) > 0 {
		obj.Id = v
	}

	return obj

}

// Flatteners

func flattenAgent(d *schema.ResourceData, in *gitopspb.Agent) error {
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
	var ret []interface{}
	ret, err = flattenAgentSpec(in.Spec, v)
	if err != nil {
		return err
	}
	// XXX Debug
	// w1 = spew.Sprintf("%+v", ret)
	// log.Println("flattenAgent after ", w1)

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenAgentSpec(in *gitopspb.AgentSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenAgentSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Type) > 0 {
		obj["type"] = in.Type
	}

	if in.Cluster != nil {
		v, ok := obj["cluster"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["cluster"] = flattenAgentClusterMeta(in.Cluster, v)
	}

	obj["active"] = in.Active

	return []interface{}{obj}, nil
}

func flattenAgentClusterMeta(in *gitopspb.ClusterMeta, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Name) > 0 {
		obj["name"] = in.Name
	}

	if len(in.Id) > 0 {
		obj["id"] = in.Id
	}

	return []interface{}{obj}
}
