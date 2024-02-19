terraform {
  backend "gcs" {}
}

module "shared" {
  source = "../../../../../modules"

  env = var.env
  project_id = var.project_id
  signing_keyring_location = var.signing_keyring_location
  bucket_env = var.bucket_env
  tf_state_location = var.tf_state_location
}

module "triggers" {
  source =  "../../../../../modules/cloudbuild_triggers"

  env = var.env
}
