provider "google" {
  project = var.project_id
}

data "google_project" "project" {
  project_id = var.project_id
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

    /*
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
        "/*"
      ]
      service = module.lb-http.backend_services["default"].id
    }
    */

    default_service = module.lb-http.backend_services["default"].id

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

resource "google_compute_backend_bucket" "firmware_log_ci" {
  name        = "firmware-log-ci-backend"
  description = "Contains CI firmware transparency log"
  bucket_name = var.bucket_firmware_log_ci
  enable_cdn = false
}

resource "google_compute_backend_bucket" "firmware_artefacts_ci" {
  name        = "firmware-artefacts-ci-backend"
  description = "Contains CI firmware artefacts"
  bucket_name = var.bucket_firmware_artefacts_ci
  enable_cdn = false
}

resource "google_compute_global_network_endpoint_group" "distributor" {
  name                  = "distributor"
  default_port          = var.distributor_port
  network_endpoint_type = "INTERNET_FQDN_PORT"
}

resource "google_compute_global_network_endpoint" "distributor" {
  global_network_endpoint_group = google_compute_global_network_endpoint_group.distributor.name
  port                          = var.distributor_port
  fqdn                          = var.distributor_host
}