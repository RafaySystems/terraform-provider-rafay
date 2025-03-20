resource "rafay_driver" "driver" {
  metadata {
    name    = var.name
    project = var.project
  }
  spec {
    config {
      type            = var.type
      timeout_seconds = 100
      max_retry_count = 3
      container {
        image     = var.image
        arguments = ["--log-level=3"]
        commands  = ["run main.go"]
        image_pull_credentials {
          password = "gibberesh"
          registry = "hub.docker.io"
          username = "randomuser"
        }
        kube_config_options {
          kube_config    = "path/to/kubeconfig.json"
          out_of_cluster = true
        }
        kube_options {
          labels = {
            "name" : "terraform"
          }
          namespace = "default"
          resources = ["pods", "deployments"]
        }
        volumes {
          use_pvc {
            value = true
          }
          mount_path = "/tmp/var"
          pvc_size_gb = "20"
          pvc_storage_class = "hdb"
        }
      }
    }
  }
}