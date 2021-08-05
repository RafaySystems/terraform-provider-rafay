resource "rafay_cluster_override" "cluster_override" {
  projectname               = "dev1-proj"
  cluster_override_filepath = "/Users/sougat/Downloads/rafay_workload_cluster_overrides/cli/nginx_cluster_override_1.yaml"
}
#cluster_override_filepath is the local filepath to the override.yaml file we want to add 
