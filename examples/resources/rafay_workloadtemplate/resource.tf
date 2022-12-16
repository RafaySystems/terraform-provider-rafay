# Create workloadtemplate of Helm package type by uploading files from local system 
resource "rafay_workloadtemplate" "tftestworkloadtemplate1" {
  metadata {
    name    = "tftestworkloadtemplate1"
    project = "terraform"
  }
  spec {
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
    sharing {
      enabled = true
      projects {
        name = "tftestproject2"
      }
    }
  }
}

# Create workloadtemplate of Helm package type from Helm repo
resource "rafay_workloadtemplate" "tftestworkloadtemplate2" {
  metadata {
    name    = "tftestworkloadtemplate2"
    project = "terraform"
  }
  spec {
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

# Create workloadtemplate of Helm package type from git repo
resource "rafay_workloadtemplate" "tftestworkloadtemplate3" {
  metadata {
    name    = "tftestworkloadtemplate3"
    project = "terraform"
  }
  spec {
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

# Create a workloadtemplate of K8s Package type by uploading from local system 
resource "rafay_workloadtemplate" "tftestworkloadtemplate4" {
  metadata {
    name    = "tftestworkloadtemplate4"
    project = "terraform"
  }
  spec {
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

# Create a workload template of K8s Package type from git repo
resource "rafay_workloadtemplate" "tftestworkloadtemplate5" {
  metadata {
    name    = "tftestworkloadtemplate5"
    project = "terraform"
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "relative/path/to/some/chart.yaml"
        }
        repository = "git-user-repo-name"
        revision   = "branchname"
      }
    }
  }
}

# Create a workload template from catalog
resource "rafay_workloadtemplate" "tftestworkloadtemplate6" {
  metadata {
    name    = "tftestworkloadtemplate6"
    project = "terraform"
  }
  spec {
    artifact {
      type = "Helm"
      artifact{
        repository = "catalogName"
        chart_name = "chartName"
        chart_version = "chartVersion"
        values_paths {
          name = "file://relative/path/to/some/chart/values.yaml"
        }
      }
    }
  }
}