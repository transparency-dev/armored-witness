locals {
  project_id               = "armored-witness"
  tf_state_location        = "europe-west2"
  signing_keyring_location = "global"
  env                      = path_relative_to_include()
  rest_distributor_base_url = "https://api.transparency.dev"
  firmware_base_url = "https://api.transparency.dev/armored-witness-firmware"
}

remote_state {
  backend = "gcs"

  config = {
    project  = local.project_id
    location = local.tf_state_location
    bucket   = "${local.project_id}-build-and-release-bucket-tfstate-${local.env}"
    prefix   = "${path_relative_to_include()}/terraform.tfstate"

    gcs_bucket_labels = {
      name  = "terraform_state_storage"
    }
  }
}

