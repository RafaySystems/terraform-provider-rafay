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
	"github.com/RafaySystems/rafay-common/proto/types/hub/gitopspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAgentPool() *schema.Resource {
	modSchema := resource.AgentPoolSchema.Schema
	return &schema.Resource{
		CreateContext: resourceAgentPoolCreate,
		ReadContext:   resourceAgentPoolRead,
		UpdateContext: resourceAgentPoolUpdate,
		DeleteContext: resourceAgentPoolDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        modSchema,
	}
}

func resourceAgentPoolCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("agent pool create starts")
	diags := resourceAgentPoolUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		log.Printf("agent pool create got error, perform cleanup")
		ag, err := expandAgentPool(d)
		if err != nil {
			log.Printf("agent pool expandAgentPool error")
			return diags
		}

		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.GitopsV3().AgentPool().Delete(ctx, options.DeleteOptions{
			Name:    ag.Metadata.Name,
			Project: ag.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceAgentPoolUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("agent pool update starts")
	return resourceAgentPoolUpsert(ctx, d, m)
}

func resourceAgentPoolUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("agent pool upsert starts")
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

	agentPool, err := expandAgentPool(d)
	if err != nil {
		log.Printf("agent pool expandAgentPool error")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().AgentPool().Apply(ctx, agentPool, options.ApplyOptions{})
	if err != nil {
		// XXX Debug
		n1 := spew.Sprintf("%+v", agentPool)
		log.Println("agent pool apply agent:", n1)
		log.Printf("agent pool apply error")
		return diag.FromErr(err)
	}

	d.SetId(agentPool.Metadata.Name)
	return diags

}

func resourceAgentPoolRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("resourceAgentRead ")
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

	tfAgentPoolState, err := expandAgentPool(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	ag, err := client.GitopsV3().AgentPool().Get(ctx, options.GetOptions{
		//Name:    tfAgentState.Metadata.Name,
		Name:    meta.Name,
		Project: tfAgentPoolState.Metadata.Project,
	})
	if err != nil {
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if tfAgentPoolState.GetSpec().GetSharing() != nil && !tfAgentPoolState.GetSpec().GetSharing().GetEnabled() && ag.GetSpec().GetSharing() == nil {
		ag.Spec.Sharing = &commonpb.SharingSpec{}
		ag.Spec.Sharing.Enabled = false
		ag.Spec.Sharing.Projects = tfAgentPoolState.GetSpec().GetSharing().GetProjects()
	}

	err = flattenAgentPool(d, ag)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourceAgentPoolDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	ag, err := expandAgentPool(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.GitopsV3().AgentPool().Delete(ctx, options.DeleteOptions{
		Name:    ag.Metadata.Name,
		Project: ag.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandAgentPool(in *schema.ResourceData) (*gitopspb.AgentPool, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "expand agent empty input")
	}
	obj := &gitopspb.AgentPool{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandMetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandAgentPoolSpec(v)
		if err != nil {
			return nil, err
		}
		log.Println("expandAgentPoolSpec got spec")
		obj.Spec = objSpec
	}

	obj.ApiVersion = "gitops.k8smgmt.io/v3"
	obj.Kind = "AgentPool"

	return obj, nil
}

func expandAgentPoolSpec(p []interface{}) (*gitopspb.AgentPoolSpec, error) {
	obj := &gitopspb.AgentPoolSpec{}
	if len(p) == 0 || p[0] == nil {
		return obj, fmt.Errorf("%s", "expandAgentSpec empty input")
	}

	in := p[0].(map[string]interface{})

	if agents, ok := in["agents"].([]interface{}); ok && len(agents) > 0 {
		for _, agent := range agents {
			if agentStr, ok := agent.(string); ok && agentStr != "" {
				obj.Agents = append(obj.Agents, agentStr)
			}
		}
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		obj.Sharing = expandSharingSpec(v)
	}

	return obj, nil
}

// Flatteners

func flattenAgentPool(d *schema.ResourceData, in *gitopspb.AgentPool) error {
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
	ret, err = flattenAgentPoolSpec(in.Spec, v)
	if err != nil {
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenAgentPoolSpec(in *gitopspb.AgentPoolSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenAgentPoolSpec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Agents) > 0 {
		obj["agents"] = in.Agents
	}

	obj["sharing"] = flattenSharingSpec(in.Sharing)

	return []interface{}{obj}, nil
}
