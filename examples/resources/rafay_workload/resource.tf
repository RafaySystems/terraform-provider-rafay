# Create workload of Helm package type by uploading files from local system 
resource "rafay_workload" "tftestworkload1" {
  metadata {
    name    = "tftestworkload1"
    project = "terraform"
  }
  spec {
    namespace = "test-workload1"
    version   = "v1"
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
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Helm"
      artifact {
        values_paths {
          name = "file://relative/path/to/some/chart/values.yaml"
        }
        repository    = "helm-repo-name"
        chart_name    = "chartname"
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
    version   = "v1"
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
    version   = "v1"
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
    version   = "v1"
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


# Create Helm workload from Git repo. Chart & default values from one repo, override values from another repo
resource "rafay_workload" "tftestworkload6" {
  metadata {
    name    = "tftestworkload6"
    project = "terraform"
  }
  spec {
    namespace = "test-workload6"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Helm"
      artifact {
        repository = "test-repo1"
        revision   = "main"
        chart_path {
          name = "chart/path/chart.tgz"
        }
        #default value from same repo as chart
        values_paths {
          name = "value/path/values.yaml"
        }
        #override value from another repo
        values_ref {
          repository = "test-repo2"
          revision   = "main"
          values_paths {
            name = "value/path/values.yaml"
          }
        }
      }
    }
  }
}

# Create a workload from web URL
resource "rafay_workload" "tftestworkload8" {
  metadata {
    name    = "tftestworkload8"
    project = "terraform"
  }
  spec {
    namespace = "test-workload5"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-1"
    }
    artifact {
      type = "Yaml"
      artifact {
        url = ["https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/nginx-app.yaml"]
      }
    }
  }
}

# Create a Helm4 workload by uploading a chart
resource "rafay_workload" "helm4_upload" {
  metadata {
    name    = "helm4-upload-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    artifact {
      type = "Helm4"
      artifact {
        chart_path {
          name = "file://relative/path/to/chart.tgz"
        }
        values_paths {
          name = "file://relative/path/to/values.yaml"
        }
      }
      options {
        set                 = ["replicaCount=3"]
        set_string          = ["image.tag=latest"]
        wait_strategy       = "watcher"
        wait_for_jobs       = true
        timeout             = "5m0s"
        max_history         = 10
        cleanup_on_fail     = true
        rollback_on_failure = true
      }
    }
  }
}

# Create a Helm4 workload from a Helm repository
resource "rafay_workload" "helm4_helm_repository" {
  metadata {
    name    = "helm4-helm-repo-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    artifact {
      type = "Helm4"
      artifact {
        repository    = "default-bitnami"
        chart_name    = "nginx"
        chart_version = "25.0.16"
      }
      options {
        server_side_apply = "auto"
        dry_run_strategy  = "none"
        description       = "NGINX workload managed by Terraform"
      }
    }
  }
}

# Create a Helm4 workload from a Git repository
resource "rafay_workload" "helm4_git_repository" {
  metadata {
    name    = "helm4-git-repository-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    drift {
      action  = "Notify"
      enabled = false
    }
    artifact {
      type = "Helm4"
      artifact {
        repository = "git-repository-name"
        revision   = "main"
        chart_path {
          name = "relative/path/to/chart.tgz"
        }
        values_paths {
          name = "relative/path/to/values.yaml"
        }
      }
      options {
        wait_strategy = "legacy"
        timeout       = "10m0s"
        skip_crds     = true
      }
    }
  }
}

# Create a Helm4 workload from a catalog
resource "rafay_workload" "helm4_catalog" {
  metadata {
    name    = "helm4-catalog-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    artifact {
      type = "Helm4"
      artifact {
        catalog       = "default-bitnami"
        chart_name    = "nginx"
        chart_version = "15.14.0"
        values_ref {
          repository = "git-repository-name"
          revision   = "main"
          values_paths {
            name = "relative/path/to/values.yaml"
          }
        }
      }
      options {
        server_side_apply = "true"
        skip_crds         = true
        sub_notes         = true
      }
    }
  }
}

# Create a Helm4 workload from an uploaded chart without a values file
resource "rafay_workload" "helm4_upload_without_values" {
  metadata {
    name    = "helm4-upload-defaults-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    artifact {
      type = "Helm4"
      artifact {
        chart_path {
          name = "file://relative/path/to/chart.tgz"
        }
      }
      options {
        wait_strategy       = "watcher"
        timeout             = "5m0s"
        rollback_on_failure = true
      }
    }
  }
}

# Create a Helm4 workload from a Helm repository with a values file
resource "rafay_workload" "helm4_helm_repository_with_values" {
  metadata {
    name    = "helm4-helm-repo-values-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    artifact {
      type = "Helm4"
      artifact {
        repository    = "default-bitnami"
        chart_name    = "nginx"
        chart_version = "25.0.16"
        values_paths {
          name = "file://relative/path/to/values.yaml"
        }
      }
      options {
        server_side_apply = "auto"
        wait_strategy     = "watcher"
        timeout           = "5m0s"
      }
    }
  }
}

# Create a Helm4 workload from Git without values, using selector placement and drift protection
resource "rafay_workload" "helm4_git_defaults_labels_drift" {
  metadata {
    name    = "helm4-git-labels-drift-workload"
    project = "project-name"
  }
  spec {
    namespace = "namespace-name"
    version   = "v1"
    placement {
      selector = "rafay.dev/clusterName=cluster-name"
    }
    drift {
      # Increment spec.version when changing the drift action.
      action  = "Deny"
      enabled = true
    }
    artifact {
      type = "Helm4"
      artifact {
        repository = "git-repository-name"
        revision   = "main"
        chart_path {
          name = "relative/path/to/chart.tgz"
        }
      }
      options {
        wait_strategy = "legacy"
        timeout       = "10m0s"
      }
    }
  }
}


