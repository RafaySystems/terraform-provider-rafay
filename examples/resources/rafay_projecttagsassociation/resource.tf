resource "rafay_project_tags_association" "tfptagassociation1" {
  metadata {
    name		 = "tfptagassociation1"
    project		 = "defaultproject"
  }
  spec {
	associations {
		tag_key  = "tfkey1"
		tag_type = "Cost"
	}
  }
}
