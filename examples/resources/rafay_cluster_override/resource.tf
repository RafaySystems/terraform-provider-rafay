resource "rafay_cluster_override" "tfdemocluster-override1" {
  metadata {
    name    = "tfdemocluster-override1"
    project = "terraform"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector  = "rafay.dev/clusterName in (cluster-1)"
    cluster_placement {
      placement_type = "ClusterSpecific"
      cluster_labels {
        key = "rafay.dev/clusterName"
        value = "cluster-1"
      }
    }
    resource_selector = "rafay.dev/name=override-addon"
    type              = "ClusterOverrideTypeAddon"
    override_values   = <<-EOS
    replicaCount: 1
    image:
      repository: nginx
      pullPolicy: Always
      tag: "1.19.8"
    service:
      type: ClusterIP
      port: 8080
    EOS
  }
}


resource "rafay_cluster_override" "tfdemocluster-override2" {
  metadata {
    name    = "tfdemocluster-override2"
    project = "terraform"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector  = "key in (value)"
    cluster_placement {
      placement_type = "ClusterLabels"
      cluster_labels {
        key = "key"
        value = "value"
      }
    }
    resource_selector = "rafay.dev/name=aws-lb-controller"
    type              = "ClusterOverrideTypeAddon"
    value_repo_ref    = "git-repo-name"
    values_repo_artifact_meta {
      git_options {
        revision = "main"
        repo_artifact_files {
          name          = "overrides.yaml"
          relative_path = "yaml/overrides.yaml"
          file_type     = "FileTypeNotSet"
        }
      }
    }
  }
}
