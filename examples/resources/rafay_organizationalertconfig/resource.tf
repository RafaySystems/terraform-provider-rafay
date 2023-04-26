resource "rafay_organizationalertconfig" "basic_rafay_organizationalertconfig" {
  metadata {
    name = "test-organizationalertconfig-terraform-1"
  }
  spec {
    alerts = {
      "cluster" = true,
      "pod" = true,
      "pvc" = false,
      "node" = false,
      "agentHealth" = false,
    }
    emails = [
      "nilay+org1@rafay.co",
      "nilay+org2@rafay.co"
    ]
  }
}
