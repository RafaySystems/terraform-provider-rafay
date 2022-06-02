resource "rafay_workloadtemplate" "tfdemoworkloadtemplate1" {
  metadata {
    name    = "tfdemoworkloadtemplate1"
    project = "terraform"
  }
  spec {
    artifact {
      type = "Helm"
      artifact {
        chart_path {
          name = "relative/path/to/some/chart.tgz"
        }
      }
    }
  }
}