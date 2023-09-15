resource "rafay_gke_cluster" "tf-example" {
  metadata {
    name    = var.name
    project = "defaultproject"
  }
  spec {
    type = "gke"
    blueprint {
      name    = "minimal"
      version = "latest"
    }
    cloud_credentials = "-dev-"
    config {
      gcp_project           = "dev-382813"
      control_plane_version = "1.26"
      location {
        type = "zonal"
        config {
          zone = "us-central1-c"
        }
      }
      network {
        name                     = "default"
        subnet_name              = "default"
        # we can configure shared VPC as well
        # name = "projects/kr-test-200723/global/networks/km-1"
        # subnet_name = "projects/kr-test-200723/regions/us-central1/subnetworks/km-1"
        enable_vpc_nativetraffic = "true"
        max_pods_per_node        = 110
        # configure below values when using shared VPC
        # pod_secondary_range_name = "pod"
        # service_secondary_range_name = "service"
        access {
          type = "public"
        }
      }
      features {
        enable_compute_engine_persistent_disk_csi_driver = "true"
      }
      node_pools {
        name         = "np"
        node_version = "1.26"
        size         = 3
        machine_config {
          machine_type   = "e2-standard-4"
          image_type     = "COS_CONTAINERD"
          boot_disk_type = "pd-standard"
          boot_disk_size = 100
        }
        management {
          auto_upgrade = "true"
        }
        upgrade_settings {
          strategy = "SURGE"
          surge_settings {
            max_surge       = 0
            max_unavailable = 1
          }
        }
      }
    }
  }
}