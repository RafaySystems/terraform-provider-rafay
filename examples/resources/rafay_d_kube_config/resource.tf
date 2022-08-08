resource "rafay_d_kube_config" "tfkubeconfig1" {
  namespace = "vault"
  cluster = "hardik-pwc-test"
  output_folder_path = ""
  filename = "kc-testfile"
}