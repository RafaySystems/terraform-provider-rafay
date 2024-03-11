resource "rafay_workload_cd_operator" "operatordemo" {
  metadata {
    name    = "operatordemo"
    project = "demoorg"
  }
  spec {
    repo_local_path = "/tmp/application-repo"
    repo_url        = "https://github.com/stephan-rafay/test-tfcd.git"
    repo_branch     = "main"
    #credentials {
    #  username = var.git_user
    #  token = var.git_token
    #}

    
    workload {
      name = "echoserver"
      chart_helm_repo_name = "echo-server"
      helm_chart_version = "0.5.0"
      helm_chart_name = "echo-server"
      path_match_pattern = "/:project/:workload/:namespace"
      base_path = "echoserver-common"
      include_base_value = true
      delete_action = "delete"
      placement_labels = {
        "echoserver" = "enabled"
      }
    }

    workload {
      name = "hello"
      chart_git_repo_path = "/hello-common/hello-0.1.3.tgz"
      chart_git_repo_name = "hello-repo"
      helm_chart_version = "0.1.3"
      helm_chart_name = "hello"
      path_match_pattern = "/:project/:workload/:namespace"
      base_path = "hello-common"
      include_base_value = true
      delete_action = "delete"
      placement_labels = {
        "hello" = "enabled"
      }
    }

    workload {
      name = "httpecho-us"
      helm_chart_version = "0.3.4"
      helm_chart_name = "http-echo"
      path_match_pattern = "/:project/:workload/:namespace"
      base_path = "httpecho-common"
      include_base_value = true
      delete_action = "delete"
      placement_labels = {
        "httpecho-us" = "enabled"
      }
    }

    workload {
      name = "httpecho-eu"
      helm_chart_version = "0.3.4"
      helm_chart_name = "http-echo"
      path_match_pattern = "/:project/:workload/:namespace"
      base_path = "httpecho-common"
      include_base_value = true
      delete_action = "delete"
      placement_labels = {
        "httpecho-eu" = "enabled"
      }
    }
   
  }
  always_run = "${timestamp()}"
}
