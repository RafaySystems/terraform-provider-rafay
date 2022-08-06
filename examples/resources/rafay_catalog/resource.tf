resource "rafay_catalog" "basic_custom_catalog" {
  metadata {
    name = "terraform-test"
    project = "terraform"
  }
  spec {
    auto_sync = false
    repository = "istio-terraform"
    type = "HelmRepository"
    sharing {
      enabled = true
      projects {
        name = "defaultproject" 
      }
    }      
  }
}
