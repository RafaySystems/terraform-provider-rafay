resource "rafay_cluster_override" "cluster_override" {
  projectname               = "dev1-proj"
  cluster_override_filepath = "<file-path/cluster_override.yaml>"
}
#cluster_override_filepath is the local filepath to the override.yaml file we want to add 
