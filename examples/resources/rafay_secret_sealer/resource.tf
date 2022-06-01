resource "rafay_secretsealer" "tfdemoproject1" {
  metadata {
    name    = "test"
    project = "bharath"
  }
  spec {
    type = "KubeSeal"
    version = "v-3"
  }
}