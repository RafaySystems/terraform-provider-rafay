resource "rafay_workload_cd_operator" "operator-demo" {
  metadata {
    name    = "operator-demo"
    project = "benny"
  }
  spec {
    repo_local_path = "./application-repo"
    repo_url        = "https://github.com/demo-user/test-tfcd.git"
    repo_branch     = "main"
    credentials {
      username = "demo-user"
      token = "ghp_XXXXAPIKEYXXXX"
    }

    path_match_pattern = "/:project/:namespace/:workload"
    //cluster_names = "bb6,bb7"
    placement_labels = {
      "owner" = "myteam"
    }
  }
  always_run = "${timestamp()}"
}
