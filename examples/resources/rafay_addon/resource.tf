resource "rafay_addon" "addon" {
  name          = "testing6a"
  projectname   = "dev3"
  namespace     = "testing"
  addontype     = "helm3"
  yamlfilepath  = ""
  chartfile     = "/Users/krishna/Downloads/velero-2.13.6.tgz"
  valuesfile    = "/Users/krishna/Downloads/gh-custom-values.yml"
  versionname   = "v2"
  configmap     = ""
  configuration = ""
  secret        = ""
  statefulset   = ""
}
