terraform {
  required_providers {
    rafay = {
      version = ">= 1.1.35"
      source  = "registry.terraform.io/RafaySystems/rafay"
    }
  }
}

provider "rafay" {
  provider_config_file = "/Users/user/.rafay/cli/config.json"
}


resource "rafay_mks_cluster" "mks-sample-cluster" {
  api_version = "infra.k8smgmt.io/v3"
  kind        = "Cluster"
  metadata = {
    name    = "mks-sample-cluster"
    project = "terraform"
  }
  spec = {
    blueprint = {
      name = "minimal"
    }
    cloud_credentials = "mks-ssh-creds"
    config = {
      auto_approve_nodes      = true
      dedicated_control_plane = false
      kubernetes_version      = "v1.28.9"
      installer_ttl           = 365
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
      },
      cluster_ssh = {
        username         = "ubuntu"
        port             = "22"
        private_key_path = "/path/to/ssh/private_key"
      },
      kubelet_extra_args = {
        "max-pods": "250"
      },
      kubelet_configuration_overrides = <<-EOF
        maxPods: 150
        evictionHard:
          imagefs.available: "25%"
          memory.available: "300Mi"
          nodefs.available: "20%"
          nodefs.inodesFree: "15%"
        EOF
      nodes = {
        "hostname1" = {
          arch             = "amd64"
          hostname         = "hostname1"
          operating_system = "Ubuntu22.04"
          private_ip       = "10.12.25.234"
          roles            = ["ControlPlane", "Worker"]
          labels = {
            "app"   = "infra"
            "infra" = "true"
          },
          kubelet_extra_args = {
            "max-pods": "600"
          },
          kubelet_configuration_overrides = <<-EOF
              maxPods: 150
              evictionHard:
                imagefs.available: "25%"
                memory.available: "300Mi"
                nodefs.available: "20%"
                nodefs.inodesFree: "15%"
              EOF
        },
        "hostname2" = {
          arch             = "amd64"
          hostname         = "hostname2"
          operating_system = "Ubuntu22.04"
          private_ip       = "10.12.114.59"
          roles            = ["Worker"]
          labels = {
            "app"   = "infra"
            "infra" = "true"
          }
          taints = [
            {
              effect = "NoSchedule"
              key    = "infra"
              value  = "true"
            },
            {
              effect = "NoSchedule"
              key    = "app"
              value  = "infra"
            },
          ]
        }
      }
    }
    system_components_placement = {
      node_selector = {
        "app"   = "infra"
        "infra" = "true"
      }
      tolerations = [
        {
          effect   = "NoSchedule"
          key      = "infra"
          operator = "Equal"
          value    = "true"
        },
        {
          effect   = "NoSchedule"
          key      = "app"
          operator = "Equal"
          value    = "infra"
        },
        {
          effect   = "NoSchedule"
          key      = "app"
          operator = "Equal"
          value    = "platform"
        },
      ]
    }
  }
}
