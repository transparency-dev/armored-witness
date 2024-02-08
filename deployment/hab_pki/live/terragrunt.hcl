locals {
  project_id               = "armored-witness"
  region                   = "us-central1"
  signing_keyring_location = "global"
  tf_state_location        = "europe-west2"
  env                      = path_relative_to_include()
}

remote_state {
  backend = "gcs"
  config = {
    project  = local.project_id
    location = local.tf_state_location
    bucket   = "${local.project_id}-bucket-tfstate-${local.env}"
    prefix   = "${path_relative_to_include()}/terraform.tfstate"

    gcs_bucket_labels = {
      name  = "terraform_state_storage"
    }
  }
}

