resource "rafay_secret_group" "tfdemosg" {
  metadata {
    name    = "tfdemosg"
    project = "terraform"
  }
  spec {
    secrets {
      file_path = "aws/credential"
      secret    = "aws-credential"
    }
    secrets {
      file_path = "gke/credential"
      secret    = "gke-credential"
    }
  }
}