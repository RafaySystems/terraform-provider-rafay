package rafay

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	schema "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAKSClusterV3() *schema.Resource {
	return &schema.Resource{
		Description: "",
		Schema: map[string]*schema.Schema{
			"metadata": &schema.Schema{
				Description: "",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"annotations": &schema.Schema{
						Description: "annotations of the resource",
						Elem:        &schema.Schema{Type: schema.TypeString},
						Optional:    true,
						Type:        schema.TypeMap,
					},
					"description": &schema.Schema{
						Description: "description of the resource",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"labels": &schema.Schema{
						Description: "labels of the resource",
						Elem:        &schema.Schema{Type: schema.TypeString},
						Optional:    true,
						Type:        schema.TypeMap,
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
				Optional: false,
				Type:     schema.TypeList,
			},
			"spec": &schema.Schema{
				Description: "",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"blueprint_config": &schema.Schema{
						Description: "The blueprint configuration to be used for this cluster",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"name": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"version": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"cloud_credentials": &schema.Schema{
						Description: "The credentials to be used to interact with the cloud infrastructure",
						Optional:    true,
						Type:        schema.TypeString,
					},
					"config": &schema.Schema{
						Description: "",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"api_version": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"kind": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"metadata": &schema.Schema{
								Description: "",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"annotations": &schema.Schema{
										Description: "annotations of the resource",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Optional:    true,
										Type:        schema.TypeMap,
									},
									"description": &schema.Schema{
										Description: "description of the resource",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"labels": &schema.Schema{
										Description: "labels of the resource",
										Elem:        &schema.Schema{Type: schema.TypeString},
										Optional:    true,
										Type:        schema.TypeMap,
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
							"spec": &schema.Schema{
								Description: "",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"managed_cluster": &schema.Schema{
										Description: "",
										Elem: &schema.Resource{Schema: map[string]*schema.Schema{
											"additional_metadata": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"acr_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"acr_name": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"resource_group_name": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"oms_workspace_location": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"api_version": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"extended_location": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"name": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"identity": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"user_assigned_identities": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"location": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"properties": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"aad_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"admin_group_object_i_ds": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"client_app_id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"enable_azure_rbac": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"managed": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"server_app_id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"server_app_secret": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"tenant_id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"api_server_access_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"authorized_ip_ranges": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"disable_run_command": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"enable_private_cluster": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"enable_private_cluster_public_fqdn": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"private_dns_zone": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"auto_scaler_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"balance_similar_node_groups": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"expander": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"max_empty_bulk_delete": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"max_graceful_termination_sec": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"max_node_provision_time": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"max_total_unready_percentage": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"new_pod_scale_up_delay": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"ok_total_unready_count": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_delay_after_add": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_delay_after_delete": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_delay_after_failure": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_unneeded_time": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_unready_time": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scale_down_utilization_threshold": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"scan_interval": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"skip_nodes_with_local_storage": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"skip_nodes_with_system_pods": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"auto_upgrade_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{"upgrade_channel": &schema.Schema{
															Description: "",
															Optional:    true,
															Type:        schema.TypeString,
														}}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"disable_local_accounts": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"disk_encryption_set_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"dns_prefix": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"enable_pod_security_policy": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"enable_rbac": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"fqdn_subdomain": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"http_proxy_config": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"http_proxy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"https_proxy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"no_proxy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"trusted_ca": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"kubernetes_version": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"linux_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"admin_username": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"ssh": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{"public_keys": &schema.Schema{
																	Description: "",
																	Elem: &schema.Resource{Schema: map[string]*schema.Schema{"key_data": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	}}},
																	MaxItems: 0,
																	MinItems: 0,
																	Optional: true,
																	Type:     schema.TypeList,
																}}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"network_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"dns_service_ip": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"docker_bridge_cidr": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"ip_families": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"load_balancer_profile": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"allocated_outbound_ports": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"effective_outbound_i_ps": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"id": &schema.Schema{
																			Description: "",
																			Optional:    true,
																			Type:        schema.TypeString,
																		}}},
																		MaxItems: 0,
																		MinItems: 0,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																	"enable_multiple_standard_load_balancers": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeBool,
																	},
																	"idle_timeout_in_minutes": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"managed_outbound_i_ps": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																			"count": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeInt,
																			},
																			"count_i_pv_6": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeInt,
																			},
																		}},
																		MaxItems: 1,
																		MinItems: 1,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																	"outbound_i_ps": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"public_i_ps": &schema.Schema{
																			Description: "",
																			Elem: &schema.Resource{Schema: map[string]*schema.Schema{"id": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeString,
																			}}},
																			MaxItems: 0,
																			MinItems: 0,
																			Optional: true,
																			Type:     schema.TypeList,
																		}}},
																		MaxItems: 1,
																		MinItems: 1,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																	"outbound_ip_prefixes": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"public_ip_prefixes": &schema.Schema{
																			Description: "",
																			Elem: &schema.Resource{Schema: map[string]*schema.Schema{"id": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeString,
																			}}},
																			MaxItems: 0,
																			MinItems: 0,
																			Optional: true,
																			Type:     schema.TypeList,
																		}}},
																		MaxItems: 1,
																		MinItems: 1,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"load_balancer_sku": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"nat_gateway_profile": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"effective_outbound_i_ps": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"id": &schema.Schema{
																			Description: "",
																			Optional:    true,
																			Type:        schema.TypeString,
																		}}},
																		MaxItems: 0,
																		MinItems: 0,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																	"idle_timeout_in_minutes": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"managed_outbound_ip_profile": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"count": &schema.Schema{
																			Description: "",
																			Optional:    true,
																			Type:        schema.TypeInt,
																		}}},
																		MaxItems: 1,
																		MinItems: 1,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"network_mode": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"network_plugin": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"network_policy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"outbound_type": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"pod_cidr": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"pod_cidrs": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"service_cidr": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"service_cidrs": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"node_resource_group": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"pod_identity_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"allow_network_plugin_kubenet": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"enabled": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"user_assigned_identities": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"binding_selector": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"identity": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																			"client_id": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeString,
																			},
																			"object_id": &schema.Schema{
																				Description: "",
																				Optional:    true,
																				Type:        schema.TypeString,
																			},
																			"resource_id": &schema.Schema{
																				Description: "",
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
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"namespace": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																}},
																MaxItems: 0,
																MinItems: 0,
																Optional: true,
																Type:     schema.TypeList,
															},
															"user_assigned_identity_exceptions": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"name": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"namespace": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"pod_labels": &schema.Schema{
																		Description: "",
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
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"private_link_resources": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"group_id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"name": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"required_members": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"type": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 0,
														MinItems: 0,
														Optional: true,
														Type:     schema.TypeList,
													},
													"public_network_access": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"security_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"azure_key_vault_kms": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"enabled": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeBool,
																	},
																	"key_id": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"key_vault_network_access": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"key_vault_resource_id": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"defender": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"log_analytics_workspace_resource_id": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"security_monitoring": &schema.Schema{
																		Description: "",
																		Elem: &schema.Resource{Schema: map[string]*schema.Schema{"enabled": &schema.Schema{
																			Description: "",
																			Optional:    true,
																			Type:        schema.TypeBool,
																		}}},
																		MaxItems: 1,
																		MinItems: 1,
																		Optional: true,
																		Type:     schema.TypeList,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"service_principal_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"client_id": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"secret": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"storage_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"disk_csi_driver": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{"enabled": &schema.Schema{
																	Description: "",
																	Optional:    true,
																	Type:        schema.TypeBool,
																}}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"file_csi_driver": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{"enabled": &schema.Schema{
																	Description: "",
																	Optional:    true,
																	Type:        schema.TypeBool,
																}}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"snapshot_controller": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{"enabled": &schema.Schema{
																	Description: "",
																	Optional:    true,
																	Type:        schema.TypeBool,
																}}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"windows_profile": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"admin_username": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"enable_csi_proxy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"gmsa_profile": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"dns_server": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"enabled": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeBool,
																	},
																	"root_domain_name": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"license_type": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"sku": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"name": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"tier": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"tags": &schema.Schema{
												Description: "",
												Elem:        &schema.Schema{Type: schema.TypeString},
												Optional:    true,
												Type:        schema.TypeMap,
											},
											"type": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
										}},
										MaxItems: 1,
										MinItems: 1,
										Optional: true,
										Type:     schema.TypeList,
									},
									"node_pools": &schema.Schema{
										Description: "",
										Elem: &schema.Resource{Schema: map[string]*schema.Schema{
											"api_version": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"name": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"properties": &schema.Schema{
												Description: "",
												Elem: &schema.Resource{Schema: map[string]*schema.Schema{
													"availability_zones": &schema.Schema{
														Description: "",
														Elem:        &schema.Schema{Type: schema.TypeString},
														Optional:    true,
														Type:        schema.TypeList,
													},
													"count": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"enable_auto_scaling": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"enable_encryption_at_host": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"enable_fips": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"enable_node_public_ip": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"enable_ultra_ssd": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeBool,
													},
													"gpu_instance_profile": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"host_group_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"kubelet_config": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"allowed_unsafe_sysctls": &schema.Schema{
																Description: "",
																Elem:        &schema.Schema{Type: schema.TypeString},
																Optional:    true,
																Type:        schema.TypeList,
															},
															"container_log_max_files": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"container_log_max_size_mb": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"cpu_cfs_quota": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"cpu_cfs_quota_period": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"cpu_manager_policy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"fail_swap_on": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeBool,
															},
															"image_gc_high_threshold": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"image_gc_low_threshold": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"pod_max_pids": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"topology_manager_policy": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"kubelet_disk_type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"linux_os_config": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{
															"swap_file_size_mb": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeInt,
															},
															"sysctls": &schema.Schema{
																Description: "",
																Elem: &schema.Resource{Schema: map[string]*schema.Schema{
																	"fs_aio_max_nr": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"fs_file_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"fs_inotify_max_user_watches": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"fs_nr_open": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"kernel_threads_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_netdev_max_backlog": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_optmem_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_rmem_default": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_rmem_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_somaxconn": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_wmem_default": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_core_wmem_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_ip_local_port_range": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeString,
																	},
																	"net_ipv_4_neigh_default_gc_thresh_1": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_neigh_default_gc_thresh_2": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_neigh_default_gc_thresh_3": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_fin_timeout": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_keepalive_probes": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_keepalive_time": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_max_syn_backlog": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_max_tw_buckets": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_ipv_4_tcp_tw_reuse": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeBool,
																	},
																	"net_ipv_4_tcpkeepalive_intvl": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_netfilter_nf_conntrack_buckets": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"net_netfilter_nf_conntrack_max": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"vm_max_map_count": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"vm_swappiness": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																	"vm_vfs_cache_pressure": &schema.Schema{
																		Description: "",
																		Optional:    true,
																		Type:        schema.TypeInt,
																	},
																}},
																MaxItems: 1,
																MinItems: 1,
																Optional: true,
																Type:     schema.TypeList,
															},
															"transparent_huge_page_defrag": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
															"transparent_huge_page_enabled": &schema.Schema{
																Description: "",
																Optional:    true,
																Type:        schema.TypeString,
															},
														}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"max_count": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"max_pods": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"min_count": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"mode": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"node_labels": &schema.Schema{
														Description: "",
														Elem:        &schema.Schema{Type: schema.TypeString},
														Optional:    true,
														Type:        schema.TypeMap,
													},
													"node_public_ip_prefix_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"node_taints": &schema.Schema{
														Description: "",
														Elem:        &schema.Schema{Type: schema.TypeString},
														Optional:    true,
														Type:        schema.TypeList,
													},
													"orchestrator_version": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"os_disk_size_gb": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"os_disk_type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"os_sku": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"os_type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"pod_subnet_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"power_state": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{"code": &schema.Schema{
															Description: "",
															Optional:    true,
															Type:        schema.TypeString,
														}}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"proximity_placement_group_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"scale_down_mode": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"scale_set_eviction_policy": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"scale_set_priority": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"source_resource_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"spot_max_price": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeInt,
													},
													"tags": &schema.Schema{
														Description: "",
														Elem:        &schema.Schema{Type: schema.TypeString},
														Optional:    true,
														Type:        schema.TypeMap,
													},
													"type": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"upgrade_settings": &schema.Schema{
														Description: "",
														Elem: &schema.Resource{Schema: map[string]*schema.Schema{"max_surge": &schema.Schema{
															Description: "",
															Optional:    true,
															Type:        schema.TypeString,
														}}},
														MaxItems: 1,
														MinItems: 1,
														Optional: true,
														Type:     schema.TypeList,
													},
													"vm_size": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"vnet_subnet_id": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
													"workload_runtime": &schema.Schema{
														Description: "",
														Optional:    true,
														Type:        schema.TypeString,
													},
												}},
												MaxItems: 1,
												MinItems: 1,
												Optional: true,
												Type:     schema.TypeList,
											},
											"type": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
										}},
										MaxItems: 0,
										MinItems: 0,
										Optional: true,
										Type:     schema.TypeList,
									},
									"resource_group_name": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"subscription_id": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeString,
									},
								}},
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Type:     schema.TypeList,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"drift": &schema.Schema{
						Description: "Configuration for drift handling",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"action": &schema.Schema{
								Description: "flag to specify if sharing is enabled for resource",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"enabled": &schema.Schema{
								Description: "flag to specify if drift reconcillation is enabled for resource",
								Optional:    true,
								Type:        schema.TypeBool,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"proxy_config": &schema.Schema{
						Description: "The proxy configuration to be used for this cluster",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"allow_insecure_bootstrap": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeBool,
							},
							"bootstrap_ca": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"enabled": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeBool,
							},
							"http_proxy": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"https_proxy": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"no_proxy": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
							"proxy_auth": &schema.Schema{
								Description: "",
								Optional:    true,
								Type:        schema.TypeString,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"sharing": &schema.Schema{
						Description: "Configuration for sharing the resource",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"enabled": &schema.Schema{
								Description: "flag to specify if sharing is enabled for resource",
								Optional:    true,
								Type:        schema.TypeBool,
							},
							"projects": &schema.Schema{
								Description: "list of projects this resource is shared to",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{"name": &schema.Schema{
									Description: "name of the project",
									Optional:    true,
									Type:        schema.TypeString,
								}}},
								MaxItems: 0,
								MinItems: 0,
								Optional: true,
								Type:     schema.TypeList,
							},
						}},
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"system_components_placement": &schema.Schema{
						Description: "The system components for the placements of the workloads",
						Elem: &schema.Resource{Schema: map[string]*schema.Schema{
							"daemon_set_override": &schema.Schema{
								Description: "",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"node_selection_enabled": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeBool,
									},
									"tolerations": &schema.Schema{
										Description: "",
										Elem: &schema.Resource{Schema: map[string]*schema.Schema{
											"effect": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"key": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"operator": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeString,
											},
											"toleration_seconds": &schema.Schema{
												Description: "",
												Optional:    true,
												Type:        schema.TypeInt,
											},
											"value": &schema.Schema{
												Description: "",
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
								MaxItems: 1,
								MinItems: 1,
								Optional: true,
								Type:     schema.TypeList,
							},
							"node_selector": &schema.Schema{
								Description: "",
								Elem:        &schema.Schema{Type: schema.TypeString},
								Optional:    true,
								Type:        schema.TypeMap,
							},
							"tolerations": &schema.Schema{
								Description: "",
								Elem: &schema.Resource{Schema: map[string]*schema.Schema{
									"effect": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"key": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"operator": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeString,
									},
									"toleration_seconds": &schema.Schema{
										Description: "",
										Optional:    true,
										Type:        schema.TypeInt,
									},
									"value": &schema.Schema{
										Description: "",
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
						MaxItems: 1,
						MinItems: 1,
						Optional: true,
						Type:     schema.TypeList,
					},
					"type": &schema.Schema{
						Description: "The type of the cluster this spec corresponds to",
						Optional:    true,
						Type:        schema.TypeString,
					},
				}},
				MaxItems: 1,
				MinItems: 1,
				Optional: false,
				Type:     schema.TypeList,
			},
		},
	}

}

func resourceAKSClusterV3Import(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
}

func resourceAKSClusterV3Create(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
}

func resourceAKSClusterV3Update(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
}

func resourceAKSClusterV3Upsert(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
}

func resourceAKSClusterV3Delete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
}
