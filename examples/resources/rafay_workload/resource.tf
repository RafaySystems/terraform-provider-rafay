# Create workload of Helm package type by uploading files from local system 
resource "rafay_workload" "tftestworkload1" {
  metadata {
    name    = "tftestworkload1"
    project = "terraform"
  }
  spec {
    namespace = "test-workload1"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "file://relative/path/to/some/chart.tgz"
        }
        values_paths {
          name = "file://relative/path/to/some/chart.yaml"
        }
      }
    }
  }
}

# Create workload of Helm package type from Helm repo
resource "rafay_workload" "tftestworkload2" {
  metadata {
    name    = "tftestworkload2"
    project = "terraform"
  }
  spec {
    namespace = "test-workload2"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Helm"
      artifact{
        values_paths {
          name = "file://relative/path/to/some/chart/values.yaml"
        }
        repository = "helm-repo-name"
        chart_name = "chartname"
        chart_version = "versionID"
      }
    }
  }
}

# Create workload of Helm package type from git repo
resource "rafay_workload" "tftestworkload3" {
  metadata {
    name    = "tftestworkload3"
    project = "terraform"
  }
  spec {
    namespace = "test-workload3"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "relative/path/to/some/chart.tgz"
        }
        repository = "git-user-repo-name"
        revision   = "branchname"
      }
    }
  }
}

# Create a workload of K8s type by uploading from local system 
resource "rafay_workload" "tftestworkload4" {
  metadata {
    name    = "tftestworkload4"
    project = "terraform"
  }
  spec {
    namespace = "test-workload4"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://relative/path/to/some/chart.yaml"
        }
      }
    }
  }
}

# Create workload of K8s Yaml type from git repo
resource "rafay_workload" "tftestworkload5" {
  metadata {
    name    = "tftestworkload5"
    project = "terraform"
  }
  spec {
    namespace = "test-workload5"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    drift {
      action  = "Deny"
      enabled = true
    }
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "yaml/workload.yaml"
        }
        repository = "release-check-ssh"
        revision   = "main"
      }
    }
  }
}