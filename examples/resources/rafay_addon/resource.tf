resource "rafay_addon" "addon" {
  name          = "testing2"
  projectname   = "dev3"
  namespace     = "testing"
  addontype     = "alertmanager"
  yamlfilepath  = "<file-path/addon.yml>"
  chartfile     = ""
  valuesfile    = ""
  versionname   = "v2"
  configmap     = "<file-path/alertmanager-configMap.yaml>"
  configuration = "<file-path/alertmanager-configuration.yaml>"
  secret        = "<file-path/alertmanager-secret.yaml>"
  statefulset   = "<file-path/alertmanager-statefulSet.yaml>"
}
