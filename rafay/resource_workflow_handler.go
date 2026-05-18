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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
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

func resourceWorkflowHandlerCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceWorkflowHandlerUpsert(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

func resourceWorkflowHandlerRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if cc.GetSpec().GetSharing() != nil && !cc.GetSpec().GetSharing().GetEnabled() && wh.GetSpec().GetSharing() == nil {
		wh.Spec.Sharing = &commonpb.SharingSpec{}
		wh.Spec.Sharing.Enabled = false
		wh.Spec.Sharing.Projects = cc.GetSpec().GetSharing().GetProjects()
	}

	err = flattenWorkflowHandler(d, wh)
	if err != nil {
		log.Println("read flatten err")
		return diag.FromErr(err)
	}
	return diags

}

func resourceWorkflowHandlerUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	return resourceWorkflowHandlerUpsert(ctx, d, m)
}

func resourceWorkflowHandlerDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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

	if v, ok := in.Get("metadata").([]any); ok && len(v) > 0 {
		obj.Metadata = expandV3MetaData(v)
	}

	if v, ok := in.Get("spec").([]any); ok && len(v) > 0 {
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

func expandWorkflowHandlerSpec(p []any) (*eaaspb.WorkflowHandlerSpec, error) {
	log.Println("expand workflow handler spec")
	spec := &eaaspb.WorkflowHandlerSpec{}
	if len(p) == 0 || p[0] == nil {
		return spec, fmt.Errorf("%s", "expand workflow handler spec empty input")
	}

	in := p[0].(map[string]any)

	if c, ok := in["config"].([]any); ok && len(c) > 0 {
		spec.Config = expandWorkflowHandlerConfig(c)
	}

	if v, ok := in["sharing"].([]any); ok && len(v) > 0 {
		spec.Sharing = expandSharingSpec(v)
	}

	if v, ok := in["inputs"].([]any); ok && len(v) > 0 {
		spec.Inputs = expandConfigContextCompoundRefs(v)
	}

	if v, ok := in["icon_url"].(string); ok {
		spec.IconURL = v
	}

	if v, ok := in["readme"].(string); ok {
		spec.Readme = v
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

func expandWorkflowHandlerConfig(p []any) *eaaspb.WorkflowHandlerConfig {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	workflowHandlerConfig := eaaspb.WorkflowHandlerConfig{}
	in := p[0].(map[string]any)

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

	if v, ok := in["container"].([]any); ok && len(v) > 0 {
		workflowHandlerConfig.Container = expandWorkflowHandlerContainerConfig(v)
	}

	if v, ok := in["http"].([]any); ok && len(v) > 0 {
		workflowHandlerConfig.Http = expandWorkflowHandlerHttpConfig(v)
	}

	if v, ok := in["function"].([]any); ok && len(v) > 0 {
		workflowHandlerConfig.Function = expandWorkflowHandlerFunctionConfig(v)
	}

	if v, ok := in["polling_config"].([]any); ok && len(v) > 0 {
		workflowHandlerConfig.PollingConfig = expandPollingConfig(v)
	}

	if h, ok := in["timeout_seconds"].(int); ok {
		workflowHandlerConfig.TimeoutSeconds = int64(h)
	}

	return &workflowHandlerConfig
}

func expandWorkflowHandlerContainerConfig(p []any) *eaaspb.ContainerDriverConfig {
	cc := eaaspb.ContainerDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &cc
	}

	in := p[0].(map[string]any)

	if img, ok := in["image"].(string); ok && len(img) > 0 {
		cc.Image = img
	}

	if args, ok := in["arguments"].([]any); ok && len(args) > 0 {
		cc.Arguments = toArrayString(args)
	}

	if cmds, ok := in["commands"].([]any); ok && len(cmds) > 0 {
		cc.Commands = toArrayString(cmds)
	}

	if clm, ok := in["cpu_limit_milli"].(string); ok && len(clm) > 0 {
		cc.CpuLimitMilli = clm
	}

	if ev, ok := in["env_vars"].(map[string]any); ok && len(ev) > 0 {
		cc.EnvVars = toMapString(ev)
	}

	if f, ok := in["files"].(map[string]any); ok && len(f) > 0 {
		cc.Files = toMapByte(f)
	}

	if v, ok := in["image_pull_credentials"].([]any); ok && len(v) > 0 {
		cc.ImagePullCredentials = expandImagePullCredentials(v)
	}

	if v, ok := in["kube_config_options"].([]any); ok && len(v) > 0 {
		cc.KubeConfigOptions = expandKubeConfigOptions(v)
	}

	if v, ok := in["kube_options"].([]any); ok && len(v) > 0 {
		cc.KubeOptions = expandContainerKubeOptions(v)
	}

	if mlb, ok := in["memory_limit_mb"].(string); ok && len(mlb) > 0 {
		cc.MemoryLimitMb = mlb
	}

	if v, ok := in["volume_options"].([]any); ok && len(v) > 0 {
		volumes := expandContainerWorkflowHandlerVolumeOptions(v)
		if len(volumes) > 0 {
			cc.VolumeOptions = volumes[0]
		}
	}

	if v, ok := in["volumes"].([]any); ok && len(v) > 0 {
		cc.Volumes = expandContainerWorkflowHandlerVolumeOptions(v)
	}

	if wdp, ok := in["working_dir_path"].(string); ok && len(wdp) > 0 {
		cc.WorkingDirPath = wdp
	}

	return &cc
}

func expandImagePullCredentials(p []any) *eaaspb.ContainerImagePullCredentials {
	hc := eaaspb.ContainerImagePullCredentials{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]any)

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

func expandKubeConfigOptions(p []any) *eaaspb.ContainerKubeConfigOptions {
	hc := eaaspb.ContainerKubeConfigOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]any)

	if kc, ok := in["kube_config"].(string); ok && len(kc) > 0 {
		hc.KubeConfig = kc
	}

	if ofc, ok := in["out_of_cluster"].(bool); ok {
		hc.OutOfCluster = ofc
	}

	return &hc
}

func expandContainerKubeOptions(p []any) *eaaspb.ContainerKubeOptions {
	hc := eaaspb.ContainerKubeOptions{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]any)

	if lbls, ok := in["labels"].(map[string]any); ok && len(lbls) > 0 {
		hc.Labels = toMapString(lbls)
	}

	if ns, ok := in["namespace"].(string); ok && len(ns) > 0 {
		hc.Namespace = ns
	}

	if ns, ok := in["node_selector"].(map[string]any); ok && len(ns) > 0 {
		hc.NodeSelector = toMapString(ns)
	}

	if r, ok := in["resources"].([]any); ok && len(r) > 0 {
		hc.Resources = toArrayString(r)
	}

	if sc, ok := in["security_context"].([]any); ok && len(sc) > 0 {
		hc.SecurityContext = expandSecurityContext(sc)
	}

	if san, ok := in["service_account_name"].(string); ok && len(san) > 0 {
		hc.ServiceAccountName = san
	}

	if tolerations, ok := in["tolerations"].([]any); ok {
		hc.Tolerations = expandV3Tolerations(tolerations)
	}

	if a, ok := in["affinity"].([]any); ok && len(a) > 0 {
		hc.Affinity = expandKubeOptionsAffinity(a)
	}

	return &hc
}

func expandSecurityContext(p []any) *eaaspb.KubeSecurityContext {
	ksc := eaaspb.KubeSecurityContext{}
	if len(p) == 0 || p[0] == nil {
		return &ksc
	}

	in := p[0].(map[string]any)

	if privileged, ok := in["privileged"].([]any); ok && len(privileged) > 0 {
		ksc.Privileged = expandBoolValue(privileged)
	}

	if ro, ok := in["read_only_root_file_system"].([]any); ok && len(ro) > 0 {
		ksc.ReadOnlyRootFileSystem = expandBoolValue(ro)
	}

	return &ksc
}

func expandKubeOptionsAffinity(p []any) *corev1.Affinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	affinity := &corev1.Affinity{}

	if nodeAffinity, ok := in["node_affinity"].([]any); ok && len(nodeAffinity) > 0 {
		affinity.NodeAffinity = expandNodeAffinity(nodeAffinity)
	}

	if podAffinity, ok := in["pod_affinity"].([]any); ok && len(podAffinity) > 0 {
		affinity.PodAffinity = expandPodAffinity(podAffinity)
	}

	if podAntiAffinity, ok := in["pod_anti_affinity"].([]any); ok && len(podAntiAffinity) > 0 {
		affinity.PodAntiAffinity = expandPodAntiAffinity(podAntiAffinity)
	}

	return affinity
}

func expandNodeSelector(p []any) *corev1.NodeSelector {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	nodeSelector := &corev1.NodeSelector{}

	if nodeSelectorTerms, ok := in["node_selector_terms"].([]any); ok && len(nodeSelectorTerms) > 0 {
		nodeSelector.NodeSelectorTerms = expandNodeSelectorTermsList(nodeSelectorTerms)
	}

	return nodeSelector
}

func expandNodeSelectorTermsList(p []any) []corev1.NodeSelectorTerm {
	if len(p) == 0 {
		return nil
	}

	terms := make([]corev1.NodeSelectorTerm, len(p))
	for i := range p {
		in := p[i].(map[string]any)
		term := corev1.NodeSelectorTerm{}

		if matchExpressions, ok := in["match_expressions"].([]any); ok && len(matchExpressions) > 0 {
			term.MatchExpressions = expandNodeSelectorRequirements(matchExpressions)
		}

		if matchFields, ok := in["match_fields"].([]any); ok && len(matchFields) > 0 {
			term.MatchFields = expandNodeSelectorRequirements(matchFields)
		}

		terms[i] = term
	}

	return terms
}

func expandNodeSelectorTerm(p []any) corev1.NodeSelectorTerm {
	term := corev1.NodeSelectorTerm{}
	if len(p) == 0 {
		return term
	}

	data := p[0].(map[string]any)

	if matchExpressions, ok := data["match_expressions"].([]any); ok && len(matchExpressions) > 0 {
		term.MatchExpressions = expandNodeSelectorRequirements(matchExpressions)
	}

	if matchFields, ok := data["match_fields"].([]any); ok && len(matchFields) > 0 {
		term.MatchFields = expandNodeSelectorRequirements(matchFields)
	}

	return term
}

func expandNodeSelectorRequirements(p []any) []corev1.NodeSelectorRequirement {
	if len(p) == 0 {
		return nil
	}

	requirements := make([]corev1.NodeSelectorRequirement, len(p))
	for i := range p {
		in := p[i].(map[string]any)
		requirement := corev1.NodeSelectorRequirement{}

		if key, ok := in["key"].(string); ok && len(key) > 0 {
			requirement.Key = key
		}

		if operator, ok := in["operator"].(string); ok && len(operator) > 0 {
			requirement.Operator = corev1.NodeSelectorOperator(operator)
		}

		if values, ok := in["values"].([]any); ok && len(values) > 0 {
			requirement.Values = toArrayString(values)
		}

		requirements[i] = requirement
	}

	return requirements
}

func expandPodAntiAffinity(p []any) *corev1.PodAntiAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	podAntiAffinity := &corev1.PodAntiAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]any); ok && len(required) > 0 {
		podAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTerms(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]any); ok && len(preferred) > 0 {
		podAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTerm(preferred)
	}

	return podAntiAffinity
}

func expandPodAffinityTerms(p []any) []corev1.PodAffinityTerm {
	if len(p) == 0 {
		return nil
	}

	terms := make([]corev1.PodAffinityTerm, len(p))
	for i := range p {
		in := p[i].(map[string]any)
		term := corev1.PodAffinityTerm{}

		if labelSelector, ok := in["label_selector"].([]any); ok && len(labelSelector) > 0 {
			term.LabelSelector = expandLabelSelector(labelSelector)
		}

		if namespaces, ok := in["namespaces"].([]any); ok && len(namespaces) > 0 {
			term.Namespaces = toArrayString(namespaces)
		}

		if topologyKey, ok := in["topology_key"].(string); ok && len(topologyKey) > 0 {
			term.TopologyKey = topologyKey
		}

		if namespaceSelector, ok := in["namespace_selector"].([]any); ok && len(namespaceSelector) > 0 {
			term.NamespaceSelector = expandLabelSelector(namespaceSelector)
		}

		terms[i] = term
	}

	return terms
}

func expandPodAffinity(p []any) *corev1.PodAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	podAffinity := &corev1.PodAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]any); ok && len(required) > 0 {
		podAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandPodAffinityTerms(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]any); ok && len(preferred) > 0 {
		podAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandWeightedPodAffinityTerm(preferred)
	}

	return podAffinity
}

func expandWeightedPodAffinityTerm(preferred []any) []corev1.WeightedPodAffinityTerm {
	if len(preferred) == 0 {
		return nil
	}

	terms := make([]corev1.WeightedPodAffinityTerm, len(preferred))
	for i := range preferred {
		in := preferred[i].(map[string]any)
		term := corev1.WeightedPodAffinityTerm{}

		if weight, ok := in["weight"].(int); ok {
			term.Weight = int32(weight)
		}

		if podAffinityTerm, ok := in["pod_affinity_term"].([]any); ok && len(podAffinityTerm) > 0 {
			term.PodAffinityTerm = expandPodAffinityTerms(podAffinityTerm)[0]
		}

		terms[i] = term
	}

	return terms
}

func expandNodeAffinity(p []any) *corev1.NodeAffinity {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	nodeAffinity := &corev1.NodeAffinity{}

	if required, ok := in["required_during_scheduling_ignored_during_execution"].([]any); ok && len(required) > 0 {
		nodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = expandNodeSelector(required)
	}

	if preferred, ok := in["preferred_during_scheduling_ignored_during_execution"].([]any); ok && len(preferred) > 0 {
		nodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = expandPreferredSchedulingTerms(preferred)
	}

	return nodeAffinity
}

func expandPreferredSchedulingTerms(preferred []any) []corev1.PreferredSchedulingTerm {
	if len(preferred) == 0 {
		return nil
	}

	terms := make([]corev1.PreferredSchedulingTerm, len(preferred))
	for i := range preferred {
		in := preferred[i].(map[string]any)
		term := corev1.PreferredSchedulingTerm{}

		if weight, ok := in["weight"].(int); ok {
			term.Weight = int32(weight)
		}

		if preference, ok := in["preference"].([]any); ok && len(preference) > 0 {
			term.Preference = expandNodeSelectorTerm(preference)
		}

		terms[i] = term
	}

	return terms
}

func expandWorkflowHandlerHttpConfig(p []any) *eaaspb.HTTPDriverConfig {
	hc := eaaspb.HTTPDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &hc
	}

	in := p[0].(map[string]any)

	if body, ok := in["body"].(string); ok && len(body) > 0 {
		hc.Body = body
	}

	if endpoint, ok := in["endpoint"].(string); ok && len(endpoint) > 0 {
		hc.Endpoint = endpoint
	}

	if headers, ok := in["headers"].(map[string]any); ok && len(headers) > 0 {
		hc.Headers = toMapString(headers)
	}

	if method, ok := in["method"].(string); ok && len(method) > 0 {
		hc.Method = method
	}

	return &hc
}

func expandWorkflowHandlerFunctionConfig(p []any) *eaaspb.FunctionDriverConfig {
	fdc := eaaspb.FunctionDriverConfig{}
	if len(p) == 0 || p[0] == nil {
		return &fdc
	}
	in := p[0].(map[string]any)

	if name, ok := in["name"].(string); ok && len(name) > 0 {
		fdc.Name = name
	}

	if language, ok := in["language"].(string); ok && len(language) > 0 {
		fdc.Language = language
	}

	if source, ok := in["source"].(string); ok && len(source) > 0 {
		fdc.Source = source
	}

	if fd, ok := in["function_dependencies"].([]any); ok && len(fd) > 0 {
		fdc.FunctionDependencies = toArrayString(fd)
	}

	if systemPackages, ok := in["system_packages"].([]any); ok && len(systemPackages) > 0 {
		fdc.SystemPackages = toArrayString(systemPackages)
	}

	if targetPlatforms, ok := in["target_platforms"].([]any); ok && len(targetPlatforms) > 0 {
		fdc.TargetPlatforms = toArrayString(targetPlatforms)
	}

	if languageVersion, ok := in["language_version"].(string); ok && len(languageVersion) > 0 {
		fdc.LanguageVersion = languageVersion
	}

	if cpuLimitMilli, ok := in["cpu_limit_milli"].(string); ok && len(cpuLimitMilli) > 0 {
		fdc.CpuLimitMilli = cpuLimitMilli
	}

	if memoryLimitMb, ok := in["memory_limit_mb"].(string); ok && len(memoryLimitMb) > 0 {
		fdc.MemoryLimitMb = memoryLimitMb
	}

	if skipBuild, ok := in["skip_build"].([]any); ok {
		fdc.SkipBuild = expandBoolValue(skipBuild)
	}

	if image, ok := in["image"].(string); ok && len(image) > 0 {
		fdc.Image = image
	}

	if maxConcurrency, ok := in["max_concurrency"].(int); ok {
		fdc.MaxConcurrency = int64(maxConcurrency)
	}

	if numReplicas, ok := in["num_replicas"].(int); ok {
		fdc.NumReplicas = uint32(numReplicas)
	}

	if kubeOptions, ok := in["kube_options"].([]any); ok && len(kubeOptions) > 0 {
		fdc.KubeOptions = expandContainerKubeOptions(kubeOptions)
	}

	if imagePullCredentials, ok := in["image_pull_credentials"].([]any); ok && len(imagePullCredentials) > 0 {
		fdc.ImagePullCredentials = expandImagePullCredentials(imagePullCredentials)
	}

	if inactivityTimeoutSeconds, ok := in["inactivity_timeout_seconds"].(int); ok {
		fdc.InactivityTimeoutSeconds = int64(inactivityTimeoutSeconds)
	}

	if resources, ok := in["resources"].([]any); ok && len(resources) > 0 {
		fdc.Resources = expandDriverResources(resources)
	}

	if hpa, ok := in["hpa"].([]any); ok && len(hpa) > 0 {
		fdc.Hpa = expandFunctionHPAConfig(hpa)
	}

	return &fdc
}

func expandDriverResourceValues(p []any) *eaaspb.DriverResourceValues {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	in := p[0].(map[string]any)
	drv := &eaaspb.DriverResourceValues{}
	if cpu, ok := in["cpu"].(string); ok && len(cpu) > 0 {
		drv.Cpu = cpu
	}
	if memory, ok := in["memory"].(string); ok && len(memory) > 0 {
		drv.Memory = memory
	}
	return drv
}

func expandDriverResources(p []any) *eaaspb.DriverResources {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	in := p[0].(map[string]any)
	dr := &eaaspb.DriverResources{}
	if requests, ok := in["requests"].([]any); ok && len(requests) > 0 {
		dr.Requests = expandDriverResourceValues(requests)
	}
	if limits, ok := in["limits"].([]any); ok && len(limits) > 0 {
		dr.Limits = expandDriverResourceValues(limits)
	}
	if unset, ok := in["unset"].([]any); ok && len(unset) > 0 {
		dr.Unset = expandBoolValue(unset)
	}
	return dr
}

func expandFunctionHPAConfig(p []any) *eaaspb.FunctionHPAConfig {
	if len(p) == 0 || p[0] == nil {
		return nil
	}
	in := p[0].(map[string]any)
	hpa := &eaaspb.FunctionHPAConfig{}
	if enabled, ok := in["enabled"].([]any); ok && len(enabled) > 0 {
		hpa.Enabled = expandBoolValue(enabled)
	}
	if minReplicas, ok := in["min_replicas"].(int); ok {
		hpa.MinReplicas = uint32(minReplicas)
	}
	if maxReplicas, ok := in["max_replicas"].(int); ok {
		hpa.MaxReplicas = uint32(maxReplicas)
	}
	if resourceMetrics, ok := in["resource_metrics"].([]any); ok && len(resourceMetrics) > 0 {
		hpa.ResourceMetrics = expandResourceMetrics(resourceMetrics)
	}
	return hpa
}

func expandMetricTarget(in map[string]any) *autoscalingv2.MetricTarget {
	if in == nil {
		return nil
	}
	mt := &autoscalingv2.MetricTarget{}
	if t, ok := in["type"].(string); ok && len(t) > 0 {
		mt.Type = autoscalingv2.MetricTargetType(t)
	}
	if v, ok := in["value"].(string); ok && len(v) > 0 {
		q, err := k8sresource.ParseQuantity(v)
		if err == nil {
			mt.Value = &q
		}
	}
	if av, ok := in["average_value"].(string); ok && len(av) > 0 {
		q, err := k8sresource.ParseQuantity(av)
		if err == nil {
			mt.AverageValue = &q
		}
	}
	if au, ok := in["average_utilization"].(int); ok {
		au32 := int32(au)
		mt.AverageUtilization = &au32
	}
	return mt
}

func expandResourceMetricSource(in map[string]any) *autoscalingv2.ResourceMetricSource {
	if in == nil {
		return nil
	}
	rms := &autoscalingv2.ResourceMetricSource{}
	if name, ok := in["name"].(string); ok && len(name) > 0 {
		rms.Name = corev1.ResourceName(name)
	}
	if target, ok := in["target"].([]any); ok && len(target) > 0 {
		if tmap, ok := target[0].(map[string]any); ok {
			if mt := expandMetricTarget(tmap); mt != nil {
				rms.Target = *mt
			}
		}
	}
	return rms
}

func expandResourceMetrics(p []any) []*autoscalingv2.ResourceMetricSource {
	if len(p) == 0 {
		return nil
	}
	out := make([]*autoscalingv2.ResourceMetricSource, 0, len(p))
	for i := range p {
		if p[i] == nil {
			continue
		}
		if m, ok := p[i].(map[string]any); ok {
			if rms := expandResourceMetricSource(m); rms != nil {
				out = append(out, rms)
			}
		}
	}
	return out
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
func expandLabelSelector(p []any) *metav1.LabelSelector {
	if len(p) == 0 || p[0] == nil {
		return nil
	}

	in := p[0].(map[string]any)
	labelSelector := &metav1.LabelSelector{}

	if matchLabels, ok := in["match_labels"].(map[string]any); ok && len(matchLabels) > 0 {
		labelSelector.MatchLabels = toMapString(matchLabels)
	}

	if matchExpressions, ok := in["match_expressions"].([]any); ok && len(matchExpressions) > 0 {
		labelSelector.MatchExpressions = expandLabelSelectorRequirements(matchExpressions)
	}

	return labelSelector
}

// expandLabelSelectorRequirements expands a list of label selector requirements.
func expandLabelSelectorRequirements(p []any) []metav1.LabelSelectorRequirement {
	if len(p) == 0 {
		return nil
	}

	requirements := make([]metav1.LabelSelectorRequirement, len(p))
	for i := range p {
		in := p[i].(map[string]any)
		requirement := metav1.LabelSelectorRequirement{}

		if key, ok := in["key"].(string); ok && len(key) > 0 {
			requirement.Key = key
		}

		if operator, ok := in["operator"].(string); ok && len(operator) > 0 {
			requirement.Operator = metav1.LabelSelectorOperator(operator)
		}

		if values, ok := in["values"].([]any); ok && len(values) > 0 {
			requirement.Values = toArrayString(values)
		}

		requirements[i] = requirement
	}

	return requirements
}

// flattenLabelSelectorRequirements flattens a list of label selector requirements.
func flattenLabelSelectorRequirements(in []metav1.LabelSelectorRequirement, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	requirements := make([]any, len(in))
	for i, req := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}

		obj["key"] = req.Key
		obj["operator"] = string(req.Operator)
		obj["values"] = toArrayInterface(req.Values)

		requirements[i] = obj
	}

	return requirements
}

func expandWorkflowHandlerCompoundRef(p []any) (*eaaspb.WorkflowHandlerCompoundRef, error) {
	wfHandler := &eaaspb.WorkflowHandlerCompoundRef{}
	if len(p) == 0 || p[0] == nil {
		return wfHandler, nil
	}
	in := p[0].(map[string]any)

	if v, ok := in["name"].(string); ok && len(v) > 0 {
		wfHandler.Name = v
	}

	var err error
	if v, ok := in["data"].([]any); ok && len(v) > 0 {
		wfHandler.Data, err = expandWorkflowHandlerInline(v)
		if err != nil {
			return nil, err
		}
	}

	return wfHandler, nil
}

func expandWorkflowHandlerInline(p []any) (*eaaspb.WorkflowHandlerInline, error) {
	wfHandlerInline := &eaaspb.WorkflowHandlerInline{}
	if len(p) == 0 || p[0] == nil {
		return wfHandlerInline, nil
	}

	in := p[0].(map[string]any)

	if v, ok := in["config"].([]any); ok && len(v) > 0 {
		wfHandlerInline.Config = expandWorkflowHandlerConfig(v)
	}

	if v, ok := in["inputs"].([]any); ok && len(v) > 0 {
		wfHandlerInline.Inputs = expandConfigContextCompoundRefs(v)
	}

	if v, ok := in["outputs"].(string); ok && len(v) > 0 {
		outputs, err := expandWorkflowHandlerOutputs(v)
		if err != nil {
			return nil, err
		}
		wfHandlerInline.Outputs = outputs
	}

	return wfHandlerInline, nil
}

func expandPollingConfig(p []any) *eaaspb.PollingConfig {
	pc := &eaaspb.PollingConfig{}
	if len(p) == 0 || p[0] == nil {
		return pc
	}

	in := p[0].(map[string]any)

	if h, ok := in["repeat"].(string); ok {
		pc.Repeat = h
	}

	if h, ok := in["until"].(string); ok {
		pc.Until = h
	}

	return pc
}

func expandContainerWorkflowHandlerVolumeOptions(p []any) []*eaaspb.ContainerDriverVolumeOptions {
	volumes := make([]*eaaspb.ContainerDriverVolumeOptions, 0)
	if len(p) == 0 {
		return volumes
	}

	for indx := range p {
		volume := &eaaspb.ContainerDriverVolumeOptions{}
		if p[indx] == nil {
			return volumes
		}
		in := p[indx].(map[string]any)

		if mp, ok := in["mount_path"].(string); ok && len(mp) > 0 {
			volume.MountPath = mp
		}

		if pvcsz, ok := in["pvc_size_gb"].(string); ok && len(pvcsz) > 0 {
			volume.PvcSizeGB = pvcsz
		}

		if pvcsc, ok := in["pvc_storage_class"].(string); ok && len(pvcsc) > 0 {
			volume.PvcStorageClass = pvcsc
		}

		if usepvc, ok := in["use_pvc"].([]any); ok && len(usepvc) > 0 {
			volume.UsePVC = expandBoolValue(usepvc)
		}

		if enableBackupAndRestore, ok := in["enable_backup_and_restore"].(bool); ok {
			volume.EnableBackupAndRestore = enableBackupAndRestore
		}

		volumes = append(volumes, volume)

	}

	return volumes
}

// Flatteners

func flattenLabelSelector(in *metav1.LabelSelector, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["match_labels"] = toMapInterface(in.MatchLabels)
	v, _ := obj["match_expressions"].([]any)
	obj["match_expressions"] = flattenLabelSelectorRequirements(in.MatchExpressions, v)
	return []any{obj}
}

func flattenNodeSelector(in *corev1.NodeSelector, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["node_selector_terms"].([]any)
	obj["node_selector_terms"] = flattenNodeSelectorTermsList(in.NodeSelectorTerms, v)
	return []any{obj}
}

func flattenNodeSelectorTermsList(in []corev1.NodeSelectorTerm, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	terms := make([]any, len(in))
	for i, term := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		v, _ := obj["match_expressions"].([]any)
		obj["match_expressions"] = flattenNodeSelectorRequirements(term.MatchExpressions, v)
		v, _ = obj["match_fields"].([]any)
		obj["match_fields"] = flattenNodeSelectorRequirements(term.MatchFields, v)
		terms[i] = obj
	}
	return terms
}

func flattenNodeSelectorRequirements(in []corev1.NodeSelectorRequirement, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	requirements := make([]any, len(in))
	for i, req := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
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

	v, ok := d.Get("spec").([]any)
	if !ok {
		v = []any{}
	}

	var ret []any
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

func flattenPodAntiAffinity(in *corev1.PodAntiAffinity, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["required_during_scheduling_ignored_during_execution"].([]any)
	obj["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	v, _ = obj["preferred_during_scheduling_ignored_during_execution"].([]any)
	obj["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	return []any{obj}
}

func flattenPodAffinity(in *corev1.PodAffinity, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["required_during_scheduling_ignored_during_execution"].([]any)
	obj["required_during_scheduling_ignored_during_execution"] = flattenPodAffinityTerms(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	v, _ = obj["preferred_during_scheduling_ignored_during_execution"].([]any)
	obj["preferred_during_scheduling_ignored_during_execution"] = flattenWeightedPodAffinityTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	return []any{obj}
}

func flattenPodAffinityTerms(in []corev1.PodAffinityTerm, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	terms := make([]any, len(in))
	for i, term := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		v, _ := obj["label_selector"].([]any)
		obj["label_selector"] = flattenLabelSelector(term.LabelSelector, v)
		obj["namespaces"] = toArrayInterface(term.Namespaces)
		obj["topology_key"] = term.TopologyKey
		v, _ = obj["namespace_selector"].([]any)
		obj["namespace_selector"] = flattenLabelSelector(term.NamespaceSelector, v)
		terms[i] = obj
	}

	return terms
}

func flattenWeightedPodAffinityTerms(in []corev1.WeightedPodAffinityTerm, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	terms := make([]any, len(in))
	for i, term := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["weight"] = term.Weight
		v, _ := obj["pod_affinity_term"].([]any)
		obj["pod_affinity_term"] = flattenPodAffinityTerms([]corev1.PodAffinityTerm{term.PodAffinityTerm}, v)
		terms[i] = obj
	}

	return terms
}

func flattenPreferredSchedulingTerms(in []corev1.PreferredSchedulingTerm, p []any) []any {
	if len(in) == 0 {
		return nil
	}

	terms := make([]any, len(in))
	for i, term := range in {
		obj := make(map[string]any)
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
		}
		obj["weight"] = term.Weight
		v, _ := obj["preference"].([]any)
		obj["preference"] = flattenNodeSelectorTerm(term.Preference, v)
		terms[i] = obj
	}

	return terms
}

func flattenNodeSelectorTerm(in corev1.NodeSelectorTerm, p []any) []any {
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["match_expressions"].([]any)
	obj["match_expressions"] = flattenNodeSelectorRequirements(in.MatchExpressions, v)
	v, _ = obj["match_fields"].([]any)
	obj["match_fields"] = flattenNodeSelectorRequirements(in.MatchFields, v)
	return []any{obj}
}

func flattenWorkflowHandlerSpec(in *eaaspb.WorkflowHandlerSpec, p []any) ([]any, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flatten workflow handler spec empty input")
	}

	obj := map[string]any{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["config"].([]any)
	obj["config"] = flattenWorkflowHandlerConfig(in.Config, v)
	obj["sharing"] = flattenSharingSpec(in.Sharing)
	obj["inputs"] = flattenConfigContextCompoundRefs(in.Inputs)
	obj["outputs"] = flattenWorkflowHandlerOutputs(in.Outputs)
	obj["icon_url"] = in.IconURL
	obj["readme"] = in.Readme
	return []any{obj}, nil
}

func flattenWorkflowHandlerConfig(input *eaaspb.WorkflowHandlerConfig, p []any) []any {
	log.Println("flatten workflow handler config start", input)
	if input == nil {
		return nil
	}

	obj := map[string]any{}
	if len(p) > 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["type"] = input.Type
	obj["timeout_seconds"] = input.TimeoutSeconds
	obj["success_condition"] = input.SuccessCondition
	obj["max_retry_count"] = input.MaxRetryCount

	v, _ := obj["container"].([]any)
	obj["container"] = flattenWorkflowHandlerContainerConfig(input.Container, v)

	v, _ = obj["http"].([]any)
	obj["http"] = flattenWorkflowHandlerHttpConfig(input.Http, v)

	v, _ = obj["function"].([]any)
	obj["function"] = flattenWorkflowHandlerFunctionConfig(input.Function, v)

	obj["polling_config"] = flattenPollingConfig(input.PollingConfig)

	return []any{obj}
}

func flattenWorkflowHandlerContainerConfig(in *eaaspb.ContainerDriverConfig, p []any) []any {
	log.Println("flatten container workflow handler config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["arguments"] = toArrayInterface(in.Arguments)
	obj["commands"] = toArrayInterface(in.Commands)
	obj["cpu_limit_milli"] = in.CpuLimitMilli
	obj["env_vars"] = toMapInterface(in.EnvVars)
	obj["files"] = toMapByteInterface(in.Files)
	obj["image"] = in.Image

	v, _ := obj["image_pull_credentials"].([]any)
	obj["image_pull_credentials"] = flattenImagePullCredentials(in.ImagePullCredentials, v)

	v, _ = obj["kube_config_options"].([]any)
	obj["kube_config_options"] = flattenContainerKubeConfig(in.KubeConfigOptions, v)

	v, _ = obj["kube_options"].([]any)
	obj["kube_options"] = flattenContainerKubeOptions(in.KubeOptions, v)

	obj["memory_limit_mb"] = in.MemoryLimitMb

	v, _ = obj["volume_options"].([]any)
	obj["volume_options"] = flattenContainerWorkflowHandlerVolumeOptions(
		[]*eaaspb.ContainerDriverVolumeOptions{in.VolumeOptions}, v,
	)

	v, _ = obj["volumes"].([]any)
	obj["volumes"] = flattenContainerWorkflowHandlerVolumeOptions(in.Volumes, v)
	obj["working_dir_path"] = in.WorkingDirPath
	return []any{obj}
}

func flattenImagePullCredentials(in *eaaspb.ContainerImagePullCredentials, p []any) []any {
	log.Println("flatten container image pull credentials start")
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["registry"] = in.Registry
	obj["username"] = in.Username
	obj["password"] = in.Password
	return []any{obj}
}

func flattenContainerKubeConfig(in *eaaspb.ContainerKubeConfigOptions, p []any) []any {
	log.Println("flatten container kube config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["kube_config"] = in.KubeConfig
	obj["out_of_cluster"] = in.OutOfCluster
	return []any{obj}
}

func flattenContainerKubeOptions(in *eaaspb.ContainerKubeOptions, p []any) []any {
	log.Println("flatten container kube options start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["labels"] = toMapInterface(in.Labels)
	obj["namespace"] = in.Namespace
	obj["node_selector"] = toMapInterface(in.NodeSelector)
	obj["resources"] = toArrayInterface(in.Resources)

	v, _ := obj["security_context"].([]any)
	obj["security_context"] = flattenSecurityContext(in.SecurityContext, v)
	obj["service_account_name"] = in.ServiceAccountName

	if len(in.Tolerations) > 0 {
		v, _ = obj["tolerations"].([]any)
		obj["tolerations"] = flattenV3Tolerations(in.Tolerations, v)
	} else {
		delete(obj, "tolerations")
	}

	v, _ = obj["affinity"].([]any)
	obj["affinity"] = flattenKubeOptionsAffinity(in.Affinity, v)
	return []any{obj}
}

func flattenSecurityContext(in *eaaspb.KubeSecurityContext, p []any) []any {
	log.Println("flatten kube security context start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}

	obj["privileged"] = flattenBoolValue(in.Privileged)
	obj["read_only_root_file_system"] = flattenBoolValue(in.ReadOnlyRootFileSystem)

	return []any{obj}
}

func flattenKubeOptionsAffinity(in *corev1.Affinity, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["node_affinity"].([]any)
	obj["node_affinity"] = flattenNodeAffinity(in.NodeAffinity, v)
	v, _ = obj["pod_affinity"].([]any)
	obj["pod_affinity"] = flattenPodAffinity(in.PodAffinity, v)
	v, _ = obj["pod_anti_affinity"].([]any)
	obj["pod_anti_affinity"] = flattenPodAntiAffinity(in.PodAntiAffinity, v)
	return []any{obj}
}

func flattenNodeAffinity(in *corev1.NodeAffinity, p []any) []any {
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	v, _ := obj["required_during_scheduling_ignored_during_execution"].([]any)
	obj["required_during_scheduling_ignored_during_execution"] = flattenNodeSelector(in.RequiredDuringSchedulingIgnoredDuringExecution, v)
	v, _ = obj["preferred_during_scheduling_ignored_during_execution"].([]any)
	obj["preferred_during_scheduling_ignored_during_execution"] = flattenPreferredSchedulingTerms(in.PreferredDuringSchedulingIgnoredDuringExecution, v)
	return []any{obj}
}

func flattenWorkflowHandlerHttpConfig(in *eaaspb.HTTPDriverConfig, p []any) []any {
	log.Println("flatten http config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["body"] = in.Body
	obj["endpoint"] = in.Endpoint
	obj["headers"] = toMapInterface(in.Headers)
	obj["method"] = in.Method
	return []any{obj}
}

func flattenWorkflowHandlerFunctionConfig(in *eaaspb.FunctionDriverConfig, p []any) []any {
	log.Println("flatten function config start")
	if in == nil {
		return nil
	}

	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["name"] = in.Name
	obj["language"] = in.Language
	obj["source"] = in.Source
	obj["function_dependencies"] = toArrayInterface(in.FunctionDependencies)
	obj["system_packages"] = toArrayInterface(in.SystemPackages)
	obj["target_platforms"] = toArrayInterface(in.TargetPlatforms)
	obj["language_version"] = in.LanguageVersion
	obj["cpu_limit_milli"] = in.CpuLimitMilli
	obj["memory_limit_mb"] = in.MemoryLimitMb
	obj["skip_build"] = flattenBoolValue(in.SkipBuild)
	obj["image"] = in.Image
	obj["max_concurrency"] = in.MaxConcurrency
	obj["num_replicas"] = in.NumReplicas
	obj["inactivity_timeout_seconds"] = in.InactivityTimeoutSeconds

	v, _ := obj["kube_options"].([]any)
	obj["kube_options"] = flattenContainerKubeOptions(in.KubeOptions, v)

	v, _ = obj["image_pull_credentials"].([]any)
	obj["image_pull_credentials"] = flattenImagePullCredentials(in.ImagePullCredentials, v)

	v, _ = obj["resources"].([]any)
	obj["resources"] = flattenDriverResources(in.Resources, v)

	v, _ = obj["hpa"].([]any)
	obj["hpa"] = flattenFunctionHPAConfig(in.Hpa, v)

	return []any{obj}
}

func flattenDriverResourceValues(in *eaaspb.DriverResourceValues, p []any) []any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["cpu"] = in.Cpu
	obj["memory"] = in.Memory
	return []any{obj}
}

func flattenDriverResources(in *eaaspb.DriverResources, p []any) []any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	var v []any
	v, _ = obj["requests"].([]any)
	obj["requests"] = flattenDriverResourceValues(in.Requests, v)
	v, _ = obj["limits"].([]any)
	obj["limits"] = flattenDriverResourceValues(in.Limits, v)
	obj["unset"] = flattenBoolValue(in.Unset)
	return []any{obj}
}

func flattenFunctionHPAConfig(in *eaaspb.FunctionHPAConfig, p []any) []any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]any)
	}
	obj["enabled"] = flattenBoolValue(in.Enabled)
	obj["min_replicas"] = in.MinReplicas
	obj["max_replicas"] = in.MaxReplicas
	var v []any
	v, _ = obj["resource_metrics"].([]any)
	obj["resource_metrics"] = flattenResourceMetrics(in.ResourceMetrics, v)
	return []any{obj}
}

func flattenMetricTarget(in *autoscalingv2.MetricTarget) map[string]any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	obj["type"] = string(in.Type)
	if in.Value != nil {
		obj["value"] = in.Value.String()
	}
	if in.AverageValue != nil {
		obj["average_value"] = in.AverageValue.String()
	}
	if in.AverageUtilization != nil {
		obj["average_utilization"] = int(*in.AverageUtilization)
	}
	return obj
}

func flattenResourceMetricSource(in *autoscalingv2.ResourceMetricSource) map[string]any {
	if in == nil {
		return nil
	}
	obj := make(map[string]any)
	obj["name"] = string(in.Name)
	obj["target"] = []any{flattenMetricTarget(&in.Target)}
	return obj
}

func flattenResourceMetrics(in []*autoscalingv2.ResourceMetricSource, p []any) []any {
	if len(in) == 0 {
		return nil
	}
	out := make([]any, len(in))
	for i, rms := range in {
		out[i] = flattenResourceMetricSource(rms)
	}
	return out
}

func flattenContainerWorkflowHandlerVolumeOptions(input []*eaaspb.ContainerDriverVolumeOptions, p []any) []any {
	if len(input) == 0 {
		return nil
	}

	var out []any
	for i, in := range input {
		if in == nil {
			continue
		}
		log.Println("flatten container workflow handler volume options", in)
		obj := map[string]any{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]any)
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

func flattenWorkflowHandlerCompoundRef(input *eaaspb.WorkflowHandlerCompoundRef) []any {
	log.Println("flatten workflow handler compound ref start")
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"name": input.Name,
		"data": flattenWorkflowHandlerInline(input.Data),
	}
	return []any{obj}
}

func flattenWorkflowHandlerInline(input *eaaspb.WorkflowHandlerInline) []any {
	if input == nil {
		return nil
	}
	obj := map[string]any{
		"config":  flattenWorkflowHandlerConfig(input.Config, []any{}),
		"inputs":  flattenConfigContextCompoundRefs(input.Inputs),
		"outputs": flattenWorkflowHandlerOutputs(input.Outputs),
	}
	return []any{obj}
}

func flattenPollingConfig(in *eaaspb.PollingConfig) []any {
	if in == nil {
		return nil
	}

	obj := map[string]any{
		"repeat": in.Repeat,
		"until":  in.Until,
	}
	return []any{obj}
}

func resourceWorkflowHandlerImport(_ context.Context, d *schema.ResourceData, _ any) ([]*schema.ResourceData, error) {
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
