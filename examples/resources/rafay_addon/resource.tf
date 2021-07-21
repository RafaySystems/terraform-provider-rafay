resource "rafay_addon" "addon" {
  name        = "testing2"
  projectname = "dev3"
  namespace   = "testing"
  addontype   = "NativeYaml"
  yamlfilepath = "/Users/krishna/Downloads/metallb-cm-final.yml"
  chartfile   = ""
  valuesfile  = ""
  versionname = "v2"
}
