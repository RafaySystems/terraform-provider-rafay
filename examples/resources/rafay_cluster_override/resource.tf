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

// YAML BASED cluster overrides

resource "rafay_cluster_override" "tfdemocluster-yamloverride1" {
  metadata {
    name    = "tfdemocluster-yamloverride1"
    project = "tfdemoproject1"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "manifestsFile"
    }
  }
  spec {
    artifact_type = "NativeYAML"
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
      apiVersion: apps/v1
      kind: Deployment
      metadata:
        name: nginx
      patch:
      - op: replace
        path: /spec/replica
        value: 3
    EOS
  }
}

resource "rafay_cluster_override" "tfdemocluster-override-share1" {
  metadata {
    name    = "tfdemocluster-override-share-1"
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
    sharing {
      enabled = true  // set false to unshare from all projects
      projects {
        name = "project1"
      }
      projects {
        name = "project2"
      }
    }
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

resource "rafay_cluster_override" "tfdemocluster-clusterquotaoverride1" {
  metadata {
    name    = "tfdemocluster-clusterquotaoverride1"
    project = "terraform"
    labels = {
      "rafay.dev/overrideScope" = "clusterLabels"
      "rafay.dev/overrideType"  = "manifestsFile"
    }
  }
  spec {
    artifact_type = "NativeYAML" // NativeYAML or GitRepoWithNativeYAML
    cluster_selector  = "rafay.dev/clusterName in (cluster-1)"
    cluster_placement {
      placement_type = "ClusterSpecific"
      cluster_labels {
        key = "rafay.dev/clusterName"
        value = "cluster-1"
      }
    }
    resource_selector = "rafay.dev/system=true,rafay.dev/component=cluster-resource-quota"
    type              = "ClusterOverrideTypeClusterQuota"
    override_values   = <<-EOS
      apiVersion: system.k8smgmt.io/v3
      kind: Project
      metadata:
        name: terraform
      patch:
      - op: replace
        path: /spec/clusterResourceQuota/cpuLimits
        value: 30m
    EOS
  }
}