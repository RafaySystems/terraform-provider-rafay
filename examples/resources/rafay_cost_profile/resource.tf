#Example cost profile for AWS
resource "rafay_cost_profile" "tfdemocostprofile1" {
  metadata {
    name    = "tfdemocostprofile1"
    project = "terraform"
  }
  spec {
    version = "v0"
    provider_type = "aws"
    installation_params {
      aws {
        spot_integration {
          aws_spot_data_region = "sample"
          aws_spot_data_bucket = "sample"
          aws_spot_data_prefix = "sample"
          aws_account_id = "sample"
        }
        cur_integration {
          athena_region = "sample"
          athena_bucket_name = "sample"
          athena_database = "sample"
          athena_table = "sample"
          master_payer_arn = "sample"
        }
        aws_credentials {
          cloud_credentials_name = "sample"
        }
      }
    }
    sharing {
      enabled = true
      projects {
        name = "terraformproject2"
      }
    }
  }
}
#Example cost profile for Azure
resource "rafay_cost_profile" "tfdemocostprofile-azure" {
  metadata {
    name    = "tfdemocostprofile-azure"
    project = "terraform"
  }
  spec {
    version = "v0"
    provider_type = "azure"
    installation_params {
      azure {
        custom_pricing {
          cloud_credentials_name = "sample"
        }
      }
    }
  }
}
#Example cost profile for Gcp
resource "rafay_cost_profile" "tfdemocostprofile-gcp" {
  metadata {
    name    = "tfdemocostprofile-gcp"
    project = "terraform"
  }
  spec {
    version = "v0"
    provider_type = "gcp"
    installation_params {
      gcp {
        gcp_credentials {
          cloud_credentials_name = "sample"
        }
      }
    }
  }
}
#Example cost profile for Other providers
resource "rafay_cost_profile" "tfdemocostprofile-other" {
  metadata {
    name    = "tfdemocostprofile-other"
    project = "terraform"
  }
  spec {
    version = "v0"
    provider_type = "other"
    installation_params {
      other {
        cpu = "2.5"
        gpu = "3.61"
        memory = "4.3"
      }
    }
  }
}
