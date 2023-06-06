resource "rafay_gke_cluster_v3" "pv-gke-terraform" {
  metadata {
    name    = "pv-gke-tf-1"
    project = "defaultproject"
  }
  spec {
    type          = "gke"
    blueprint_config {
      name = "default"
    }
    cloud_credentials = "dev"
    
  }
}