resource "rafay_gke_cluster_v3" "pv-gke-terraform" {
  metadata {
    name    = "pv-gke-tf-1"
    project = "defaultproject"
  }
  spec {
    type          = "Gke"
    blueprint {
      name = "minimal"
      version = "latest"
    }
    cloud_credentials = "pv-gke-dev-cred"
    config {
      gcp_project = "dev-382813"
      control_plane_version = "1.25"
      location {
        type = "zonal"
        config {
          zone = "us-west1-c"
        }
      }
      network {
        name = "default"
        subnet_name = "default"
        enable_vpc_nativetraffic = "true"
        max_pods_per_node = 110
        access {
          type = "public"
        }
      }
      node_pools {
        name = "np"
        node_version = "1.24"
        size = 3
        machine_config {
          machine_type = "e2-standard-4"
          image_type = "COS_CONTAINERD"
          boot_disk_type = "pd-standard"
          boot_disk_size = 100
        }
      }
    }
  }
}