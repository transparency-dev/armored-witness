# configure remote terraform backend for state.
terraform {
  backend "gcs" {
    bucket = "armored-witness-bucket-tfstate"
    prefix = "terraform/api.transparency.dev/state"
  }
}

# Project
provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
}

data "terraform_remote_state" "ci_build_artefacts" {
  backend   = "gcs"
  workspace = terraform.workspace
  config = {
    bucket = "${var.project_name}-build-and-release-bucket-tfstate-ci"
    prefix = "ci/terraform.tfstate"
  }
}
data "terraform_remote_state" "prod_build_artefacts" {
  backend   = "gcs"
  workspace = terraform.workspace
  config = {
    bucket = "${var.project_name}-build-and-release-bucket-tfstate-prod"
    prefix = "prod/terraform.tfstate"
  }
}

# Enable necessary APIs
resource "google_project_service" "cloudkms_googleapis_com" {
  service = "cloudkms.googleapis.com"
}
resource "google_project_service" "compute_googleapis_com" {
  service = "compute.googleapis.com"
}
resource "google_project_service" "serviceusage_googleapis_com" {
  service = "serviceusage.googleapis.com"
}
resource "google_project_service" "storage_api_googleapis_com" {
  service = "storage-api.googleapis.com"
}
resource "google_project_service" "storage_component_googleapis_com" {
  service = "storage-component.googleapis.com"
}
resource "google_project_service" "storage_googleapis_com" {
  service = "storage.googleapis.com"
}

module "lb-http" {
  source  = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 10.0"

  name                  = var.lb_name
  project               = var.project_id
  load_balancing_scheme = "EXTERNAL_MANAGED"

  ssl                             = var.tls
  managed_ssl_certificate_domains = [var.serve_domain]
  https_redirect                  = var.tls

  create_ipv6_address = true
  enable_ipv6         = true

  backends = {
    distributor-prod = {
      description = "Distributor API backend (prod)"
      protocol    = "HTTPS"
      port_name   = "https"
      port        = 443
      groups = [
        {
          group = google_compute_global_network_endpoint_group.distributor_prod.id
        }
      ]

      health_check = null

      enable_cdn = false

      iap_config = {
        enable = false
      }
      log_config = {
        enable = false
      }
    }
    distributor-ci = {
      description = "Distributor API backend (ci)"
      protocol    = "HTTPS"
      port_name   = "https"
      port        = 443
      groups = [
        {
          group = google_compute_global_network_endpoint_group.distributor_ci.id
        }
      ]

      health_check = null

      enable_cdn = false

      iap_config = {
        enable = false
      }
      log_config = {
        enable = false
      }
    }
    distributor-dev = {
      description = "Distributor API backend (dev)"
      protocol    = "HTTPS"
      port_name   = "https"
      port        = 443
      groups = [
        {
          group = google_compute_global_network_endpoint_group.distributor_dev.id
        }
      ]

      health_check = null

      enable_cdn = false

      iap_config = {
        enable = false
      }
      log_config = {
        enable = false
      }
    }

    witness-dev = {
      description = "Witness API backend (dev)"
      protocol    = "HTTPS"
      port_name   = "https"
      port        = 443
      groups = [
        {
          group = google_compute_global_network_endpoint_group.witness_dev.id
        }
      ]

      health_check = null

      enable_cdn = false

      iap_config = {
        enable = false
      }
      log_config = {
        enable = false
      }
    }
  }

  create_url_map = false
  url_map        = google_compute_url_map.default.self_link
}

resource "random_id" "url_map" {
  keepers = {
    # Generate a new id each time the instance list changes
    instances = base64encode(jsonencode(module.lb-http.backend_services))
  }

  byte_length = 4
}

resource "google_compute_url_map" "default" {
  name = "api-transparency-dev-url-map-${random_id.url_map.hex}"

  default_url_redirect {
    https_redirect = true
    host_redirect  = "transparency.dev"
    path_redirect  = "/"
    strip_query    = true
  }

  host_rule {
    hosts        = [var.serve_domain]
    path_matcher = "allpaths"
  }

  path_matcher {
    name = "allpaths"

    # If we don't know what this is, send them to the website.
    default_url_redirect {
      https_redirect = true
      host_redirect  = "transparency.dev"
      path_redirect  = "/"
      strip_query    = true
    }

    #####
    # Distributor rules

    path_rule {
      # We're currently aliasing the prod distributor on two prefixes:
      #  - the "/distributor/..." path for checkpoint consumers (this is "vanity" so that the API url doesn't highlight the env)
      #  - on `/prod/distributor/...` for witness devices (this is to avoid special-casing prod builds)
      paths = [
        "/distributor/*",
        "/prod/distributor/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/distributor/"
          host_rewrite        = var.distributor_prod_host
        }
      }
      service = module.lb-http.backend_services["distributor-prod"].id
    }
    path_rule {
      paths = [
        # match on /distributor/ to prevent /metrics being exposed publicly
        "/ci/distributor/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/distributor/"
          host_rewrite        = var.distributor_ci_host
        }
      }
      service = module.lb-http.backend_services["distributor-ci"].id
    }
    path_rule {
      paths = [
        # match on /distributor/ to prevent /metrics being exposed publicly
        "/dev/distributor/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/distributor/"
          host_rewrite        = var.distributor_dev_host
        }
      }
      service = module.lb-http.backend_services["distributor-dev"].id
    }

    #####
    # Witness rules

    path_rule {
      paths = [
        "/dev/witness/little-garden/add-checkpoint",
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/add-checkpoint"
          host_rewrite        = var.witness_dev_host
        }
      }
      service = module.lb-http.backend_services["witness-dev"].id
    }

    #####
    ## CI log & aretefacts rules
    dynamic "path_rule" {
      for_each = data.terraform_remote_state.ci_build_artefacts.outputs.firmware_log_buckets
      iterator = i

      content {
        paths = [
          "/armored-witness-firmware/ci/log/${i.key}/*"
        ]
        route_action {
          url_rewrite {
            path_prefix_rewrite = "/"
          }
        }
        service = google_compute_backend_bucket.firmware_log_ci["${i.key}"].id
      }
    }
    dynamic "path_rule" {
      for_each = data.terraform_remote_state.ci_build_artefacts.outputs.firmware_artefact_buckets
      iterator = i

      content {
        paths = [
          "/armored-witness-firmware/ci/artefacts/${i.key}/*"
        ]
        route_action {
          url_rewrite {
            path_prefix_rewrite = "/"
          }
        }
        service = google_compute_backend_bucket.firmware_artefacts_ci["${i.key}"].id
      }
    }

    ## Prod (wave 0 devices) log & aretefacts rules
    dynamic "path_rule" {
      for_each = data.terraform_remote_state.prod_build_artefacts.outputs.firmware_log_buckets
      iterator = i

      content {
        paths = [
          "/armored-witness-firmware/prod/log/${i.key}/*"
        ]
        route_action {
          url_rewrite {
            path_prefix_rewrite = "/"
          }
        }
        service = google_compute_backend_bucket.firmware_log_prod["${i.key}"].id
      }
    }
    dynamic "path_rule" {
      for_each = data.terraform_remote_state.prod_build_artefacts.outputs.firmware_artefact_buckets
      iterator = i

      content {
        paths = [
          "/armored-witness-firmware/prod/artefacts/${i.key}/*"
        ]
        route_action {
          url_rewrite {
            path_prefix_rewrite = "/"
          }
        }
        service = google_compute_backend_bucket.firmware_artefacts_prod["${i.key}"].id
      }
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}

## Corresponding load balancer backend buckets.
# CI logs
resource "google_compute_backend_bucket" "firmware_log_ci" {
  for_each = data.terraform_remote_state.ci_build_artefacts.outputs.firmware_log_buckets

  name        = "firmware-log-ci-backend-${each.key}"
  description = "Contains CI firmware transparency log ${each.key}"
  bucket_name = each.value
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_ci" {
  for_each = data.terraform_remote_state.ci_build_artefacts.outputs.firmware_artefact_buckets

  name        = "firmware-artefacts-ci-backend-${each.key}"
  description = "Contains CI firmware artefacts for FT log ${each.key}"
  bucket_name = each.value
  enable_cdn  = false
}

# Prod logs (Q1 2024 - wave 0 devices)
resource "google_compute_backend_bucket" "firmware_log_prod" {
  for_each = data.terraform_remote_state.prod_build_artefacts.outputs.firmware_log_buckets

  name        = "firmware-log-prod-backend-${each.key}"
  description = "Contains prod firmware transparency log ${each.key}"
  bucket_name = each.value
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_prod" {
  for_each = data.terraform_remote_state.prod_build_artefacts.outputs.firmware_artefact_buckets

  name        = "firmware-artefacts-prod-backend-${each.key}"
  description = "Contains prod firmware artefacts for FT log ${each.key}"
  bucket_name = each.value
  enable_cdn  = false
}

resource "google_compute_global_network_endpoint_group" "distributor_prod" {
  name                  = "distributor-prod"
  project               = var.project_id
  provider              = google-beta
  default_port          = var.distributor_prod_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}
resource "google_compute_global_network_endpoint_group" "distributor_ci" {
  name                  = "distributor-ci"
  project               = var.project_id
  provider              = google-beta
  default_port          = var.distributor_ci_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}
resource "google_compute_global_network_endpoint_group" "distributor_dev" {
  name                  = "distributor-dev"
  project               = var.project_id
  provider              = google-beta
  default_port          = var.distributor_dev_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}
resource "google_compute_global_network_endpoint_group" "witness_dev" {
  name                  = "witness-dev"
  project               = var.project_id
  provider              = google-beta
  default_port          = var.witness_dev_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}


resource "google_compute_global_network_endpoint" "distributor_prod" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.distributor_prod.name
  port                          = var.distributor_prod_port
  fqdn                          = var.distributor_prod_host
}
resource "google_compute_global_network_endpoint" "distributor_ci" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.distributor_ci.name
  port                          = var.distributor_ci_port
  fqdn                          = var.distributor_ci_host
}
resource "google_compute_global_network_endpoint" "distributor_dev" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.distributor_dev.name
  port                          = var.distributor_dev_port
  fqdn                          = var.distributor_dev_host
}
resource "google_compute_global_network_endpoint" "witness_dev" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.witness_dev.name
  port                          = var.witness_dev_port
  fqdn                          = var.witness_dev_host
}

## Terraform keys
## Commented out here as they're provided in the build_and_release unit.
#resource "google_kms_key_ring" "terraform_state" {
#  name     = "armored-witness-bucket-tfstate"
#  location = var.tf_state_location
#}
#
#resource "google_kms_crypto_key" "terraform_state_bucket" {
#  name     = "terraform-state-bucket"
#  key_ring = google_kms_key_ring.terraform_state.id
#}
#
#resource "google_storage_bucket" "terraform_state" {
#  name          = "armored-witness-bucket-tfstate"
#  force_destroy = false
#  location      = var.tf_state_location
#  storage_class = "STANDARD"
#  versioning {
#    enabled = true
#  }
#  encryption {
#    default_kms_key_name = google_kms_crypto_key.terraform_state_bucket.id
#  }
#  uniform_bucket_level_access = true
#}

