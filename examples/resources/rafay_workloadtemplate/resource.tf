resource "rafay_workload_template" "tfdemoworkloadtemplate1" {
  metadata {
    name    = "tfdemoworkloadtemplate1"
    project = "upgrade"
  }
  spec {
    type = "Helm"
    artifact {
      chart_path {
        name = "relative/path/to/some/chart.tgz"
      }
    }
  }
}