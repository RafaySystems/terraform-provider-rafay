resource "rafay_alertconfig" "basic_rafay_alertconfig" {
  metadata {
    name = "test-alertconfig-terraform-1"
    project = "defaultproject"
  }
  spec {
    alerts = {
      "cluster" = true,
      "pod" = true,
      "pvc" = false,
      "node" = true,
      "agent" = true,
    }
    emails = [
      "nilay@rafay.co",
      "nilay+project1@rafay.co"
    ]
  }
}
