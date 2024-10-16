terraform {
  backend "gcs" {}
}

resource "google_service_account" "builder" {
  account_id   = "cloudbuild-${var.env}"
  display_name = "Armored Witness ${var.env} Builder Service Account"
}

resource "google_cloudbuild_trigger" "applet_build" {
  name            = "applet-build-${var.env}"
  location        = "global"
  service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "armored-witness-applet"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }

  build {
    options {
      logging = "CLOUD_LOGGING_ONLY"
    }
    ### Build the Trusted Applet.
    step {
      name       = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=0.0.0 \
          --build-arg=FT_LOG_URL=${var.firmware_base_url}/${var.env}/log/${var.log_shard} \
          --build-arg=FT_BIN_URL=${var.firmware_base_url}/${var.env}/artefacts/${var.log_shard} \
          --build-arg=LOG_ORIGIN=${var.origin_prefix}/${var.log_shard} \
          --build-arg=LOG_PUBLIC_KEY=${var.log_public_key} \
          --build-arg=APPLET_PUBLIC_KEY=${var.applet_public_key} \
          --build-arg=OS_PUBLIC_KEY1=${var.os_public_key1} \
          --build-arg=OS_PUBLIC_KEY2=${var.os_public_key2} \
          --build-arg=REST_DISTRIBUTOR_BASE_URL=${var.rest_distributor_base_url}/${var.env} \
          --build-arg=BASTION_ADDR=${var.bastion_addr} \
          --build-arg=BEE=${var.bee} \
          --build-arg=DEBUG=${var.debug} \
          -t builder-image \
          .
       EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "os_build" {
  name            = "os-build-${var.env}"
  location        = "global"
  service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "armored-witness-os"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }

  build {
    options {
      logging = "CLOUD_LOGGING_ONLY"
    }
    ### Build the Trusted OS.
    step {
      name       = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=0.0.0 \
          --build-arg=LOG_ORIGIN=${var.origin_prefix}/${var.log_shard} \
          --build-arg=LOG_PUBLIC_KEY=${var.log_public_key} \
          --build-arg=APPLET_PUBLIC_KEY=${var.applet_public_key} \
          --build-arg=OS_PUBLIC_KEY1=${var.os_public_key1} \
          --build-arg=OS_PUBLIC_KEY2=${var.os_public_key2} \
          --build-arg=BEE=${var.bee} \
          --build-arg=DEBUG=${var.debug} \
          --build-arg=SRK_HASH=${var.srk_hash} \
          -t builder-image \
          .
       EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "boot_build" {
  name            = "boot-build-${var.env}"
  location        = "global"
  service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "armored-witness-boot"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }

  build {
    options {
      logging = "CLOUD_LOGGING_ONLY"
    }
    ### Build the bootloader.
    step {
      name       = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=0.0.0 \
          --build-arg=LOG_ORIGIN=${var.origin_prefix}/${var.log_shard} \
          --build-arg=LOG_PUBLIC_KEY=${var.log_public_key} \
          --build-arg=OS_PUBLIC_KEY1=${var.os_public_key1} \
          --build-arg=OS_PUBLIC_KEY2=${var.os_public_key1} \
          --build-arg=BEE=${var.bee} \
          --build-arg=CONSOLE=${var.console} \
          -t builder-image \
          .
        EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "recovery_build" {
  name            = "recovery-build-${var.env}"
  location        = "global"
  service_account = google_service_account.builder.id

  github {
    owner = "transparency-dev"
    name  = "armored-witness-boot"

    pull_request {
      branch          = ".*"
      comment_control = "COMMENTS_ENABLED_FOR_EXTERNAL_CONTRIBUTORS_ONLY"
    }
  }

  build {
    options {
      logging = "CLOUD_LOGGING_ONLY"
    }
    ### Build the bootloader.
    step {
      name       = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.recovery_tamago_version} \
          --build-arg=ARMORY_UMS_VERSION=${var.armory_ums_version} \
          -t builder-image \
          recovery
        EOT
      ]
    }
  }
}
