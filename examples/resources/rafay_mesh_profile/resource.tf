resource "rafay_mesh_profile" "tfdemomeshprofile1" {
  metadata {
    name    = "tfdemomeshprofile1"
    project = "terraform"
  }
  spec {
    version = "v0"
    sharing {
        enabled = false
    }
  }
}

#Example profile with installation params
resource "rafay_mesh_profile" "tfdemomeshprofile1" {
  metadata {
    name    = "tfdemomeshprofile-ip"
    project = "terraform"
  }
  spec {
    version = "v0"
    installation_params {
      cert_type = 0
      enable_ingress = false
      enable_namespaces_by_default = false
      resource_quota {
        cpu_requests = "500m"
        memory_requests = "2048Mi"
      }
    }
  }
}
