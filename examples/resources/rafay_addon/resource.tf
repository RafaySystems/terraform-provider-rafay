# YAML Upload Example
resource "rafay_addon" "tfdemoaddon1" {
  metadata {
    name    = "tfdemoaddon1"
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
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
    project = "upgrade"
    labels = {
      env  = "dev"
      name = "app"
    }
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
          timeout = "5m0s"
      }
    }
    sharing {
      enabled = true
      projects {
          name = "addons"
      }
      projects {
          name = "ankurp"
      }
    }
  }
}