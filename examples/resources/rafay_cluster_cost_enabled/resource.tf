# Example Cluster Cost Enabled Policy for All Clusters in All projects
resource "rafay_cluster_cost_enabled" "tfdemoclustercostenabledpolicy1" {
  metadata {
    name = "tfdemoclustercostenabledpolicy1"
  }
  spec {
    selection_type = "allClusters"
    policy_project = "*"
  
  }
}

# Example Cluster Cost Enabled Policy Policy for All Clusters in a specific project
resource "rafay_cluster_cost_enabled" "tfdemoclustercostenabledpolicy2" {
  metadata {
    name = "tfdemoclustercostenabledpolicy2"
  }
  spec {
    selection_type = "allClusters"
    policy_project = "project-1"
   
  }
}

# Example Cluster Cost Enabled Policy Policy for Specific Clusters in a specific project
resource "rafay_cluster_cost_enabled" "tfdemoclustercostenabledpolicy3" {
  metadata {
    name = "tfdemoclustercostenabledpolicy3"
  }
  spec {
    selection_type = "clusterNames"
    policy_project = "project-1"
    clusters = [
      "cluster-1", "cluster-2"
    ]
    
  }
}

# Example Cluster Cost Enabled Policy Policy for Specific Cluster labels in a specific project
resource "rafay_cluster_cost_enabled" "tfdemoclustercostenabledpolicy4" {
  metadata {
    name = "tfdemoclustercostenabledpolicy4"
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
  }
}

