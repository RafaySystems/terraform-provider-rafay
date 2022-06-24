# Create workloadtemplate of Helm package type by uploading files from local system 
resource "rafay_workloadtemplate" "tfdemoworkloadtemplate1" {
  metadata {
    name    = "tfdemoworkloadtemplate1"
    project = "tfdemoproject"
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
        name = "tfdemoproject2"
      }
    }
  }
}

# Create workloadtemplate of Helm package type from Helm repo
resource "rafay_workloadtemplate" "tfdemoworkloadtemplate2" {
  metadata {
    name    = "tfdemoworkloadtemplate2"
    project = "tfdemoproject"
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
resource "rafay_workloadtemplate" "tfdemoworkloadtemplate3" {
  metadata {
    name    = "tfdemoworkloadtemplate3"
    project = "tfdemoproject"
  }
  spec {
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "relative/path/to/some/chart.yaml"
        }
        repository = "git-user-repo-name"
        revision   = "branchname"
      }
    }
  }
}

# Create a workloadtemplate of K8s Package type by uploading from local system 
resource "rafay_workloadtemplate" "tfdemoworkloadtemplate4" {
  metadata {
    name    = "tfdemoworkloadtemplate4"
    project = "tfdemoproject"
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
resource "rafay_workloadtemplate" "tfdemoworkloadtemplate5" {
  metadata {
    name    = "tfdemoworkloadtemplate5"
    project = "tfdemoproject"
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
