resource "rafay_ztkarule" "demo_rafay_ztkarule_local_file" {
  metadata {
    name = "test-ztkarule-terraform-1"
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifacts/ztka-rule/rule.yaml"
        }
      }
      options {
        force                       = true
        disable_open_api_validation = true
      }
    }
    version = "v1"
  }
}

resource "rafay_ztkarule" "demo_rafay_ztkarule_local_file_1" {
  metadata {
    name = "test-ztkarule-terraform-2"
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "file://artifact/ztka-rule/rule.yaml"
        }
      }
      options {
        force                       = true
        disable_open_api_validation = true
      }
    }
    cluster_selector {
      select_all = true
    }
    project_selector {
      match_names = [
        "defaultproject"
      ]
      select_all = false
    }
    version = "v1"
  }
}

resource "rafay_ztkarule" "demo_rafay_ztkarule_git_file" {
  metadata {
    name = "test-ztkarule-terraform-3"
  }
  spec {
    artifact {
      type = "Yaml"
      artifact {
        paths {
          name = "library/general/allowedrepos/samples/repo-must-be-openpolicyagent/constraint.yaml"
        }
        project    = "defaultproject"
        repository = "opa-lib"
        revision   = "master"
      }
      options {
        force                       = true
        disable_open_api_validation = true
      }
    }
    cluster_selector {
      match_labels = {
        "type" = "eks"
      }
      match_names = [
        "test-cluster"
      ]
      select_all = false
    }
    project_selector {
      match_names = [
        "test-project"
      ]
      select_all = false
    }
    version = "v1"
  }
}