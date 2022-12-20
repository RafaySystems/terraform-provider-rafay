resource "rafay_download_kubeconfig" "tfkubeconfig" {
  cluster            = "terraform"
  output_folder_path = "/tmp"
  filename           = "kubeconfig"
}
