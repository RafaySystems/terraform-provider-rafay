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
                        runner = "cluster"
                        image = "bitnami/kubectl"
                        arguments = ["get", "po", "-A"]
                    }
                }
                prehooks {
                    name = "prehooks2"
                    description = "list all pods 2"
                    inject = ["KUBECONFIG"]
                    container_config {
                        runner = "cluster"
                        image = "bitnami/kubectl"	
                        arguments = ["get", "po", "-A"]
                    }
	    	    }
            }
            operations {
                name = "op2"
                action {
                    type        = "nodeGroupsUpgrade"
                    description = "upgrading control plane and nodegroup"
                    node_groups_upgrade_config {
                        version = "1.24.10"
                        names = ["ng1", "ng2"]
                    }
                    name = "action2"
                }
                posthooks {
                    name = "posthooks1"
                    description = "list all pods 1"
                    inject = ["KUBECONFIG"]
                    container_config {
                        runner = "cluster"
                        image = "bitnami/kubectl"
                        arguments = ["get", "po", "-A"]
                    }
                }
            }
        }

        agents {
            name = "kalyan-docker"
        }
    }
}

