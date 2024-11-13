terraform {
  required_providers {
    rafay = {
      version = "1.1.28"
      source  = "rafay/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = "/Users/vihari/Downloads/rafay-org-vihari@rafay.co.json"
}


resource "rafay_mks_cluster" "mks-sample-cluster" {
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"
  metadata = {
    name    = "vihari-mks-tf-cluster"
    project = "defaultproject"
  }
  spec = {
    blueprint = {
      name = "minimal"
    }
    config = {
      auto_approve_nodes      = true
      dedicated_control_plane = false
      kubernetes_version      = "v1.30.4"
      installer_ttl           = 365
      kubelet_extra_args      = {
        "max-pods" = "900"
      }
      kubernetes_upgrade = {
        strategy = "sequential"
        params = {
          worker_concurrency = "50%"
        }
      }
      network = {
        cni = {
          name    = "Calico"
          version = "3.26.1"
        }
        pod_subnet     = "10.244.0.0/16"
        service_subnet = "10.96.0.0/12"
      }
      nodes = {
        "vih-a4" = {
          arch             = "amd64"
          hostname         = "vih-a4"
          operating_system = "Ubuntu22.04"
          private_ip       = "10.0.0.136"
          roles            = ["ControlPlane", "Worker", "Storage"]
          ssh = {
            ip_address = "129.146.58.186"
            port = "22"
            private_key_path = "/Users/vihari/.ssh/vihari_oci_ssh"
            username = "ubuntu"
          }
        },
        "vih-a5" = {
          arch             = "amd64"
          hostname         = "vih-a5"
          operating_system = "Ubuntu22.04"
          kubelet_extra_args = {
            "max-pods" = "600"
          }
          private_ip       = "10.0.0.183"
          roles            = ["Worker", "Storage"]
          ssh = {
            ip_address = "129.146.6.223"
            port = "22"
            private_key_path = "/Users/vihari/.ssh/vihari_oci_ssh"
            username = "ubuntu"
          }
        }
      }
      cluster_ssh = {
        port = "22"
        private_key_path = "/Users/vihari/.ssh/vihari_oci_ssh"
        username = "ubuntu"
      }
        # "hostname2" = {
        #   arch             = "amd64"
        #   hostname         = "hostname2"
        #   operating_system = "Ubuntu22.04"
        #   private_ip       = "10.12.114.59"
        #   roles            = ["Worker"]
        #   labels = {
        #     "app"   = "infra"
        #     "infra" = "true"
        #   }
        #   taints = [
        #     {
        #       effect = "NoSchedule"
        #       key    = "infra"
        #       value  = "true"
        #     },
        #     {
        #       effect = "NoSchedule"
        #       key    = "app"
        #       value  = "infra"
        #     },
        #   ]
        # }
      
    # system_components_placement = {
    #   node_selector = {
    #     "app"   = "infra"
    #     "infra" = "true"
    #   }
    #   tolerations = [
    #     {
    #       effect   = "NoSchedule"
    #       key      = "infra"
    #       operator = "Equal"
    #       value    = "true"
    #     },
    #     {
    #       effect   = "NoSchedule"
    #       key      = "app"
    #       operator = "Equal"
    #       value    = "infra"
    #     },
    #     {
    #       effect   = "NoSchedule"
    #       key      = "app"
    #       operator = "Equal"
    #       value    = "platform"
    #     },
    #   ]
    # }
    }
  }
}
