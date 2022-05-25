#Kubernetes agent

resource "rafay_agent" "tfdemoagent1" {
  metadata {
    name    = "tfdemoagent1"
    project = "terraform"
  }
  spec {
        type = "ClusterAgent"
        cluster {
            name = "dev-test"
        }
        active = true
  }
}

#Docker agent
resource "rafay_agent" "tfdemoagent2" {
  metadata {
    name    = "tfdemoagent2"
    project = "terraform"
  }
  spec {
        type = "Docker"
        active = true
  }
}
