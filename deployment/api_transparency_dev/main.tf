# Configure remote terraform backend for state.
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
    prod = {
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
    ci = {
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
  }

  create_url_map = false
  url_map        = google_compute_url_map.default.self_link
}

resource "google_compute_url_map" "default" {
  name = "api-transparency-dev-url-map"

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
      paths = [
        "/distributor/*"
      ]
      route_action {
        url_rewrite {
          host_rewrite = var.distributor_prod_host
        }
      }
      service = module.lb-http.backend_services["prod"].id
    }
    path_rule {
      paths = [
        "/distributor-ci/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
          host_rewrite        = var.distributor_ci_host
        }
      }
      service = module.lb-http.backend_services["ci"].id
    }

    #####
    ## CI log & aretefacts rules
    # CI log rev 0
    path_rule {
      paths = [
        "/armored-witness-firmware/ci/log/0/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_log_ci.id
    }
    path_rule {
      paths = [
        "/armored-witness-firmware/ci/artefacts/0/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_artefacts_ci.id
    }

    # CI log rev 1
    path_rule {
      paths = [
        "/armored-witness-firmware/ci/log/1/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_log_ci_1.id
    }
    path_rule {
      paths = [
        "/armored-witness-firmware/ci/artefacts/1/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_artefacts_ci_1.id
    }

    # Prod log rev 0 (wave 0 devices)
    path_rule {
      paths = [
        "/armored-witness-firmware/prod/log/0/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_log_prod_0.id
    }
    path_rule {
      paths = [
        "/armored-witness-firmware/prod/artefacts/0/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_artefacts_prod_0.id
    }

    # Prod log rev 1
    path_rule {
      paths = [
        "/armored-witness-firmware/prod/log/1/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_log_prod_1.id
    }
    path_rule {
      paths = [
        "/armored-witness-firmware/prod/artefacts/1/*"
      ]
      route_action {
        url_rewrite {
          path_prefix_rewrite = "/"
        }
      }
      service = google_compute_backend_bucket.firmware_artefacts_prod_1.id
    }
  }
}

## Corresponding load balancer backend buckets.
# CI log rev 0
resource "google_compute_backend_bucket" "firmware_log_ci" {
  name        = "firmware-log-ci-backend"
  description = "Contains CI firmware transparency log"
  bucket_name = "armored-witness-firmware-log-ci" # google_storage_bucket.armored_witness_firmware_log_ci.name
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_ci" {
  name        = "firmware-artefacts-ci-backend"
  description = "Contains CI firmware artefacts for FT log"
  bucket_name = "armored-witness-firmware-ci" # google_storage_bucket.armored_witness_firmware_ci.name
  enable_cdn  = false
}

# CI log rev 1
resource "google_compute_backend_bucket" "firmware_log_ci_1" {
  name        = "firmware-log-ci-backend-1"
  description = "Contains CI firmware transparency log 1"
  bucket_name = "armored-witness-firmware-log-ci-1" # google_storage_bucket.armored_witness_firmware_log_ci_1.name
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_ci_1" {
  name        = "firmware-artefacts-ci-backend-1"
  description = "Contains CI firmware artefacts for FT log 1"
  bucket_name = "armored-witness-firmware-ci-1" # google_storage_bucket.armored_witness_firmware_ci_1.name
  enable_cdn  = false
}

# Prod log 0 (Q1 2024 - wave 0 devices)
resource "google_compute_backend_bucket" "firmware_log_prod_0" {
  name        = "firmware-log-prod-backend-0"
  description = "Contains prod firmware transparency log 0"
  bucket_name = "armored-witness-firmware-log"
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_prod_0" {
  name        = "firmware-artefacts-prod-backend-0"
  description = "Contains prod firmware artefacts for FT log 0"
  bucket_name = "armored-witness-firmware"
  enable_cdn  = false
}
# Prod log 1
resource "google_compute_backend_bucket" "firmware_log_prod_1" {
  name        = "firmware-log-prod-backend-1"
  description = "Contains prod firmware transparency log 1"
  bucket_name = "armored-witness-firmware-log-1"
  enable_cdn  = false
}
resource "google_compute_backend_bucket" "firmware_artefacts_prod_1" {
  name        = "firmware-artefacts-prod-backend-1"
  description = "Contains prod firmware artefacts for FT log 1"
  bucket_name = "armored-witness-firmware-1"
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
