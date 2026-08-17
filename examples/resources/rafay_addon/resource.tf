# YAML Upload Example
resource "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/tfdemoaddon1/busybox.yaml"
        }

      }
    }
    sharing {
      enabled = false
    }
  }
}


# Helm Chart Upload Example
resource "rafay_addon" "tfdemoaddon4" {
  metadata {
    name    = "tfdemoaddon4"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "file://artifacts/tfdemoaddon4/apache-9.0.9.tgz"
        }
      }
      options {
        max_history = 10
        timeout     = "5m0s"
      }
    }
    sharing {
      enabled = true
      projects {
        name = "project1"
      }
      projects {
        name = "project2"
      }
    }
  }
}

# Catalog Example
resource "rafay_addon" "tfdemoaddon2" {
  metadata {
    name    = "tfdemoaddon2"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm"
      artifact {
        catalog       = "catalogName"
        chart_name    = "chartName"
        chart_version = "chartVersion"
        values_paths {
          name = "file://relative/path/to/some/chart/values.yaml"
        }
      }
      options {
        max_history = 10
        timeout     = "5m0s"
      }
    }
  }
}


# Web YAML
resource "rafay_addon" "tfdemoaddon5" {
  metadata {
    name    = "tfdemoaddon5"
    project = "terraform"
  }
  spec {
    namespace     = "tfdemonamespace1"
    version       = "v1.0"
    version_state = "active"
    artifact {
      type = "Yaml"
      artifact {
        url = ["https://raw.githubusercontent.com/kubernetes/website/main/content/en/examples/application/nginx-app.yaml"]
      }
    }
    sharing {
      enabled = false
    }
  }
}

resource "rafay_addon" "tfdemoaddon6" {
  metadata {
    name    = "tfdemoaddon6"
    project = "terraform"
  }
  spec {
    namespace = "default"
    version   = "production"
    artifact {
      type = "Kustomize"
      artifact {
        path = "production"
        file {
          name = "file://artifacts/tfdemoaddon6/archive.tar.gz"
        }
      }
    }
    sharing {
      enabled = false
    }
  }
}

resource "rafay_addon" "tfdemoaddon7" {
  metadata {
    name    = "tfdemoaddon7"
    project = "terraform"
  }
  spec {
    namespace = "default"
    version   = "prod"
    artifact {
      type = "Kustomize"
      artifact {
        repository = "kustomize-repo"
        revision   = "master"
        directory  = "examples/multibases"
        path       = "production"
      }
    }
    sharing {
      enabled = false
    }
  }
}

# Helm4 Chart Upload Example
resource "rafay_addon" "helm4_upload" {
  metadata {
    name    = "helm4-upload-addon"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        chart_path {
          name = "file://artifacts/tfdemoaddon6/set-test-chart-0.1.0.tgz"
        }
        values_paths {
          name = "file://artifacts/tfdemoaddon6/values.yaml"
        }
      }
      options {
        set                 = ["replicaCount=3"]
        set_string          = ["image.tag=v1.0.0"]
        wait_strategy       = "watcher"
        wait_for_jobs       = true
        timeout             = "5m0s"
        max_history         = 10
        cleanup_on_fail     = true
        rollback_on_failure = true
      }
    }
    sharing {
      enabled = false
    }
  }
}

# Helm4 Chart from Helm Repository Example
resource "rafay_addon" "helm4_helm_repository" {
  metadata {
    name    = "helm4-helm-repository-addon"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        repository    = "helm4-repo"
        chart_name    = "redis"  # chart name
        chart_version = "27.0.12"   # chart version
      }
      options {
        server_side_apply = "auto"
        dry_run_strategy  = "none"
        description       = "Redis add-on managed using Terraform"
      }
    }
    sharing {
      enabled = false
    }
  }
}

# Helm4 Chart from Git Repository Example
resource "rafay_addon" "helm4_git_repository" {
  metadata {
    name    = "helm4-git-repository-addon"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        repository = "git-helm-charts-repo"
        revision   = "main"
        chart_path {
          name = "path/to/chart/file/in/git/test-chart-2.4.1.tgz"
        }
        values_ref {
          repository = "git-helm-values-repo"
          revision   = "main"
          values_paths {
            name = "path/to/values/values.yaml"
          }
        }
      }
      options {
        labels = {
          environment = "production"
        }
        wait_strategy   = "legacy"
        reuse_values    = true
        force_conflicts = false
        take_ownership  = false
        enable_dns      = true
      }
    }
    sharing {
      enabled = false
    }
  }
}

# Helm4 Chart from Catalog Example
resource "rafay_addon" "helm4_catalog" {
  metadata {
    name    = "helm4-catalog-addon"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        catalog       = "default-bitnami"
        chart_name    = "nginx"
        chart_version = "15.14.0"
        values_paths {
          name = "file://relative/path/to/values.yaml"
        }
      }
      options {
        server_side_apply           = "true"
        skip_crds                   = false
        skip_schema_validation      = false
        disable_open_api_validation = false
        disable_hooks               = false
        sub_notes                   = true
      }
    }
    sharing {
      enabled = false
    }
  }
}
