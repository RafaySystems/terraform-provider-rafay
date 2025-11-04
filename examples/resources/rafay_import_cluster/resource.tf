resource "rafay_import_cluster" "import_cluster" {
  clustername     = "terraform-importcluster"
  projectname     = "terraform"
  blueprint       = "default"
  kubeconfig_path = "<file-path/kubeconfig.yaml>"
  location        = "losangeles-us"
  values_path     = "<optional_path/values.yaml>"
  bootstrap_path  = "<optional_path/bootstrap.yaml>"
  labels = {
    "key1" = "value1"
    "key2" = "value2"
  }
  kubernetes_provider   = "AKS"
  provision_environment = "CLOUD"

  proxy_config {
    http_proxy               = "http://10.100.0.10:8080/"
    https_proxy              = "http://10.100.0.10:8080/"
    no_proxy                 = "10.0.0.0/16,localhost,127.0.0.1,internal-service.svc,webhook.svc,10.100.0.0/24,custom-dns.example.com,10.200.0.0/16,10.101.0.0/12,169.254.169.254,.internal.example.com,168.63.129.16,proxy,master.service.consul,10.240.0.0/16,drift-service.svc,*.privatelink.example.com,.privatelink.example.com"
    enabled                  = true
    proxy_auth               = "username:password"
    allow_insecure_bootstrap = false
    bootstrap_ca             = "<ca-certificate-data>"
  }
}

output "values_data" {
  value = rafay_import_cluster.import_cluster.values_data
}

output "values_path" {
  value = rafay_import_cluster.import_cluster.values_path
}

output "bootstrap_data" {
  value = rafay_import_cluster.import_cluster.bootstrap_data
}

output "bootstrap_path" {
  value = rafay_import_cluster.import_cluster.bootstrap_path
}
