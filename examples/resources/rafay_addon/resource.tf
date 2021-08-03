resource "rafay_addon" "addon" {
  name          = "testing2"
  projectname   = "dev3"
  namespace     = "testing"
  addontype     = "<yaml/helm/helm3/alertmanager>"
  yamlfilepath  = ""
  chartfile     = ""
  valuesfile    = ""
  versionname   = ""
  configmap     = ""
  configuration = ""
  secret        = ""
  statefulset   = ""
}
