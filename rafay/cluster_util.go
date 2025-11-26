package rafay

import (
	"log"
	"reflect"
	"slices"

	"github.com/RafaySystems/rctl/pkg/cluster"
	"github.com/RafaySystems/rctl/pkg/models"
	"github.com/RafaySystems/rctl/pkg/project"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// systemComponentsPlacement schema defined
func systemComponentsPlacementFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selector": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "Key-Value pairs insuring pods to be scheduled on desired nodes.",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "Enables the kuberenetes scheduler to schedule pods with matching taints.",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
		"daemonset_override": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Allows users to override the default behaviour of DaemonSet for specific nodes, enabling the addition of additional tolerations for Daemonsets to match the taints available on the nodes.",
			Elem: &schema.Resource{
				Schema: daemonsetOverrideFields(),
			},
		},
	}
	return s
}

func tolerationsFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"key": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint key that the toleration applies to",
		},
		"operator": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "represents a key's relationship to the value",
		},
		"value": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "the taint value the toleration matches to",
		},
		"effect": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "indicates the taint effect to match",
		},
		"toleration_seconds": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "represents the period of time the toleration tolerates the taint",
		},
	}
	return s
}

func daemonsetOverrideFields() map[string]*schema.Schema {
	s := map[string]*schema.Schema{
		"node_selection_enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "enables node selection",
		},
		"tolerations": {
			Type: schema.TypeList,
			//Type:        schema.TypeString,
			Optional:    true,
			Description: "Additional tolerations for Daemonsets to match the taints available on the nodes",
			Elem: &schema.Resource{
				Schema: tolerationsFields(),
			},
		},
	}
	return s
}

// systemComponentsPlacement v1 expand functionality defined
func expandSystemComponentsPlacement(p []interface{}) *SystemComponentsPlacement {
	obj := &SystemComponentsPlacement{}
	log.Println("expandSystemComponentsPlacement")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["node_selector"].(map[string]interface{}); ok && len(v) > 0 {
		obj.NodeSelector = toMapString(v)
	}

	if v, ok := in["tolerations"].([]interface{}); ok && len(v) > 0 {
		obj.Tolerations = expandTolerations(v)
	}
	if v, ok := in["daemonset_override"].([]interface{}); ok && len(v) > 0 {
		obj.DaemonsetOverride = expandDaemonsetOverride(v)
	}
	return obj
}

func expandTolerations(p []interface{}) []*Tolerations {
	out := make([]*Tolerations, len(p))
	if len(p) == 0 || p[0] == nil {
		return out
	}
	for i := range p {
		obj := &Tolerations{}
		in := p[i].(map[string]interface{})

		if v, ok := in["key"].(string); ok && len(v) > 0 {
			obj.Key = v
		}
		if v, ok := in["operator"].(string); ok && len(v) > 0 {
			obj.Operator = v
		}
		if v, ok := in["value"].(string); ok && len(v) > 0 {
			obj.Value = v
		}
		if v, ok := in["effect"].(string); ok && len(v) > 0 {
			obj.Effect = v
		}
		if v, ok := in["toleration_seconds"].(int); ok {
			if v == 0 {
				obj.TolerationSeconds = nil
			} else {
				log.Println("setting toleration seconds")
				obj.TolerationSeconds = &v
			}
		}
		out[i] = obj
	}
	return out
}

func expandDaemonsetOverride(p []interface{}) *DaemonsetOverride {
	obj := &DaemonsetOverride{}
	log.Println("expand CNI params")

	if len(p) == 0 || p[0] == nil {
		return obj
	}
	in := p[0].(map[string]interface{})

	if v, ok := in["node_selection_enabled"].(bool); ok {
		obj.NodeSelectionEnabled = &v
	}
	if v, ok := in["tolerations"].([]interface{}); ok && len(v) > 0 {
		obj.Tolerations = expandTolerations(v)
	}
	return obj
}

// systemComponentsPlacement v1 flatten functionality defined
func flattenSystemComponentsPlacement(in *SystemComponentsPlacement, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}
	log.Println("got to flatten system comp:", in)
	log.Println("node_selectopr type: ", reflect.TypeOf(in.NodeSelector))
	if in.NodeSelector != nil && len(in.NodeSelector) > 0 {
		obj["node_selector"] = toMapInterface(in.NodeSelector)
	}
	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		log.Println("type of read tolerations:", reflect.TypeOf(in.Tolerations), in.Tolerations)
		obj["tolerations"] = flattenTolerations(in.Tolerations, v)
	}
	if in.DaemonsetOverride != nil {
		v, ok := obj["daemonset_override"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["daemonset_override"] = flattenDaemonsetOverride(in.DaemonsetOverride, v)
	}

	return []interface{}{obj}
}

func flattenTolerations(in []*Tolerations, p []interface{}) []interface{} {
	if in == nil {
		return nil
	}
	log.Println("flattenTolerations")
	out := make([]interface{}, len(in))
	for i, in := range in {
		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Key) > 0 {
			obj["key"] = in.Key
		}
		if len(in.Operator) > 0 {
			obj["operator"] = in.Operator
		}
		if len(in.Value) > 0 {
			obj["value"] = in.Value
		}
		if len(in.Effect) > 0 {
			obj["effect"] = in.Effect
		}
		if in.TolerationSeconds != nil {
			obj["toleration_seconds"] = in.TolerationSeconds
		}

		out[i] = &obj
	}
	return out
}

func flattenDaemonsetOverride(in *DaemonsetOverride, p []interface{}) []interface{} {
	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	obj["node_selection_enabled"] = in.NodeSelectionEnabled

	if in.Tolerations != nil {
		v, ok := obj["tolerations"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["tolerations"] = flattenTolerations(in.Tolerations, v)
	}

	return []interface{}{obj}
}

func flattenV1ClusterSharing(in *V1ClusterSharing) []interface{} {
	if in == nil {
		return nil
	}
	obj := make(map[string]interface{})
	if in.Enabled != nil {
		obj["enabled"] = *in.Enabled
	}
	if len(in.Projects) > 0 {
		obj["projects"] = flattenV1SharingProjects(in.Projects)
	}
	return []interface{}{obj}
}

func flattenV1SharingProjects(in []*V1ClusterSharingProject) []interface{} {
	if len(in) == 0 {
		return nil
	}
	var out []interface{}
	for _, x := range in {
		obj := make(map[string]interface{})
		if len(x.Name) > 0 {
			obj["name"] = x.Name
		}
		out = append(out, obj)
	}
	return out
}

func expandV1ClusterSharing(p []interface{}) *V1ClusterSharing {
	obj := &V1ClusterSharing{}
	if len(p) == 0 || p[0] == nil {
		return obj
	}

	in := p[0].(map[string]interface{})
	if v, ok := in["enabled"].(bool); ok {
		obj.Enabled = &v
	}
	if v, ok := in["projects"].([]interface{}); ok && len(v) > 0 {
		obj.Projects = expandV1ClusterSharingProjects(v)
	}
	return obj
}

func expandV1ClusterSharingProjects(p []interface{}) []*V1ClusterSharingProject {
	if len(p) == 0 {
		return nil
	}
	var res []*V1ClusterSharingProject
	for i := range p {
		in := p[i].(map[string]interface{})
		obj := &V1ClusterSharingProject{}
		if v, ok := in["name"].(string); ok && len(v) > 0 {
			obj.Name = v
		}
		res = append(res, obj)
	}
	return res
}

var BlueprintSyncConditions = []models.ClusterConditionType{
	models.ClusterRegister,
	models.ClusterCheckIn,
	models.ClusterNamespaceSync,
	models.ClusterBlueprintSync,
}

func getProjectIDFromName(projectName string) (string, error) {
	// derive project id from project name
	resp, err := project.GetProjectByName(projectName)
	if err != nil {
		log.Print("project name missing in the resource")
		return "", err
	}

	project, err := project.NewProjectFromResponse([]byte(resp))
	if err != nil {
		log.Printf("project does not exist")
		return "", err
	}
	return project.ID, nil
}

func getClusterConditions(edgeId, projectId string) (bool, bool, error) {
	cluster, err := cluster.GetClusterWithEdgeID(edgeId, projectId, uaDef)
	if err != nil {
		log.Printf("error while getCluster %s", err.Error())
		return false, false, err
	}

	clusterConditions := cluster.Cluster.Conditions
	failureFlag := false
	readyFlag := false
	for _, condition := range clusterConditions {
		if slices.Contains(BlueprintSyncConditions, condition.Type) && condition.Status == models.Failed {
			failureFlag = true
		}
		if condition.Type == models.ClusterReady && condition.Status == models.Success {
			readyFlag = true
		}
	}
	return failureFlag, readyFlag, nil
}
