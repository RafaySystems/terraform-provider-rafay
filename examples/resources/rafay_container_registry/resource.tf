# JFrog example
resource "rafay_container_registry" "tfdemocr" {
  metadata {
    name    = "cr-tf"
    project = "defaultproject"
  }

  spec {
    provider = "JFrog"
    credentials {
      password = "Tesla@321"
      username = "gayibot147@cnxcoin.com"
    }
    endpoint = "testqe.jfrog.io"
  }
}
# #Custom Example
# resource "rafay_container_registry" "tfdemocr" {
#   metadata {
#     name    = "cr-tf"
#     project = "defaultproject"
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


