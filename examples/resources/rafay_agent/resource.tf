resource "rafay_agent" "tfdemoagent1" {
  metadata {
    name    = "tfdemoagent1"
    project = "upgrade"
  }
  spec {
        type = "ClusterAgent"
        cluster {
            name = "dev-test"
        }
        active = true
  }
}
