package rafay

import (
	"log"
	"reflect"

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