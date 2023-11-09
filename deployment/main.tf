provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
}

# Enable necessary APIs
resource "google_project_service" "artifactregistry_googleapis_com" {
  service = "artifactregistry.googleapis.com"
}
resource "google_project_service" "cloudasset_googleapis_com" {
  service = "cloudasset.googleapis.com"
}
resource "google_project_service" "cloudbuild_googleapis_com" {
  service = "cloudbuild.googleapis.com"
}
resource "google_project_service" "cloudfunctions_googleapis_com" {
  service = "cloudfunctions.googleapis.com"
}
resource "google_project_service" "cloudkms_googleapis_com" {
  service = "cloudkms.googleapis.com"
}
resource "google_project_service" "compute_googleapis_com" {
  service = "compute.googleapis.com"
}
resource "google_project_service" "containerregistry_googleapis_com" {
  service = "containerregistry.googleapis.com"
}
resource "google_project_service" "logging_googleapis_com" {
  service = "logging.googleapis.com"
}
resource "google_project_service" "pubsub_googleapis_com" {
  service = "pubsub.googleapis.com"
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
  source = "GoogleCloudPlatform/lb-http/google//modules/serverless_negs"
  version = "~> 9.0"

  name                  = var.lb_name
  project               = var.project_id
  load_balancing_scheme = "EXTERNAL_MANAGED"

  ssl                             = var.tls
  managed_ssl_certificate_domains = [var.serve_domain]
  https_redirect                  = var.tls
  
  create_ipv6_address             = true
  enable_ipv6                     = true

  backends = {
    default = {
      description = "Distributor API backend"
      protocol    = "HTTPS"
      port_name   = "https"
      port        = 443
      groups = [
        {
          group = google_compute_global_network_endpoint_group.distributor.id
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
          host_rewrite = var.distributor_host
        }
      }
      service = module.lb-http.backend_services["default"].id
    }

    #####
    # CI log & aretefacts rules
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

    # TODO(prod logs & artefacts)
  }
}

# GCS buckets
resource "google_storage_bucket" "armored_witness_firmware" {
  location                    = "EU"
  name                        = "armored-witness-firmware"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_ci" {
  location                    = "EU"
  name                        = "armored-witness-firmware-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log" {
  location                    = "US"
  name                        = "armored-witness-firmware-log"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket" "armored_witness_firmware_log_ci" {
  location                    = "US"
  name                        = "armored-witness-firmware-log-ci"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}

# Corresponding load balancer backend buckets.
resource "google_compute_backend_bucket" "firmware_log_ci" {
  name        = "firmware-log-ci-backend"
  description = "Contains CI firmware transparency log"
  bucket_name = google_storage_bucket.armored_witness_firmware_log_ci.name
  enable_cdn = false
}

resource "google_compute_backend_bucket" "firmware_artefacts_ci" {
  name        = "firmware-artefacts-ci-backend"
  description = "Contains CI firmware artefacts"
  bucket_name = google_storage_bucket.armored_witness_firmware_ci.name
  enable_cdn = false
}

resource "google_compute_global_network_endpoint_group" "distributor" {
  name                  = "distributor"
  project               = var.project_id
  provider              = google-beta
  default_port          = var.distributor_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}

resource "google_compute_global_network_endpoint" "distributor" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.distributor.name
  port                          = var.distributor_port
  fqdn                          = var.distributor_host
}

# KMS key rings
resource "google_kms_key_ring" "firmware_release_ci" {
  location = "global"
  name     = "firmware-release-ci"
}
resource "google_kms_key_ring" "firmware_release_prod" {
  location = "global"
  name     = "firmware-release-prod"
}

# TODO(jayhou): This configuration cannot be applied right now because of the
# algorithm. Uncomment again when it is supported.
# ### KMS keys
# resource "google_kms_crypto_key" "bootloader_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "bootloader-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "recovery_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "recovery-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_applet_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-applet-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_1_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-os-1-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_2_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "trusted-os-2-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "ft_log_ci" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#   name                       = "ft-log-ci"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "bootloader_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "bootloader-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "recovery_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "recovery-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_applet_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "trusted-applet-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "trusted_os_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "trusted-os-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }
# resource "google_kms_crypto_key" "ft_log_prod" {
#   key_ring                   = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#   name                       = "ft-log-prod"
#   purpose                    = "ASYMMETRIC_SIGN"
#   version_template {
#     algorithm        = "EC_SIGN_ED25519"
#     protection_level = "SOFTWARE"
#   }
# }

# TODO(jayhou): add GCF stuff.