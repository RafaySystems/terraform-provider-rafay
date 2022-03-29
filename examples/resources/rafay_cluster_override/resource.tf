resource "rafay_cluster_override" "tfdemocluster-override1" {
  metadata {
    name    = "tfdemocluster-override1"
    project = "upgrade"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector  = "env=test"
    resource_selector = "rafay.dev/name=override-workload"
    type              = "ClusterOverrideTypeWorkload"
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
    
    cluster_placement {
      placement_type = "ClusterLabels"
      cluster_labels {
        key = "rafay.dev/clusterName"
        value = "hardik-terraform4"
      }
    }
  }
}


resource "rafay_cluster_override" "tfdemocluster-override2" {
  metadata {
    name    = "tfdemocluster-override2"
    project = "upgrade"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector  = "rafay.dev/clusterName in (my-cluster-name)"
    resource_selector = "rafay.dev/name=aws-lb-controller"
    type              = "ClusterOverrideTypeAddon"
    override_values   = <<-EOS
    clusterName: my-cluster-name1
    EOS
  }
}


resource "rafay_cluster_override" "tfdemocluster-override3" {
  metadata {
    name    = "tfdemocluster-override3"
    project = "upgrade"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "valuesFile"
    }
  }
  spec {
    cluster_selector  = "rafay.dev/clusterName in (my-cluster-name)"
    resource_selector = "rafay.dev/name=aws-lb-controller"
    type              = "ClusterOverrideTypeAddon"
    value_repo_ref    = "release-check-ssh"
    values_repo_artifact_meta {
      git_options {
        revision = "main"
        repo_artifact_files {
          name          = "dev.yml"
          relative_path = "yaml/dev.yml"
          file_type     = "FileTypeNotSet"
        }
      }
    }
  }
}
