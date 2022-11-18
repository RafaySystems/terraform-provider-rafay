resource "rafay_download_kubeconfig" "tfkubeconfig" {
  namespace          = "vault"
  cluster            = "hardik-pwc-test"
  output_folder_path = ""
  filename           = "kc-testfile"
}