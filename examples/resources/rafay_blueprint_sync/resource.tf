# Trigger a blueprint sync on a cluster and wait for it to complete.
resource "rafay_blueprint_sync" "cluster1" {
  cluster_name = "demo-cluster1"
  project      = "defaultproject"
}

resource "rafay_blueprint_sync" "cluster2" {
  cluster_name = "demo-cluster2"
  project      = "defaultproject"
  force_sync   = true
}

# Assign a blueprint name/version to the cluster before syncing. The cluster
# is updated to use this blueprint first, then the sync is published.
resource "rafay_blueprint_sync" "cluster3" {
  cluster_name      = "demo-cluster3"
  project           = "defaultproject"
  blueprint_name    = "custom-bp"
  blueprint_version = "v1"
}

# Every `terraform apply` re-publishes the blueprint, same as clicking
# "publish" in the UI. force_sync only controls what happens if a sync is
# already in progress: false errors out, true restarts it. Pass
# -var 'force_sync=true' when you need to override an in-progress sync.
resource "rafay_blueprint_sync" "cluster4" {
  cluster_name      = "demo-cluster4"
  project           = "defaultproject"
  blueprint_name    = "custom-bp"
  blueprint_version = "v1"
  force_sync        = var.force_sync
}
