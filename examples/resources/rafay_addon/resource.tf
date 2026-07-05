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
resource "rafay_addon" "tfdemoaddon8" {
  metadata {
    name    = "tfdemoaddon8"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        chart_path {
          name = "file://artifacts/tfdemoaddon4/apache-9.0.9.tgz"
        }
        values_paths {
          name = "file://artifacts/tfdemoaddon4/values.yaml"
        }
      }
      options {
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

# Helm4 Chart from Helm Repository Example
resource "rafay_addon" "tfdemoaddon9" {
  metadata {
    name    = "tfdemoaddon9"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        repository    = "helm-repo"
        chart_name    = "apache"
        chart_version = "9.0.9"
        values_paths {
          name = "file://relative/path/to/some/chart/values.yaml"
        }
      }
      options {
        set               = ["replicaCount=3"]
        server_side_apply = "auto"
        dry_run_strategy  = "none"
        skip_crds         = false
        disable_hooks     = false
        description       = "apache addon managed by terraform"
      }
    }
  }
}

# Helm4 Chart from Git Repository Example
resource "rafay_addon" "tfdemoaddon10" {
  metadata {
    name    = "tfdemoaddon10"
    project = "terraform"
  }
  spec {
    namespace = "tfdemonamespace1"
    version   = "v1.0"
    artifact {
      type = "Helm4"
      artifact {
        repository = "git-repo"
        revision   = "main"
        chart_path {
          name = "charts/apache"
        }
        values_ref {
          repository = "git-repo"
          revision   = "main"
          values_paths {
            name = "charts/apache/values.yaml"
          }
        }
      }
      options {
        labels = {
          "env" = "production"
        }
        dependency_update       = true
        reuse_values            = false
        reset_then_reuse_values = false
        force_conflicts         = false
        take_ownership          = false
        enable_dns              = true
      }
    }
  }
}