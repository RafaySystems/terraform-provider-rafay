# Trigger a blueprint sync on a cluster and wait for it to complete.
#resource "rafay_blueprint_sync" "sync" {
#  cluster_name = "krishna-jun18"
#  project      = "defaultproject"
#}

#resource "rafay_blueprint_sync" "sync" {
#  cluster_name = "krishna-jun18"
#  project      = "defaultproject"
#  force_sync   = true
#}

# Assign a blueprint name/version to the cluster before syncing. The cluster
# is updated to use this blueprint first, then the sync is published.
#resource "rafay_blueprint_sync" "sync" {
#  cluster_name      = "krishna-jun18"
#  project           = "defaultproject"
#  blueprint_name    = "custom-bp"
#  blueprint_version = "v1"
#}

# Every `terraform apply` re-publishes the blueprint, same as clicking
# "publish" in the UI. force_sync only controls what happens if a sync is
# already in progress: false errors out, true restarts it. Pass
# -var 'force_sync=true' when you need to override an in-progress sync.
resource "rafay_blueprint_sync" "sync" {
  cluster_name      = "krishna-jul17"
  project           = "defaultproject"
  blueprint_name    = "fail-bp"
  blueprint_version = "v2"
  force_sync        = var.force_sync
}
