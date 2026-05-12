# example 1: Download config files automatically
data "rafay_agent_docker_config" "agent_docker_config" {
  depends_on = [rafay_agent.a1]
  project    = "defaultproject"
  agent_name = "test-agent"

  download_config_files = true
  download_directory    = "/some/path/here" # default is current dir
}

resource "null_resource" "start_agent" {
  depends_on = [data.rafay_agent_docker_config.agent_docker_config]

  provisioner "local-exec" {
    command = data.rafay_agent_docker_config.agent_docker_config.start_command
  }
}
