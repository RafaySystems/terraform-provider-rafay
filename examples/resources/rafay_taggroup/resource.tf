resource "rafay_tag_group" "tftaggroup1" {
  metadata {
    name      = "tftaggroup1"
    project   = "defaultproject"
  }
  spec {
	tags {
		key   = "tfkey1"
		value = "tfvalue1"
	}
  }
}
