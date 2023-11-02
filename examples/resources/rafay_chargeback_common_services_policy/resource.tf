# Example Chargeback Common Services Policy for All Clusters in All projects
resource "rafay_chargeback_common_services_policy" "tfdemocommonservicespolicy1" {
  metadata {
    name = "tfdemocommonservicespolicy1"
  }
  spec {
    selection_type = "allClusters"
    policy_project = "*"
    common_services_namespaces = [
      "ns-1", "ns-2"
    ]
    common_services_namespace_labels {
      key   = "ns-label-key-1"
      value = "ns-label-value-1"
    }
    common_services_namespace_labels {
      key   = "ns-label-key-2"
      value = "ns-label-value-2"
    }
  }
}

# Example Chargeback Common Services Policy for All Clusters in a specific project
resource "rafay_chargeback_common_services_policy" "tfdemocommonservicespolicy2" {
  metadata {
    name = "tfdemocommonservicespolicy2"
  }
  spec {
    selection_type = "allClusters"
    policy_project = "project-1"
    common_services_namespaces = [
      "ns-1", "ns-2"
    ]
    common_services_namespace_labels {
      key   = "ns-label-key-1"
      value = "ns-label-value-1"
    }
    common_services_namespace_labels {
      key   = "ns-label-key-2"
      value = "ns-label-value-2"
    }
  }
}

# Example Chargeback Common Services Policy for Specific Clusters in a specific project
resource "rafay_chargeback_common_services_policy" "tfdemocommonservicespolicy3" {
  metadata {
    name = "tfdemocommonservicespolicy3"
  }
  spec {
    selection_type = "clusterNames"
    policy_project = "project-1"
    clusters = [
      "cluster-1", "cluster-2"
    ]
    common_services_namespaces = [
      "ns-1", "ns-2"
    ]
    common_services_namespace_labels {
      key   = "ns-label-key-1"
      value = "ns-label-value-1"
    }
    common_services_namespace_labels {
      key   = "ns-label-key-2"
      value = "ns-label-value-2"
    }
  }
}

# Example Chargeback Common Services Policy for Specific Cluster labels in a specific project
resource "rafay_chargeback_common_services_policy" "tfdemocommonservicespolicy4" {
  metadata {
    name = "tfdemocommonservicespolicy4"
  }
  spec {
    selection_type = "clusterLabels"
    policy_project = "project-1"
    cluster_labels {
      key   = "cluster-label-key-1"
      value = "cluster-label-value-1"
    }
    cluster_labels {
      key   = "cluster-label-key-2"
      value = "cluster-label-value-2"
    }
    common_services_namespaces = [
      "ns-1", "ns-2"
    ]
    common_services_namespace_labels {
      key   = "ns-label-key-1"
      value = "ns-label-value-1"
    }
    common_services_namespace_labels {
      key   = "ns-label-key-2"
      value = "ns-label-value-2"
    }
  }
}
