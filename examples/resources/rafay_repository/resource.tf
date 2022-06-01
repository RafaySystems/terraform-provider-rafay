
#Username/Password example
resource "rafay_repositories" "tfdemorepository1" {
  metadata {
    name    = "tfdemorepository1"
    project = "bharath"
  }

  spec {
    type = "Git"
    credentials {
      password = "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"
      username = "bharath-rafay"
    }
    endpoint = "https://github.com/bharath-rafay/Golang.git"
  }
}

# Git private agent example
# resource "rafay_repositories" "tfdemorepository2" {
#   metadata {
#     name    = "tfdemorepository2"
#     project = "bharath"
#   }

#   spec {
#     type     = "Git"
#     endpoint = "git@github.com/test/apps.git"
#     credentials {
#       private_key = file("key_file")
#     }
#     agents {
#       name = "gitops-agent"
#     }
#   }
# }

# # Public Helm repo example
# resource "rafay_repositories" "tfdemorepository3" {
#   metadata {
#     name    = "tfdemorepository3"
#     project = "terraform"
#   }

#   spec {
#     type     = "Helm"
#     endpoint = "https://charts.bitnami.com/bitnami"
#   }
# }
