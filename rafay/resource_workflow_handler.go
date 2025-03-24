package rafay

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/pkg/hub/client/options"
	"github.com/RafaySystems/rafay-common/pkg/hub/client/typed"
	"github.com/RafaySystems/rafay-common/pkg/hub/terraform/resource"
	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rafay-common/proto/types/hub/eaaspb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceWorkflowHandler() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWorkflowHandlerCreate,
		ReadContext:   resourceWorkflowHandlerRead,
		UpdateContext: resourceWorkflowHandlerUpdate,
		DeleteContext: resourceWorkflowHandlerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceWorkflowHandlerImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		SchemaVersion: 1,
		Schema:        resource.WorkflowHandlerSchema.Schema,
	}
}

func resourceWorkflowHandlerCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Println("workflow handler create")
	diags := resourceWorkflowHandlerUpsert(ctx, d, m)
	if diags.HasError() {
		tflog := os.Getenv("TF_LOG")
		if tflog == "TRACE" || tflog == "DEBUG" {
			ctx = context.WithValue(ctx, "debug", "true")
		}
		cc, err := expandWorkflowHandler(d)
		if err != nil {
			return diags
		}
		auth := config.GetConfig().GetAppAuthProfile()
		client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
		if err != nil {
			return diags
		}

		err = client.EaasV1().WorkflowHandler().Delete(ctx, options.DeleteOptions{
			Name:    cc.Metadata.Name,
			Project: cc.Metadata.Project,
		})
		if err != nil {
			return diags
		}
	}
	return diags
}

func resourceWorkflowHandlerUpsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("workflow handler upsert starts")
	tflog := os.Getenv("TF_LOG")
	if tflog == "TRACE" || tflog == "DEBUG" {
		ctx = context.WithValue(ctx, "debug", "true")
	}

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.EaasV1().WorkflowHandler().Apply(ctx, cc, options.ApplyOptions{})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(cc.Metadata.Name)
	return diags
}

func resourceWorkflowHandlerRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("workflow handler read starts ")
	meta := GetMetaData(d)
	if meta == nil {
		return diag.FromErr(fmt.Errorf("%s", "failed to read resource "))
	}
	if d.State() != nil && d.State().ID != "" {
		meta.Name = d.State().ID
	}

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		log.Println("read client err")
		return diag.FromErr(err)
	}

	wh, err := client.EaasV1().WorkflowHandler().Get(ctx, options.GetOptions{
		Name:    meta.Name,
		Project: cc.Metadata.Project,
	})
	if err != nil {
		log.Println("read get err")
		if strings.Contains(err.Error(), "code 404") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	err = flattenWorkflowHandler(d, wh)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceWorkflowHandlerUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return resourceWorkflowHandlerUpsert(ctx, d, m)
}

func resourceWorkflowHandlerDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Println("workflow handler delete starts")
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

	cc, err := expandWorkflowHandler(d)
	if err != nil {
		log.Println("error while expanding workflow handler during delete")
		return diag.FromErr(err)
	}

	auth := config.GetConfig().GetAppAuthProfile()
	client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.EaasV1().WorkflowHandler().Delete(ctx, options.DeleteOptions{
		Name:    cc.Metadata.Name,
		Project: cc.Metadata.Project,
	})

	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func expandWorkflowHandler(in *schema.ResourceData) (*eaaspb.WorkflowHandler, error) {
	log.Println("expand workflow handler resource")
	if in == nil {
		return nil, fmt.Errorf("%s", "expand workflow handler empty input")
	}
	obj := &eaaspb.WorkflowHandler{}

	if v, ok := in.Get("metadata").([]interface{}); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]interface{}); ok && len(v) > 0 {
		objSpec, err := expandWorkflowHandlerSpec(v)
		if err != nil {
			return nil, err
		}
		obj.Spec = objSpec
	}

	obj.ApiVersion = "eaas.envmgmt.io/v1"
	obj.Kind = "WorkflowHandler"
	return obj, nil
}

func expandWorkflowHandlerSpec(p []interface{}) (*eaaspb.WorkflowHandlerSpec, error) {
	log.Println("expand workflow handler spec")
	spec := &eaaspb.WorkflowHandlerSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand workflow handler spec empty input")
	}

	in := p[0].(map[string]interface{})

	if c, ok := in["config"].([]interface{}); ok && len(c) > 0 {
		spec.Config = expandWorkflowHandlerConfig(c)
	}

	if v, ok := in["sharing"].([]interface{}); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["inputs"].([]interface{}); ok && len(v) > 0 {
		spec.Inputs = expandConfigContextCompoundRefs(v)
	}

	var err error
	if v, ok := in["outputs"].(string); ok && len(v) > 0 {
		spec.Outputs, err = expandWorkflowHandlerOutputs(v)
		if err != nil {
			return nil, err
		}
	}

	return spec, nil
}

func expandWorkflowHandlerConfig(p []interface{}) *eaaspb.WorkflowHandlerConfig {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	workflowHandlerConfig := eaaspb.WorkflowHandlerConfig{}
	in := p[0].(map[string]interface{})

	if typ, ok := in["type"].(string); ok && len(typ) > 0 {
		workflowHandlerConfig.Type = typ
	}

	if ts, ok := in["timeout_seconds"].(int); ok {
		workflowHandlerConfig.TimeoutSeconds = int64(ts)
	}

	if sc, ok := in["success_condition"].(string); ok && len(sc) > 0 {
		workflowHandlerConfig.SuccessCondition = sc
	}

	if ts, ok := in["max_retry_count"].(int); ok {
		workflowHandlerConfig.MaxRetryCount = int32(ts)
	}

	if v, ok := in["container"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.Container = expandWorkflowHandlerContainerConfig(v)
	}

	if v, ok := in["http"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.Http = expandWorkflowHandlerHttpConfig(v)
	}

	if v, ok := in["polling_config"].([]interface{}); ok && len(v) > 0 {
		workflowHandlerConfig.PollingConfig = expandPollingConfig(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		workflowHandlerConfig.TimeoutSeconds = int64(h)
	}

	return &workflowHandlerConfig
}

func expandWorkflowHandlerContainerConfig(p []interface{}) *eaaspb.ContainerDriverConfig {
	cc := eaaspb.ContainerDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &cc
	}

	in := p[0].(map[string]interface{})

	if img, ok := in["image"].(string); ok && len(img) > 0 {
		cc.Image = img
	}

	if args, ok := in["arguments"].([]interface{}); ok && len(args) > 0 {
		cc.Arguments = toArrayString(args)
	}

	if cmds, ok := in["commands"].([]interface{}); ok && len(cmds) > 0 {
		cc.Commands = toArrayString(cmds)
	}

	if clm, ok := in["cpu_limit_milli"].(string); ok && len(clm) > 0 {
		cc.CpuLimitMilli = clm
	}

	if ev, ok := in["env_vars"].(map[string]interface{}); ok && len(ev) > 0 {
		cc.EnvVars = toMapString(ev)
	}

	if f, ok := in["files"].(map[string]interface{}); ok && len(f) > 0 {
		cc.Files = toMapByte(f)
	}

	if v, ok := in["image_pull_credentials"].([]interface{}); ok && len(v) > 0 {
		cc.ImagePullCredentials = expandImagePullCredentials(v)
	}

	if v, ok := in["kube_config_options"].([]interface{}); ok && len(v) > 0 {
		cc.KubeConfigOptions = expandKubeConfigOptions(v)
	}

	if v, ok := in["kube_options"].([]interface{}); ok && len(v) > 0 {
		cc.KubeOptions = expandContainerKubeOptions(v)
	}

	if mlb, ok := in["memory_limit_mb"].(string); ok && len(mlb) > 0 {
		cc.MemoryLimitMb = mlb
	}

	if v, ok := in["volume_options"].([]interface{}); ok && len(v) > 0 {
		volumes := expandContainerWorkflowHandlerVolumeOptions(v)
		if len(volumes) > 0 {
			cc.VolumeOptions = volumes[0]
		}
	}

	if v, ok := in["volumes"].([]interface{}); ok && len(v) > 0 {
		cc.Volumes = expandContainerWorkflowHandlerVolumeOptions(v)
	}

	if wdp, ok := in["working_dir_path"].(string); ok && len(wdp) > 0 {
		cc.WorkingDirPath = wdp
	}

	return &cc
}

func expandImagePullCredentials(p []interface{}) *eaaspb.ContainerImagePullCredentials {
	hc := eaaspb.ContainerImagePullCredentials{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if pass, ok := in["password"].(string); ok && len(pass) > 0 {
		hc.Password = pass
	}

	if registry, ok := in["registry"].(string); ok && len(registry) > 0 {
		hc.Registry = registry
	}

	if username, ok := in["username"].(string); ok && len(username) > 0 {
		hc.Username = username
	}

	return &hc
}

func expandKubeConfigOptions(p []interface{}) *eaaspb.ContainerKubeConfigOptions {
	hc := eaaspb.ContainerKubeConfigOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if kc, ok := in["kube_config"].(string); ok && len(kc) > 0 {
		hc.KubeConfig = kc
	}

	if ofc, ok := in["out_of_cluster"].(bool); ok {
		hc.OutOfCluster = ofc
	}

	return &hc
}

func expandContainerKubeOptions(p []interface{}) *eaaspb.ContainerKubeOptions {
	hc := eaaspb.ContainerKubeOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if lbls, ok := in["labels"].(map[string]interface{}); ok && len(lbls) > 0 {
		hc.Labels = toMapString(lbls)
	}

	if ns, ok := in["namespace"].(string); ok && len(ns) > 0 {
		hc.Namespace = ns
	}

	if ns, ok := in["node_selector"].(map[string]interface{}); ok && len(ns) > 0 {
		hc.NodeSelector = toMapString(ns)
	}

	if r, ok := in["resources"].([]interface{}); ok && len(r) > 0 {
		hc.Resources = toArrayString(r)
	}

	if sc, ok := in["security_context"].([]interface{}); ok && len(sc) > 0 {
		hc.SecurityContext = expandSecurityContext(sc)
	}

	if san, ok := in["service_account_name"].(string); ok && len(san) > 0 {
		hc.ServiceAccountName = san
	}

	if tolerations, ok := in["tolerations"].([]interface{}); ok {
		hc.Tolerations = expandV3Tolerations(tolerations)
	}

	if a, ok := in["affinity"].([]interface{}); ok && len(a) > 0 {
		hc.Affinity = expandKubeOptionsAffinity(a)
	}

	return &hc
}

func expandSecurityContext(p []interface{}) *eaaspb.KubeSecurityContext {
	ksc := eaaspb.KubeSecurityContext{}
	if len(p) == 0 || p[0] == nil {
		return &ksc
	}

	in := p[0].(map[string]interface{})

	if privileged, ok := in["privileged"].([]interface{}); ok && len(privileged) > 0 {
		ksc.Privileged = expandBoolValue(privileged)
	}

	if ro, ok := in["read_only_root_file_system"].([]interface{}); ok && len(ro) > 0 {
		ksc.ReadOnlyRootFileSystem = expandBoolValue(ro)
	}

	return &ksc
}

func expandKubeOptionsAffinity(p []interface{}) *corev1.Affinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	affinity := &corev1.Affinity{}

	if nodeAffinity, ok := in["node_affinity"].([]interface{}); ok && len(nodeAffinity) > 0 {
		affinity.NodeAffinity = expandNodeAffinity(nodeAffinity)
	}

	if podAffinity, ok := in["pod_affinity"].([]interface{}); ok && len(podAffinity) > 0 {
		affinity.PodAffinity = expandPodAffinity(podAffinity)
	}

	if podAntiAffinity, ok := in["pod_anti_affinity"].([]interface{}); ok && len(podAntiAffinity) > 0 {
		affinity.PodAntiAffinity = expandPodAntiAffinity(podAntiAffinity)
	}

	return affinity
}

func expandNodeSelector(p []interface{}) *corev1.NodeSelector {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	nodeSelector := &corev1.NodeSelector{}

	if nodeSelectorTerms, ok := in["node_selector_terms"].([]interface{}); ok && len(nodeSelectorTerms) > 0 {
		nodeSelector.NodeSelectorTerms = expandNodeSelectorTermsList(nodeSelectorTerms)
	}

	return nodeSelector
}

func expandNodeSelectorTermsList(p []interface{}) []corev1.NodeSelectorTerm {
	if len(p) == 0 {
		return nil
	}

	terms := make([]corev1.NodeSelectorTerm, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		term := corev1.NodeSelectorTerm{}

		if matchExpressions, ok := in["match_expressions"].([]interface{}); ok && len(matchExpressions) > 0 {
			term.MatchExpressions = expandNodeSelectorRequirements(matchExpressions)
		}

		if matchFields, ok := in["match_fields"].([]interface{}); ok && len(matchFields) > 0 {
			term.MatchFields = expandNodeSelectorRequirements(matchFields)
		}

		terms[i] = term
	}

	return terms
}

func expandNodeSelectorTerm(p []interface{}) corev1.NodeSelectorTerm {
	term := corev1.NodeSelectorTerm{}
	if len(p) == 0 {
		return term
	}

	data := p[0].(map[string]interface{})

	if matchExpressions, ok := data["match_expressions"].([]interface{}); ok && len(matchExpressions) > 0 {
		term.MatchExpressions = expandNodeSelectorRequirements(matchExpressions)
	}

	if matchFields, ok := data["match_fields"].([]interface{}); ok && len(matchFields) > 0 {
		term.MatchFields = expandNodeSelectorRequirements(matchFields)
	}

	return term
}

func expandNodeSelectorRequirements(p []interface{}) []corev1.NodeSelectorRequirement {
	if len(p) == 0 {
		return nil
	}

	requirements := make([]corev1.NodeSelectorRequirement, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		requirement := corev1.NodeSelectorRequirement{}

		if key, ok := in["key"].(string); ok && len(key) > 0 {
			requirement.Key = key
		}

		if operator, ok := in["operator"].(string); ok && len(operator) > 0 {
			requirement.Operator = corev1.NodeSelectorOperator(operator)
		}

		if values, ok := in["values"].([]interface{}); ok && len(values) > 0 {
			requirement.Values = toArrayString(values)
		}

		requirements[i] = requirement
	}

	return requirements
}

func expandPodAntiAffinity(p []interface{}) *corev1.PodAntiAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	podAntiAffinity := &corev1.PodAntiAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(required) > 0 {
		podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTerms(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(preferred) > 0 {
		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTerm(preferred)
	}

	return podAntiAffinity
}

func expandPodAffinityTerms(p []interface{}) []corev1.PodAffinityTerm {
	if len(p) == 0 {
		return nil
	}

	terms := make([]corev1.PodAffinityTerm, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		term := corev1.PodAffinityTerm{}

		if labelSelector, ok := in["label_selector"].([]interface{}); ok && len(labelSelector) > 0 {
			term.LabelSelector = expandLabelSelector(labelSelector)
		}

		if namespaces, ok := in["namespaces"].([]interface{}); ok && len(namespaces) > 0 {
			term.Namespaces = toArrayString(namespaces)
		}

		if topologyKey, ok := in["topology_key"].(string); ok && len(topologyKey) > 0 {
			term.TopologyKey = topologyKey
		}

		if namespaceSelector, ok := in["namespace_selector"].([]interface{}); ok && len(namespaceSelector) > 0 {
			term.NamespaceSelector = expandLabelSelector(namespaceSelector)
		}

		terms[i] = term
	}

	return terms
}

func expandPodAffinity(p []interface{}) *corev1.PodAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	podAffinity := &corev1.PodAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(required) > 0 {
		podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTerms(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(preferred) > 0 {
		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTerm(preferred)
	}

	return podAffinity
}

func expandWeightedPodAffinityTerm(preferred []interface{}) []corev1.WeightedPodAffinityTerm {
	if len(preferred) == 0 {
		return nil
	}

	terms := make([]corev1.WeightedPodAffinityTerm, len(preferred))
	for i := range preferred {
		in := preferred[i].(map[string]interface{})
		term := corev1.WeightedPodAffinityTerm{}

		if weight, ok := in["weight"].(int); ok {
			term.Weight = int32(weight)
		}

		if podAffinityTerm, ok := in["pod_affinity_term"].([]interface{}); ok && len(podAffinityTerm) > 0 {
			term.PodAffinityTerm = expandPodAffinityTerms(podAffinityTerm)[0]
		}

		terms[i] = term
	}

	return terms
}

func expandNodeAffinity(p []interface{}) *corev1.NodeAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	nodeAffinity := &corev1.NodeAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(required) > 0 {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandNodeSelector(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]interface{}); ok && len(preferred) > 0 {
		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandPreferredSchedulingTerms(preferred)
	}

	return nodeAffinity
}

func expandPreferredSchedulingTerms(preferred []interface{}) []corev1.PreferredSchedulingTerm {
	if len(preferred) == 0 {
		return nil
	}

	terms := make([]corev1.PreferredSchedulingTerm, len(preferred))
	for i := range preferred {
		in := preferred[i].(map[string]interface{})
		term := corev1.PreferredSchedulingTerm{}

		if weight, ok := in["weight"].(int); ok {
			term.Weight = int32(weight)
		}

		if preference, ok := in["preference"].([]interface{}); ok && len(preference) > 0 {
			term.Preference = expandNodeSelectorTerm(preference)
		}

		terms[i] = term
	}

	return terms
}

func expandWorkflowHandlerHttpConfig(p []interface{}) *eaaspb.HTTPDriverConfig {
	hc := eaaspb.HTTPDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]interface{})

	if body, ok := in["body"].(string); ok && len(body) > 0 {
		hc.Body = body
	}

	if endpoint, ok := in["endpoint"].(string); ok && len(endpoint) > 0 {
		hc.Endpoint = endpoint
	}

	if headers, ok := in["headers"].(map[string]interface{}); ok && len(headers) > 0 {
		hc.Headers = toMapString(headers)
	}

	if method, ok := in["method"].(string); ok && len(method) > 0 {
		hc.Method = method
	}

	return &hc
}

func expandWorkflowHandlerOutputs(p string) (*structpb.Struct, error) {
	if len(p) == 0 {
		return nil, nil
	}

	var s structpb.Struct
	if err := s.UnmarshalJSON([]byte(p)); err != nil {
		return nil, err
	}
	return &s, nil
}

// expandLabelSelector expands a label selector from a list of interfaces.
func expandLabelSelector(p []interface{}) *metav1.LabelSelector {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]interface{})
	labelSelector := &metav1.LabelSelector{}

	if matchLabels, ok := in["match_labels"].(map[string]interface{}); ok && len(matchLabels) > 0 {
		labelSelector.MatchLabels = toMapString(matchLabels)
	}

	if matchExpressions, ok := in["match_expressions"].([]interface{}); ok && len(matchExpressions) > 0 {
		labelSelector.MatchExpressions = expandLabelSelectorRequirements(matchExpressions)
	}

	return labelSelector
}

// expandLabelSelectorRequirements expands a list of label selector requirements.
func expandLabelSelectorRequirements(p []interface{}) []metav1.LabelSelectorRequirement {
	if len(p) == 0 {
		return nil
	}

	requirements := make([]metav1.LabelSelectorRequirement, len(p))
	for i := range p {
		in := p[i].(map[string]interface{})
		requirement := metav1.LabelSelectorRequirement{}

		if key, ok := in["key"].(string); ok && len(key) > 0 {
			requirement.Key = key
		}

		if operator, ok := in["operator"].(string); ok && len(operator) > 0 {
			requirement.Operator = metav1.LabelSelectorOperator(operator)
		}

		if values, ok := in["values"].([]interface{}); ok && len(values) > 0 {
			requirement.Values = toArrayString(values)
		}

		requirements[i] = requirement
	}

	return requirements
}

// flattenLabelSelectorRequirements flattens a list of label selector requirements.
func flattenLabelSelectorRequirements(in []metav1.LabelSelectorRequirement, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	requirements := make([]interface{}, len(in))
	for i, req := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["key"] = req.Key
		obj["operator"] = string(req.Operator)
		obj["values"] = toArrayInterface(req.Values)

		requirements[i] = obj
	}

	return requirements
}

// Flatteners

func flattenLabelSelector(in *metav1.LabelSelector, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.MatchLabels) > 0 {
		obj["match_labels"] = toMapInterface(in.MatchLabels)
	}

	if len(in.MatchExpressions) > 0 {
		v, ok := obj["match_expressions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["match_expressions"] = flattenLabelSelectorRequirements(in.MatchExpressions, v)
	}

	return []interface{}{obj}
}

func flattenNodeSelector(in *corev1.NodeSelector, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.NodeSelectorTerms) > 0 {
		v, ok := obj["node_selector_terms"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_selector_terms"] = flattenNodeSelectorTermsList(in.NodeSelectorTerms, v)
	}

	return []interface{}{obj}
}

func flattenNodeSelectorTermsList(in []corev1.NodeSelectorTerm, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	terms := make([]interface{}, len(in))
	for i, term := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(term.MatchExpressions) > 0 {
			v, ok := obj["match_expressions"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["match_expressions"] = flattenNodeSelectorRequirements(term.MatchExpressions, v)
		}

		if len(term.MatchFields) > 0 {
			v, ok := obj["match_fields"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["match_fields"] = flattenNodeSelectorRequirements(term.MatchFields, v)
		}

		terms[i] = obj
	}

	return terms
}

func flattenNodeSelectorRequirements(in []corev1.NodeSelectorRequirement, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	requirements := make([]interface{}, len(in))
	for i, req := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["key"] = req.Key
		obj["operator"] = string(req.Operator)
		obj["values"] = toArrayInterface(req.Values)

		requirements[i] = obj
	}

	return requirements
}

func flattenWorkflowHandler(d *schema.ResourceData, in *eaaspb.WorkflowHandler) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenV3MetaData(in.Metadata))
	if err != nil {
		log.Println("flatten metadata err")
		return err
	}

	v, ok := d.Get("spec").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenWorkflowHandlerSpec(in.Spec, v)
	if err != nil {
		log.Println("flatten workflow handler spec err")
		return err
	}

	err = d.Set("spec", ret)
	if err != nil {
		log.Println("set spec err")
		return err
	}
	return nil
}

func flattenPodAntiAffinity(in *corev1.PodAntiAffinity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		v, ok := obj["required_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	}

	if in.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		v, ok := obj["preferred_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	}

	return []interface{}{obj}
}

func flattenPodAffinity(in *corev1.PodAffinity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		v, ok := obj["required_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	}

	if in.PreferredDuringSchedulingIgnoredDuringExecution != nil {
		v, ok := obj["preferred_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	}

	return []interface{}{obj}
}

func flattenPodAffinityTerms(in []corev1.PodAffinityTerm, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	terms := make([]interface{}, len(in))
	for i, term := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if term.LabelSelector != nil {
			v, ok := obj["label_selector"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["label_selector"] = flattenLabelSelector(term.LabelSelector, v)
		}

		if len(term.Namespaces) > 0 {
			obj["namespaces"] = toArrayInterface(term.Namespaces)
		}

		if len(term.TopologyKey) > 0 {
			obj["topology_key"] = term.TopologyKey
		}

		if term.NamespaceSelector != nil {
			v, ok := obj["namespace_selector"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["namespace_selector"] = flattenLabelSelector(term.NamespaceSelector, v)
		}

		terms[i] = obj
	}

	return terms
}

func flattenWeightedPodAffinityTerms(in []corev1.WeightedPodAffinityTerm, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	terms := make([]interface{}, len(in))
	for i, term := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["weight"] = term.Weight
		obj["pod_affinity_term"] = flattenPodAffinityTerms([]corev1.PodAffinityTerm{term.PodAffinityTerm}, obj["pod_affinity_term"].([]interface{}))

		terms[i] = obj
	}

	return terms
}

func flattenPreferredSchedulingTerms(in []corev1.PreferredSchedulingTerm, p []interface{}) []interface{} {
	if len(in) == 0 {
		return nil
	}

	terms := make([]interface{}, len(in))
	for i, term := range in {
		obj := make(map[string]interface{})
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		obj["weight"] = term.Weight
		obj["preference"] = flattenNodeSelectorTerm(term.Preference, obj["preference"].([]interface{}))

		terms[i] = obj
	}

	return terms
}

func flattenNodeSelectorTerm(in corev1.NodeSelectorTerm, p []interface{}) []interface{} {
	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.MatchExpressions) > 0 {
		v, ok := obj["match_expressions"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["match_expressions"] = flattenNodeSelectorRequirements(in.MatchExpressions, v)
	}

	if len(in.MatchFields) > 0 {
		v, ok := obj["match_fields"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["match_fields"] = flattenNodeSelectorRequirements(in.MatchFields, v)
	}

	return []interface{}{obj}
}

func flattenWorkflowHandlerSpec(in *eaaspb.WorkflowHandlerSpec, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten workflow handler spec empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.Config != nil {
		v, ok := obj["config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["config"] = flattenWorkflowHandlerConfig(in.Config, v)
	}
	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["inputs"] = flattenConfigContextCompoundRefs(in.Inputs)
	obj["outputs"] = flattenWorkflowHandlerOutputs(in.Outputs)
	return []interface{}{obj}, nil
}

func flattenWorkflowHandlerConfig(input *eaaspb.WorkflowHandlerConfig, p []interface{}) []interface{} {
	log.Println("flatten workflow handler config start", input)
	if input == nil {
		return nil
	}

	obj := map[string]interface{}{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(input.Type) > 0 {
		obj["type"] = input.Type
	}

	obj["timeout_seconds"] = input.TimeoutSeconds

	if len(input.SuccessCondition) > 0 {
		obj["success_condition"] = input.SuccessCondition
	}

	obj["max_retry_count"] = input.MaxRetryCount

	if input.Container != nil {
		v, ok := obj["container"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["container"] = flattenWorkflowHandlerContainerConfig(input.Container, v)
	}

	if input.Http != nil {
		v, ok := obj["http"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["http"] = flattenWorkflowHandlerHttpConfig(input.Http, v)
	}

	if input.PollingConfig != nil {
		v, ok := obj["polling_config"].([]interface{})
		if !ok {
			v = []interface{}{}
		}

		obj["polling_config"] = flattenPollingConfig(input.PollingConfig, v)
	}

	return []interface{}{obj}
}

func flattenWorkflowHandlerContainerConfig(in *eaaspb.ContainerDriverConfig, p []interface{}) []interface{} {
	log.Println("flatten container workflow handler config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["arguments"] = toArrayInterface(in.Arguments)
	obj["commands"] = toArrayInterface(in.Commands)
	obj["cpu_limit_milli"] = in.CpuLimitMilli
	obj["env_vars"] = toMapInterface(in.EnvVars)
	obj["files"] = toMapByteInterface(in.Files)
	obj["image"] = in.Image
	obj["image_pull_credentials"] = flattenImagePullCredentials(in.ImagePullCredentials, obj["image_pull_credentials"].([]any))
	obj["kube_config_options"] = flattenContainerKubeConfig(in.KubeConfigOptions, obj["kube_config_options"].([]any))
	obj["kube_options"] = flattenContainerKubeOptions(in.KubeOptions, obj["kube_options"].([]interface{}))
	obj["memory_limit_mb"] = in.MemoryLimitMb
	obj["volume_options"] = flattenContainerWorkflowHandlerVolumeOptions(
		[]*eaaspb.ContainerDriverVolumeOptions{in.VolumeOptions}, obj["volume_options"].([]interface{}),
	)
	obj["volumes"] = flattenContainerWorkflowHandlerVolumeOptions(in.Volumes, obj["volumes"].([]interface{}))
	obj["working_dir_path"] = in.WorkingDirPath
	return []interface{}{obj}
}

func flattenImagePullCredentials(in *eaaspb.ContainerImagePullCredentials, p []interface{}) []interface{} {
	log.Println("flatten container image pull credentials start")
	if in == nil {
		return nil
	}
	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["registry"] = in.Registry
	obj["username"] = in.Username
	obj["password"] = in.Password
	return []interface{}{obj}
}

func flattenContainerKubeConfig(in *eaaspb.ContainerKubeConfigOptions, p []interface{}) []interface{} {
	log.Println("flatten container kube config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["kube_config"] = in.KubeConfig
	obj["out_of_cluster"] = in.OutOfCluster
	return []interface{}{obj}
}

func flattenContainerKubeOptions(in *eaaspb.ContainerKubeOptions, p []interface{}) []interface{} {
	log.Println("flatten container kube options start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["labels"] = toMapInterface(in.Labels)
	obj["namespace"] = in.Namespace
	obj["node_selector"] = toMapInterface(in.NodeSelector)
	obj["resources"] = toArrayInterface(in.Resources)
	obj["security_context"] = flattenSecurityContext(in.SecurityContext, obj["security_context"].([]interface{}))
	obj["service_account_name"] = in.ServiceAccountName
	if len(in.Tolerations) > 0 {
		obj["tolerations"] = flattenV3Tolerations(in.Tolerations, obj["tolerations"].([]interface{}))
	} else {
		delete(obj, "tolerations")
	}
	obj["affinity"] = flattenKubeOptionsAffinity(in.Affinity, obj["affinity"].([]interface{}))
	return []interface{}{obj}
}

func flattenSecurityContext(in *eaaspb.KubeSecurityContext, p []interface{}) []interface{} {
	log.Println("flatten kube security context start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["privileged"] = flattenBoolValue(in.Privileged)
	obj["read_only_root_file_system"] = flattenBoolValue(in.ReadOnlyRootFileSystem)

	return []interface{}{obj}
}

func flattenKubeOptionsAffinity(in *corev1.Affinity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.NodeAffinity != nil {
		v, ok := obj["node_affinity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["node_affinity"] = flattenNodeAffinity(in.NodeAffinity, v)
	}

	if in.PodAffinity != nil {
		v, ok := obj["pod_affinity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_affinity"] = flattenPodAffinity(in.PodAffinity, v)
	}

	if in.PodAntiAffinity != nil {
		v, ok := obj["pod_anti_affinity"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["pod_anti_affinity"] = flattenPodAntiAffinity(in.PodAntiAffinity, v)
	}

	return []interface{}{obj}
}

func flattenNodeAffinity(in *corev1.NodeAffinity, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if in.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		v, ok := obj["required_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["required_during_scheduling_ignored_during_execution"] = flattenNodeSelector(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	}

	if len(in.PreferredDuringSchedulingIgnoredDuringExecution) > 0 {
		v, ok := obj["preferred_during_scheduling_ignored_during_execution"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["preferred_during_scheduling_ignored_during_execution"] = flattenPreferredSchedulingTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	}

	return []interface{}{obj}
}

func flattenWorkflowHandlerHttpConfig(in *eaaspb.HTTPDriverConfig, p []interface{}) []interface{} {
	log.Println("flatten http config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]interface{})
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	obj["body"] = in.Body
	obj["endpoint"] = in.Endpoint
	obj["headers"] = toMapInterface(in.Headers)
	obj["method"] = in.Method
	return []interface{}{obj}
}

func expandContainerWorkflowHandlerVolumeOptions(p []interface{}) []*eaaspb.ContainerDriverVolumeOptions {
	volumes := make([]*eaaspb.ContainerDriverVolumeOptions, 0)
	if len(p) == 0 {
		return volumes
	}

	for indx := range p {
		volume := &eaaspb.ContainerDriverVolumeOptions{}
		if p[indx] == nil {
			return volumes
		}
		in := p[indx].(map[string]interface{})

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			volume.MountPath = mp
		}

		if pvcsz, ok := in["pvc_size_gb"].(string); ok && len(pvcsz) > 0 {
			volume.PvcSizeGB = pvcsz
		}

		if pvcsc, ok := in["pvc_storage_class"].(string); ok && len(pvcsc) > 0 {
			volume.PvcStorageClass = pvcsc
		}

		if usepvc, ok := in["use_pvc"].([]interface{}); ok && len(usepvc) > 0 {
			volume.UsePVC = expandBoolValue(usepvc)
		}

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
		}

		volumes = append(volumes, volume)

	}

	return volumes
}

func flattenContainerWorkflowHandlerVolumeOptions(input []*eaaspb.ContainerDriverVolumeOptions, p []interface{}) []interface{} {
	if len(input) == 0 {
		return nil
	}

	var out []interface{}
	for i, in := range input {
		if in == nil {
			continue
		}
		log.Println("flatten container workflow handler volume options", in)
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}
		obj["use_pvc"] = flattenBoolValue(in.UsePVC)
		obj["mount_path"] = in.MountPath
		obj["pvc_size_gb"] = in.PvcSizeGB
		obj["pvc_storage_class"] = in.PvcStorageClass
		obj["enable_backup_and_restore"] = in.EnableBackupAndRestore
		out = append(out, &obj)
	}

	if len(out) == 0 {
		return nil
	}
	return out
}

func flattenWorkflowHandlerOutputs(in *structpb.Struct) string {
	if in == nil {
		return ""
	}
	b, _ := in.MarshalJSON()
	return string(b)
}

func resourceWorkflowHandlerImport(_ context.Context, d *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
	log.Printf("WorkflowHandler Import Starts")

	idParts := strings.SplitN(d.Id(), "/", 2)
	log.Println("resourceWorkflowHandlerImport idParts:", idParts)

	log.Println("resourceWorkflowHandlerImport Invoking expandWorkflowHandler")
	cc, err := expandWorkflowHandler(d)
	if err != nil {
		log.Printf("resourceWorkflowHandlerImport  expand error %s", err.Error())
	}

	var metaD commonpb.Metadata
	metaD.Name = idParts[0]
	metaD.Project = idParts[1]
	cc.Metadata = &metaD

	err = d.Set("metadata", flattenV3MetaData(&metaD))
	if err != nil {
		log.Println("import set metadata err ", err)
		return nil, err
	}
	d.SetId(cc.Metadata.Name)
	return []*schema.ResourceData{d}, nil
}
