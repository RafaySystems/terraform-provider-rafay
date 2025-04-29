resource "rafay_alertconfig" "basic_rafay_alertconfig" {
  metadata {
    name    = "test-alertconfig-terraform-1"
    project = "defaultproject"
  }
  spec {
    alerts {
      cluster      = true
      pod          = true
      pvc          = true
      node         = false
      agent_health = false
    }
    emails = [
      "nilay@rafay.co",
      "nilay+project1@rafay.co"
    ]
  }
}
