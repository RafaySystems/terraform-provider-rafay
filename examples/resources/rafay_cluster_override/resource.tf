resource "rafay_groupassociation" "groupassociation" {
  clustername      = "dev1"
  projectname    = "dev1-proj"
  cluster_override_filepath      = "PROJECT_READ_ONLY"
}
#cluster_override_filepath is the local filepath to the override.yaml file we want to add 
