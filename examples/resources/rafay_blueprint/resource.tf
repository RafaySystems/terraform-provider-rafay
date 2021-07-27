resource "rafay_blueprint" "blueprint" {
  name        = "rctl-test-blueprint"
  projectname = "dev3"
  yamlfilepath = "<file-path/blueprint.yaml>"
  description  = "blue print with terraform provider"
}
