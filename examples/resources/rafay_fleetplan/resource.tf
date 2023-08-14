resource "rafay_fleetplan" "fp1" {
    metadata {
        name    = "fp1"
        project = "defaultproject"
    }
    spec {
        fleet {
            kind = "clusters"
            labels = {
                role = "qa"
            }

            projects {
                name = "defaultproject"
            }
        }

        operation_workflow {
            operations {
                name = "op1"
                action {
                    type        = "controlPlaneUpgrade"
                    description = "upgrading control plane"
                    control_plane_upgrade_config {
                        version = "1.24.10"
                    }
                    name = "action1"
                }
                prehooks {
                    name = "prehooks1"
                    description = "list all pods 10"
                    inject = ["KUBECONFIG"]
                    container_config {
                        runner {
                            type = "cluster"
                        }
                        image = "bitnami/kubectl"
                        arguments = ["get", "po", "-A"]
                    }
                }
                prehooks {
                    name = "prehooks2"
                    description = "list all pods 2"
                    inject = ["KUBECONFIG"]
                    container_config {
                        runner {
                            type = "cluster"
                        }
                        image = "bitnami/kubectl"	
                        arguments = ["get", "po", "-A"]
                    }
	    	    }
            }
            operations {
                name = "op2"
                action {
                    type        = "patch"
                    description = "upgrading control plane and nodegroup"
                    patch_config {
                        op = "replace"
                        path = ".spec.config.managedNodeGroups[0].maxSize"
			value = jsonencode(18)
                    }
		    patch_config {
                        op = "replace"
                        path = ".spec.blueprintConfig.name"
                        value = jsonencode("minimal")
                    }
                    name = "action2"
                }
                posthooks {
                    name = "posthooks1"
                    description = "list all pods 1"
                    inject = ["KUBECONFIG"]
                    container_config {
                        runner {
                            type = "cluster"
                        }
                        image = "bitnami/kubectl"
                        arguments = ["get", "po", "-A"]
                    }
                }
            }
            operations {
                name = "op3"
                action {
                    name = "action3"
                    type = "nodeGroupsUpgrade"
                    description = "upgrading nodegroup"
                    node_groups_upgrade_config {
                        version = "1.24.10"
                        names = ["ng1", "ng2"]
                    }
                }
            }
        }

        agents {
            name = "kalyan-docker"
        }
    }
}

