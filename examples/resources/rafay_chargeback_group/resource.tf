# Example Chargeback Group resource with Summary Type
resource "rafay_chargeback_group" "tfdemosummarycbgroup1" {
  metadata {
    name = "tfdemosummarycbgroup1"
  }
  spec {
    type = "summary"
    inclusions {
      project = "project-1"
    }
    aggregate {
      project = false
      cluster = false
      namespace = false
      label = [
        "test-1"
      ]
    }
  }
}

# Example Chargeback Group resource with Detailed Type
resource "rafay_chargeback_group" "tfdemodetailedcbgroup1" {
  metadata {
    name = "tfdemodetailedcbgroup1"
  }
  spec {
    type = "detailed"
    inclusions {
      project = "project-1"
      label = [
         "test-1", "test-2"
      ]
    }
    exclusions {
      cluster = "eks-cluster-1"
    }
  }
}
