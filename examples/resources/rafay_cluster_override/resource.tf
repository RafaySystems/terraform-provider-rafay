resource "rafay_groupassociation" "groupassociation" {
  clustername      = "dev1"
  projectname    = "dev1-proj"
  cluster_override_filepath      = "PROJECT_READ_ONLY"
}
#avaliable roles: ["ADMIN", "PROJECT_ADMIN", "PROJECT_READ_ONLY", "INFRA_ADMIN", "INFRA_READ_ONLY", "NAMESPACE_READ_ONLY", "NAMESPACE_ADMIN"]
#avaliable namespaces as for your configuration, only provide when selected roles are namespace options