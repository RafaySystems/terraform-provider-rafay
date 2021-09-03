resource "rafay_pipeline" "pipeline" {
  projectname       = "dev1-proj"
  pipeline_filepath = "<filepath>/pipeline.yaml"
}
#pipeline_filepath is the local filepath to the pipeline.yaml file we want to add
