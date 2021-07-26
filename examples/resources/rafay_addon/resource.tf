resource "rafay_addon" "addon" {
  name        = "testing2"
  projectname = "dev3"
  namespace   = "testing"
  addontype   = "alertmanager"
  yamlfilepath = "<file-path/addon.yml>"
  chartfile   = ""
  valuesfile  = ""
  versionname = "v2"
  configmap   = "/Users/krishna/Downloads/alertmanager-configMap.yaml"
  configuration = "/Users/krishna/Downloads/alertmanager-configuration.yaml"
  secret      = "/Users/krishna/Downloads/alertmanager-secret.yaml"
  statefulset = "/Users/krishna/Downloads/alertmanager-statefulSet.yaml"
}
