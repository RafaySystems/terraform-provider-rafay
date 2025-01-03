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

resource "rafay_project_tags_association" "tfptagassociation2" {
  metadata {
    name     = "tfptagassociation2"
    project    = "defaultproject"
  }
  spec {
    associations{
      tag_value = "tag_value"
      resource = "demo-namespace"
      tag_key = "tag_key"
      tag_type = "namespacelabel"
    }
  }
}
