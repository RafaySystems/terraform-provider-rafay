
terraform {
  required_providers {
    rafay = {
      version = ">= 1.1.34"
      source = "registry.terraform.io/RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = "/Users/vasu/.rafay/cli/config.json"
}


resource "rafay_mks_cluster" "mks-tf-noha-101" {
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"

  metadata = {
    annotations = {
      "key2" = "value2"
    }
    description  = "This is a sample MKS cluster."
    display_name = "mks-tf-noha-101"
    labels = {
      "env" = "development"
    }
    name    = "mks-tf-noha-101"
    project = "terraform"
  }

  spec = {
    blueprint = {
      name    = "minimal"
    }
    cloud_credentials = "vasu-mks-ssh-010"
    config = {
      auto_approve_nodes     = true
      dedicated_control_plane = false
      high_availability       = false
      kubernetes_version      = "v1.28.9"   
      kubernetes_upgrade = {
        strategy = "sequential"
        params = {
          worker_concurrency = "50%"
        }
      }

      location = "mumbai-in"
      network = {
        cni = {
          name    = "Calico"
          version = "3.26.1"
        }
        pod_subnet     = "10.244.0.0/16"
        service_subnet = "10.96.0.0/12"
      }

      nodes = [ 
        {
        arch            = "amd64"
        hostname        = "vasu-mks-tf-test-cp"
        operating_system = "Ubuntu22.04"
        private_ip      = "10.12.71.72"
        roles           = ["ControlPlane", "Worker"]

        labels =  {
          "app"   = "infra"
          "infra" = "true"
        }

        taints = [
          {
            effect = "NoSchedule"
            key    = "app"
            value  = "infra"
          },
          {
            effect = "NoSchedule"
            key = "infra"
            value = "true"
          }
        ]
       }
      ]
    }
    system_components_placement = {
      daemon_set_override = {
        daemon_set_tolerations = [ 
          {
            effect             = "NoSchedule"
            key                = "app"
            operator           = "Equal"
            value              = "infra"
          },
          {
            effect             = "NoSchedule"
            key                = "infra"
            operator           = "Equal"
            value              = "true"
          },
      ]
        node_selection_enabled = true
      }
      node_selector = {
        "app" = "infra"
        "infra" = "true"
      }
      tolerations = [
        {
          effect   = "NoSchedule"
          key      = "app"
          operator = "Equal"
          value    = "infra"
        },
        {
          effect   = "NoSchedule"
          key      = "infra"
          operator = "Equal"
          value    = "true"
        },
      ]
    }
    type = "mks"
  }
}

