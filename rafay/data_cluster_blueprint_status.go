package rafay

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/RafaySystems/rafay-common/proto/types/hub/commonpb"
	"github.com/RafaySystems/rctl/pkg/config"
	"github.com/RafaySystems/rctl/pkg/rerror"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ClusterBPStatus = &schema.Resource{
	Description: "ClusterBlueprintStatus",
	Schema: map[string]*schema.Schema{
		"metadata": &schema.Schema{
			Description: "Metadata of the cluster resource",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"annotations": &schema.Schema{
					Description: "annotations of the resource",
					Elem:        &schema.Schema{Type: schema.TypeString},
					Optional:    true,
					Type:        schema.TypeMap,
				},
				"created_by": &schema.Schema{
					Description: "User who created this resource",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Id of the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"is_sso_user": &schema.Schema{
							Description: "Whether person is logged in using sso",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"username": &schema.Schema{
							Description: "Username fo the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},
				"description": &schema.Schema{
					Description: "description of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"display_name": &schema.Schema{
					Description: "Display Name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"labels": &schema.Schema{
					Description: "labels of the resource",
					Elem:        &schema.Schema{Type: schema.TypeString},
					Optional:    true,
					Type:        schema.TypeMap,
				},
				"modified_by": &schema.Schema{
					Description: "User who last modified this resource",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Description: "Id of the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"is_sso_user": &schema.Schema{
							Description: "Whether person is logged in using sso",
							Optional:    true,
							Type:        schema.TypeBool,
						},
						"username": &schema.Schema{
							Description: "Username fo the Person",
							Optional:    true,
							Type:        schema.TypeString,
						},
					}},
					MaxItems: 1,
					MinItems: 1,
					Optional: true,
					Type:     schema.TypeList,
				},
				"name": &schema.Schema{
					Description: "name of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
				"project": &schema.Schema{
					Description: "Project of the resource",
					Optional:    true,
					Type:        schema.TypeString,
				},
			}},
			MaxItems: 1,
			MinItems: 1,
			Optional: true,
			Type:     schema.TypeList,
		},
		"status": &schema.Schema{
			Description: "status of the cluster blueprint",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"addon_status_list": &schema.Schema{
					Description: "cluster blueprint addon status",
					Elem: &schema.Resource{Schema: map[string]*schema.Schema{
						"addon_name": &schema.Schema{
							Description: "name of the blueprint addon",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"namespace": &schema.Schema{
							Description: "namespce of the blueprint addon",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"publish_status": &schema.Schema{
							Description: "status of the blueprint addon",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"publish_time": &schema.Schema{
							Description: "time of the blueprint addon publish",
							Optional:    true,
							Type:        schema.TypeString,
						},
						"conditions": &schema.Schema{
							Description: "conditions blueprint addon status",
							Elem: &schema.Resource{Schema: map[string]*schema.Schema{
								"type": &schema.Schema{
									Description: "condition type of the blueprint addon",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"status": &schema.Schema{
									Description: "condition status of the blueprint addon",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"last_updated": &schema.Schema{
									Description: "condition updated time of the blueprint addon",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"reason": &schema.Schema{
									Description: "condition reason of the blueprint addon publish",
									Optional:    true,
									Type:        schema.TypeString,
								},
								"retrycount": &schema.Schema{
									Description: "retry count of blueprint addon publish",
									Optional:    true,
									Type:        schema.TypeInt,
								},
								"max_retrycount": &schema.Schema{
									Description: "maxium retry count of blueprint addon publish",
									Optional:    true,
									Type:        schema.TypeInt,
								},
								"components": &schema.Schema{
									Description: "components of blueprint addon",
									Elem: &schema.Resource{Schema: map[string]*schema.Schema{
										"name": &schema.Schema{
											Description: "component name of the blueprint addon",
											Optional:    true,
											Type:        schema.TypeString,
										},
										"reason": &schema.Schema{
											Description: "component failure reason of the blueprint addon publish",
											Optional:    true,
											Type:        schema.TypeString,
										},
										"failures": &schema.Schema{
											Description: "component failures blueprint addon",
											Elem: &schema.Resource{Schema: map[string]*schema.Schema{
												"message": &schema.Schema{
													Description: "component failure message of the blueprint addon",
													Optional:    true,
													Type:        schema.TypeString,
												},
												"timestamp": &schema.Schema{
													Description: "component failure timestamp of the blueprint addon publish",
													Optional:    true,
													Type:        schema.TypeString,
												},
												"reason": &schema.Schema{
													Description: "component failure reason the blueprint addon publish",
													Optional:    true,
													Type:        schema.TypeString,
												},
												"name": &schema.Schema{
													Description: "component name blueprint addon publish",
													Optional:    true,
													Type:        schema.TypeString,
												},
											}},
											MaxItems: 0,
											MinItems: 0,
											Optional: true,
											Type:     schema.TypeList,
										},
									}},
									MaxItems: 0,
									MinItems: 0,
									Optional: true,
									Type:     schema.TypeList,
								},
							}},
							MaxItems: 0,
							MinItems: 0,
							Optional: true,
							Type:     schema.TypeList,
						},
					}},
					MaxItems: 0,
					MinItems: 0,
					Optional: true,
					Type:     schema.TypeList,
				},
			}},
			MaxItems: 0,
			MinItems: 0,
			Optional: true,
			Type:     schema.TypeList,
		},
	},
}

type ClusterBlueprintWorkloadConditionsComponentsFailures struct {
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Name      string `json:"name,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

type ClusterBlueprintWorkloadConditionsComponents struct {
	Name     string                                                 `json:"name"`
	Reason   string                                                 `json:"reason"`
	Failures []ClusterBlueprintWorkloadConditionsComponentsFailures `json:"failures"`
}

type ClusterBlueprintWorkloadConditions struct {
	Type        string                                         `json:"type,omitempty"`
	Status      string                                         `json:"status,omitempty"`
	LastUpdated string                                         `json:"lastUpdated,omitempty"`
	Reason      string                                         `json:"reason,omitempty"`
	RetryCount  int                                            `json:"retryCount,omitempty"`
	MaxRetry    int                                            `json:"maxRetryCount,omitempty"`
	Components  []ClusterBlueprintWorkloadConditionsComponents `json:"components,omitempty"`
}

type ClusterBlueprintWorkloads struct {
	WorkloadName  string                               `json:"workloadName,omitempty"`
	Namespace     string                               `json:"namespace,omitempty"`
	PublishedAt   string                               `json:"publishedAt,omitempty"`
	PublishStatus string                               `json:"publishStatus,omitempty"`
	Conditions    []ClusterBlueprintWorkloadConditions `json:"conditions,omitempty"`
}

type ClusterBlueprintStatus struct {
	Metadata    *commonpb.Metadata          `protobuf:"bytes,3,opt,name=metadata,proto3" json:"metadata,omitempty"`
	ClusterName string                      `json:"clusterName,omitempty"`
	ClusterID   string                      `json:"clusterID,omitempty"`
	Workloads   []ClusterBlueprintWorkloads `json:"workloads,omitempty"`
}

func dataClusterBlueprintStatus() *schema.Resource {
	return &schema.Resource{
		Description: "The Cluster Blueprint Status data allows access to the blueprint status of cluster resource",
		ReadContext: dataClusterBlueprintStatusRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(10 * time.Minute),
		},
		SchemaVersion: 1,
		Schema:        ClusterBPStatus.Schema,
	}
}

func dataClusterBlueprintStatusRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Println("dataClusterBlueprintStatusRead ")
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

	// XXX Debug
	// w1 := spew.Sprintf("%+v", tfProjectState)
	// log.Println("dataProjectRead tfProjectState", w1)

	// auth := config.GetConfig().GetAppAuthProfile()
	// client, err := typed.NewClientWithUserAgent(auth.URL, auth.Key, TF_USER_AGENT, options.WithInsecureSkipVerify(auth.SkipServerCertValid))
	// if err != nil {
	// 	return diag.FromErr(err)
	// }

	// find cluster name and project name
	clusterName, ok := d.Get("metadata.0.name").(string)
	if !ok || clusterName == "" {
		log.Print("Cluster name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "cluster name is missing"))
	}
	projectName, ok := d.Get("metadata.0.project").(string)
	if !ok || projectName == "" {
		log.Print("Cluster project name unable to be found")
		return diag.FromErr(fmt.Errorf("%s", "project name is missing"))
	}
	projectID, err := getProjectIDFromName(projectName)
	if err != nil {
		log.Print("error converting project name to id")
		return diag.Errorf("error converting project name to project ID")
	}

	clBpStatus, err := getClusterBlueprintStatus(clusterName, projectID)
	if err != nil {
		log.Printf("error in get cluster %s blueprint status", err.Error())
		if strings.Contains(err.Error(), "not found") {
			log.Println("Resource Read ", "error", err)
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}
	clBpStatus.Metadata = meta
	log.Println("got cluster blueprint status from backend", clBpStatus)

	// XXX Debug
	// w1 := spew.Sprintf("%+v", clBpStatus)
	// log.Println("dataProjectRead wl", w1)

	err = flattenClusterBlueprintStatus(d, clBpStatus)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(clusterName)

	return diags

}

func getClusterBlueprintStatus(clusterName, projectId string) (*ClusterBlueprintStatus, error) {
	uri := fmt.Sprintf("/v2/scheduler/project/%s/cluster/%s/workloadsummary?workloadType=system", projectId, clusterName)
	auth := config.GetConfig().GetAppAuthProfile()
	respString, err := auth.AuthAndRequest(uri, "GET", nil)
	if err != nil {
		return nil, rerror.CrudErr{
			Type: "cluster labels",
			Name: clusterName,
			Op:   "get",
		}
	}

	var resp ClusterBlueprintStatus

	if err := json.Unmarshal([]byte(respString), &resp); err != nil {
		log.Printf("Error unmarshaling response from get v2 cluster: %s", err)
		return nil, err
	}

	return &resp, nil
}

// Flatteners

func flattenClusterBlueprintStatus(d *schema.ResourceData, in *ClusterBlueprintStatus) error {
	if in == nil {
		return nil
	}

	err := d.Set("metadata", flattenMetaData(in.Metadata))
	if err != nil {
		return err
	}

	v, ok := d.Get("status").([]interface{})
	if !ok {
		v = []interface{}{}
	}

	var ret []interface{}
	ret, err = flattenClusterBlueprintStatusList(in, v)
	if err != nil {
		return err
	}

	err = d.Set("status", ret)
	if err != nil {
		return err
	}
	return nil
}

func flattenClusterBlueprintStatusList(in *ClusterBlueprintStatus, p []interface{}) ([]interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("%s", "flattenAddonStatusList empty input")
	}

	obj := map[string]interface{}{}
	if len(p) != 0 && p[0] != nil {
		obj = p[0].(map[string]interface{})
	}

	if len(in.Workloads) > 0 {
		v, ok := obj["addon_status_list"].([]interface{})
		if !ok {
			v = []interface{}{}
		}
		obj["addon_status_list"] = flattenAddonStatus(in.Workloads, v)
	}

	return []interface{}{obj}, nil
}

func flattenAddonStatus(input []ClusterBlueprintWorkloads, p []interface{}) []interface{} {
	log.Println("flattenAddonStatus")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.WorkloadName) > 0 {
			obj["addon_name"] = in.WorkloadName
		}

		if len(in.Namespace) > 0 {
			obj["namespace"] = in.Namespace
		}

		if len(in.PublishStatus) > 0 {
			obj["publish_status"] = in.PublishStatus
		}

		if len(in.PublishedAt) > 0 {
			obj["publish_time"] = in.PublishedAt
		}

		if len(in.Conditions) > 0 {
			v, ok := obj["conditions"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["conditions"] = flattenAddonStatusConditions(in.Conditions, v)
		}

		out[i] = &obj
	}

	return out
}

func flattenAddonStatusConditions(input []ClusterBlueprintWorkloadConditions, p []interface{}) []interface{} {
	log.Println("flattenAddonStatusConditions")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Type) > 0 {
			obj["type"] = in.Type
		}

		if len(in.Status) > 0 {
			obj["status"] = in.Status
		}

		if len(in.LastUpdated) > 0 {
			obj["last_updated"] = in.LastUpdated
		}

		if len(in.Reason) > 0 {
			obj["reason"] = in.Reason
		}

		obj["retrycount"] = in.RetryCount

		obj["max_retrycount"] = in.MaxRetry

		if len(in.Components) > 0 {
			v, ok := obj["components"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["components"] = flattenAddonStatusConditionComps(in.Components, v)
		}
		out[i] = &obj
	}

	return out
}

func flattenAddonStatusConditionComps(input []ClusterBlueprintWorkloadConditionsComponents, p []interface{}) []interface{} {
	log.Println("flattenAddonStatusConditionComps")
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

		if len(in.Reason) > 0 {
			obj["reason"] = in.Reason
		}

		if len(in.Failures) > 0 {
			v, ok := obj["failures"].([]interface{})
			if !ok {
				v = []interface{}{}
			}
			obj["failures"] = flattenAddonStatusConditionCompFails(in.Failures, v)
		}

		out[i] = &obj
	}

	return out
}

func flattenAddonStatusConditionCompFails(input []ClusterBlueprintWorkloadConditionsComponentsFailures, p []interface{}) []interface{} {
	log.Println("flattenAddonStatusConditionCompFails")
	if input == nil {
		return nil
	}
	out := make([]interface{}, len(input))
	for i, in := range input {

		obj := map[string]interface{}{}
		if i < len(p) && p[i] != nil {
			obj = p[i].(map[string]interface{})
		}

		if len(in.Message) > 0 {
			obj["message"] = in.Message
		}

		if len(in.Timestamp) > 0 {
			obj["timestamp"] = in.Timestamp
		}

		if len(in.Name) > 0 {
			obj["name"] = in.Name
		}

		if len(in.Reason) > 0 {
			obj["reason"] = in.Reason
		}

		out[i] = &obj
	}

	return out
}
