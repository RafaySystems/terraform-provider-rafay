data "rafay_fleetplan" "environment_fleetplan" {
  metadata {
    project = "defaultproject"
    name    = "fleetplan-env"
  }
}

data "rafay_fleetplan_jobs" "fleetplan_jobs" {
  fleetplan_name = "fleetplan-env"
  project        = "defaultproject"
}

data "rafay_fleetplan_job" "job1" {
  fleetplan_name = "fleetplan-env"
  project        = "defaultproject"
  name           = "1"
}