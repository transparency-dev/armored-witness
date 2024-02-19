terraform {
  backend "gcs" {}
}

module "triggers" {
  source =  "../../../../../modules/cloudbuild_triggers"

  env = var.env
}
