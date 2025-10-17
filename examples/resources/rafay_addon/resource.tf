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
