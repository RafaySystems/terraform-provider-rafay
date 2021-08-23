resource "rafay_agent" "agent" {
  projectname    = "defaultproject"
  agent_filepath = "<file=path>/agent.yaml"
}
#agent_filepath is the local filepath to the agent.yaml file we want to add
