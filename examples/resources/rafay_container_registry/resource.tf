# JFrog example
resource "rafay_container_registry" "tfdemocr" {
  metadata {
    name    = "cr-tf-debug"
    project = "defaultproject"
  }

  spec {
    provider = "JFrog"
    credentials {
      password = "Tesla@3"
      username = "gayibot147@cnxcoin.com"
    }
    endpoint = "testqe.jfrog.io.gov"
  }
}
#Custom Example
# resource "rafay_container_registry" "tfdemocr" {
#   metadata {
#     name    = "cr-tf"
#     project = "v2"
#   }

#   spec {
#     provider = "Custom"
#     credentials {
#       password = "changeplz"
#       username = "sougz"
#     }
#     endpoint = "myregistry.example.com:5000"
#   }
# }


