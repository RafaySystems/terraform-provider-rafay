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
    cloud_credentials = "my-gcp-credential"
    config {
      gcp_project           = "my-gcp-project-id"
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
        enable_vpc_nativetraffic = "true"
        max_pods_per_node        = 110
        # data_plane_v_2           = "ADVANCED_DATAPATH"
        # enable_data_plane_v_2_metrics = "true"
        # enable_data_plane_v_2_observability = "true"
        network_policy_config = "true"
        network_policy        = "CALICO"
        access {
          type = "public"
        }
        # firewall config for private cluster
        # access {
        #   type = "private"
        #   config {
        #     control_plane_ip_range                  = "172.16.3.0/28"
        #     enable_access_control_plane_external_ip = "true"
        #     enable_access_control_plane_global      = "true"
        #     disable_snat                            = "true"
        #     firewall_rules {
        #       action      = "allow"
        #       description = "allow traffic"
        #       direction   = "INGRESS"
        #       name        = "allow-ingress"
        #       priority    = 1000
        #       source_ranges = [
        #         "172.16.3.0/28"
        #       ]
        #       rules {
        #         ports    = ["22383", "9447"]
        #         protocol = "udp"
        #       }
        #     }
        #   }
        # }
      }
      features {
        enable_compute_engine_persistent_disk_csi_driver = "true"
      }
      node_pools {
        name         = "np"
        node_version = "1.26"
        size         = 1
        machine_config {
          machine_type   = "e2-standard-4"
          image_type     = "COS_CONTAINERD"
          boot_disk_type = "pd-standard"
          boot_disk_size = 100
          reservation_affinity {
            consume_reservation_type = "any"
          }

        }
        management {
          auto_upgrade = "true"
        }
        upgrade_settings {
          strategy = "SURGE"
          config {
            max_surge       = 0
            max_unavailable = 1
          }
        }
      }
    }
  }
}
