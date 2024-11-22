resource "rafay_cost_optimisation" "tfdemocostoptimisation" {
  metadata {
    name        = "tfdemocostoptimisation"
  }
  spec {
    selection_type = "clusternames"
    config_project = "defaultproject"
    clusters = [
      "c1",
      "c2"
    ]
    cluster_labels {
      key       = "k1"
      value     = "v1"
    }
    inclusions {
      namespace = "ns1"
      namespace_label = [
        "k1",
        "v1"
      ]
    }
    exclusions {
      namespace = "ns2"
      namespace_label = [
        "k2",
        "v2"
      ]
    }
    period = 7
    mode = "dryrun"
    recommended = 20
    bound {
      cpu {
        minimum = "0.5"
        maximum = "4"
      }
      memory {
        minimum = "128"
        maximum = "4096"
      }
    }
    min_threshold {
      cpu {
        percentage = "10"
        unit = "0.1"
      }
      memory {
        percentage = "10"
        unit = "0.1"
      }
    }
  }
}
