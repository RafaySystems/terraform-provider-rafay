
# Username/Password example
resource "rafay_repositories" "tfdemorepository1" {
  metadata {
    name    = "tfdemorepository1"
    project = "upgrade"
  }

  spec {
    type = "Git"
    credentials {
      password = "sealed://credentials.password"
      username = "hardik"
    }
    endpoint = "https://hardik0060@dev.azure.com/hardik0060/gitops/_git/gitops"
    secret {
      name = "file://artifacts/tfdemorepository1/sealed-secret.yaml"
    }
  }
}

# Git private agent example
resource "rafay_repositories" "tfdemorepository2" {
  metadata {
    name    = "tfdemorepository2"
    project = "upgrade"
  }

  spec {
    type     = "Git"
    endpoint = "https://github.com/hardik-rafay/apps.git"
    credentials {
      password = "sealed://credentials.password"
      username = "hardik"
    }

    secret {
      name = "file://artifacts/tfdemorepository1/sealed-secret.yaml"
    }

    options {
      max_retires = 1
      ca_cert {
        name = "file://artifacts/tfdemorepository2/ca-cert.pem"
      }
    }

    agents {
      name = "aganet-v1.10"
    }
  }
}

# Privatekey example
resource "rafay_repositories" "tfdemorepository3" {
  metadata {
    name    = "tfdemorepository3"
    project = "upgrade"
  }
  spec {
    type     = "Helm"
    endpoint = "https://aws.github.io/eks-charts"
    options {
      max_retires = 1
    }
    credentials {
      private_key = "sealed://credentials.privateKey"
    }
    secret {
      name = "file://artifacts/tfdemorepository3/sealed-secret.yaml"
    }
  }
}


# No credentials example
resource "rafay_repositories" "tfdemorepository4" {
  metadata {
    name    = "tfdemorepository4"
    project = "upgrade"
  }

  spec {
    type     = "Helm"
    endpoint = "https://charts.bitnami.com/bitnami"
    options {
      max_retires = 1
    }
  }
}
 