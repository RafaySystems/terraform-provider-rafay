resource "rafay_organizationalertconfig" "basic_rafay_organizationalertconfig" {
  metadata {
    name = "test-organizationalertconfig-terraform-1"
  }
  spec {
    alerts {
      cluster      = false
      pod          = true
      pvc          = true
      node         = true
      agent_health = true
    }
    emails = [
      "nilay+org1@rafay.co",
      "nilay+org2@rafay.co"
    ]
  }
}