resource "rafay_project" "tfdemoproject1" {
  metadata {
    name        = "tfdemoproject1"
    description = "tfdemoproject1 description"
  }

  spec {
    # spec default value is fixed to 'false' foor now in the controller.
    # Will be allowed to enable in the future.
    default = false
    cluster_resource_quota {
      requests {
        cpu {
          string = "4"
        }
        memory {
          string = "8Gi"
        }
      }
  }
}
