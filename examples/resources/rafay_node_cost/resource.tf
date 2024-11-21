# Example Node Cost Policy
resource "rafay_node_cost" "tfdemonodecostespolicy1" {
  metadata {
    name = "tfdemonodecostpolicy1"
  }
  spec {
    node_labels {
      key   = "label-key-1"
      value = "value-1"
    }
    node_labels {
      key   = "label-key-2"
      value = "value-2"
    }
     cost_values {
        cpu = "2.5"
        gpu = "3.61"
        memory = "4.3"
      }
  }
}
