resource "rafay_addon" "addon" {
  name        = "testing2"
  projectname = "dev3"
  namespace   = "testing"
  addontype   = "NativeYaml"
  yamlfilepath = "<file-path/addon.yml>"
  chartfile   = ""
  valuesfile  = ""
  versionname = "v2"
}
