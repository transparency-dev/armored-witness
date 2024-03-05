# Configure remote terraform backend for state.
# This will be configured by terragrunt when deploying.
terraform {
  backend "gcs" {}
}

# Project
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

# GCS buckets
resource "google_storage_bucket" "firmware" {
  count = var.bucket_count

  location                    = "EU"
  name                        = "armored-witness-firmware-${var.env}-${count.index}"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket_iam_member" "firmware_object_reader" {
  count = var.bucket_count

  bucket = google_storage_bucket.firmware["${count.index}"].name
  role    = "roles/storage.legacyObjectReader"
  member  = "allUsers"
}
resource "google_storage_bucket_iam_member" "firmware_bucket_reader" {
  count = var.bucket_count

  bucket = google_storage_bucket.firmware["${count.index}"].name
  role    = "roles/storage.legacyBucketReader"
  member  = "allUsers"
}

resource "google_storage_bucket" "firmware_log" {
  count = var.bucket_count

  location                    = "US"
  name                        = "armored-witness-firmware-log-${var.env}-${count.index}"
  storage_class               = "STANDARD"
  uniform_bucket_level_access = true
}
resource "google_storage_bucket_iam_member" "firmware_log_object_reader" {
  count = var.bucket_count

  bucket = google_storage_bucket.firmware_log["${count.index}"].name
  role    = "roles/storage.legacyObjectReader"
  member  = "allUsers"
}
resource "google_storage_bucket_iam_member" "firmware_log_bucket_reader" {
  count = var.bucket_count

  bucket = google_storage_bucket.firmware_log["${count.index}"].name
  role    = "roles/storage.legacyBucketReader"
  member  = "allUsers"
}

# KMS key rings & data sources
resource "google_kms_key_ring" "firmware_release" {
  location = var.signing_keyring_location
  name     = "firmware-release-${var.env}"
}

# TODO(jayhou): This configuration cannot be applied right now because of the
# algorithm. Uncomment again when it is supported.
### KMS keys
#resource "google_kms_crypto_key" "bootloader_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "bootloader-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "recovery_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "recovery-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_applet_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-applet-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_1_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-os-1-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_2_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "trusted-os-2-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "ft_log_ci" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-ci"
#  name     = "ft-log-ci"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "bootloader_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "bootloader-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "recovery_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "recovery-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_applet_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "trusted-applet-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "trusted_os_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "trusted-os-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}
#resource "google_kms_crypto_key" "ft_log_prod" {
#  key_ring = "projects/armored-witness/locations/global/keyRings/firmware-release-prod"
#  name     = "ft-log-prod"
#  purpose  = "ASYMMETRIC_SIGN"
#  version_template {
#    algorithm        = "EC_SIGN_ED25519"
#    protection_level = "SOFTWARE"
#  }
#}

resource "google_service_account" "builder" {
  account_id   = "cloudbuild-${var.env}"
  display_name = "Armored Witness ${var.env} Builder Service Account"
}

data "terraform_remote_state" "hab_pki" {
  backend = "gcs"
  workspace = terraform.workspace
  config  = {
    bucket = "${var.project_id}-bucket-tfstate-${var.env}"
    prefix = "${var.env}/terraform.tfstate"
  }
}

resource "google_cloudbuild_trigger" "applet_build" {
  name = "applet-build-${var.env}"
  location = "global"

  github {
    owner = "transparency-dev"
    name  = "armored-witness-applet"

    push {
      branch = var.cloudbuild_trigger_branch != "" ? var.cloudbuild_trigger_branch : null
      tag = var.cloudbuild_trigger_tag != "" ? var.cloudbuild_trigger_tag : null
    }
  }

  build {
    # If the trigger is not based on `tag`, create a fake one.
    #
    # Unfortunately, GCB has no concept of dynamically creating substitutions or
    # passing ENV vars between steps, so the best we can do is to create a file
    # containing our tag in the shared workspace which other steps can inspect.
    step {
      name = "bash"
      script = (
        var.cloudbuild_trigger_tag != "" ?
        "$TAG_NAME > /workspace/git_tag && cat /workspace/git_tag" :
        "date +'0.3.%s-incompatible' > /workspace/git_tag && cat /workspace/git_tag"
      )
    }
    ### Build the Trusted OS and upload it to GCS.
    # Build an image containing the Trusted OS artifacts with the Dockerfile.
    # This step needs to be a bash script in order to read the tag content from file.
    step {
      name = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=$(cat /workspace/git_tag) \
          --build-arg=FT_LOG_URL=${var.firmware_base_url}/${var.env}/log/${var.log_shard} \
          --build-arg=FT_BIN_URL=${var.firmware_base_url}/${var.env}/artefacts/${var.log_shard} \
          --build-arg=LOG_ORIGIN=${var.origin_prefix}/${var.log_shard} \
          --build-arg=LOG_PUBLIC_KEY=${var.log_public_key} \
          --build-arg=APPLET_PUBLIC_KEY=${var.applet_public_key} \
          --build-arg=OS_PUBLIC_KEY1=${var.os_public_key1} \
          --build-arg=OS_PUBLIC_KEY2=${var.os_public_key2} \
          --build-arg=REST_DISTRIBUTOR_BASE_URL=${var.rest_distributor_base_url}/${var.env} \
          --build-arg=BEE=${var.bee} \
          --build-arg=DEBUG=${var.debug} \
          -t builder-image \
          .
       EOT
      ]
    }
    # Prepare a container with a copy of the artifacts.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "create",
        "--name",
        "builder_scratch",
        "builder-image",
      ]
    }
    # Copy the artifacts from the container to the Cloud Build VM.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
       "cp",
       "builder_scratch:/build/bin",
       "output",
      ]
    }
    # List the artifacts.
    step {
      name = "bash"
      script = "ls output"
    }
    # Copy the artifacts from the Cloud Build VM to GCS.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/trusted_applet.elf \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/trusted_applet.elf | cut -f1 -d" ")
        EOT
      ]
    }
    ### Construct log entry / Claimant Model statement.
    # This step needs to be a bash script in order to read the tag content
    # from file.
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/transparency-dev/armored-witness/cmd/manifest@main \
          create \
          --git_tag=$(cat /workspace/git_tag) \
          --git_commit_fingerprint=$COMMIT_SHA \
          --firmware_file=output/trusted_applet.elf \
          --firmware_type=TRUSTED_APPLET \
          --tamago_version=${var.tamago_version} \
          --build_env="FT_LOG_URL=${var.firmware_base_url}/${var.env}/log/${var.log_shard}" \
          --build_env="FT_BIN_URL=${var.firmware_base_url}/${var.env}/artefacts/${var.log_shard}" \
          --build_env="LOG_ORIGIN=${var.origin_prefix}/${var.log_shard}" \
          --build_env="LOG_PUBLIC_KEY=${var.log_public_key}" \
          --build_env="APPLET_PUBLIC_KEY=${var.applet_public_key}" \
          --build_env="OS_PUBLIC_KEY1=${var.os_public_key1}" \
          --build_env="OS_PUBLIC_KEY2=${var.os_public_key2}" \
          --build_env="REST_DISTRIBUTOR_BASE_URL=${var.rest_distributor_base_url}/${var.env}" \
          --build_env="BEE=${var.bee}" \
          --build_env="DEBUG=${var.debug}" \
          --build_env="SRK_HASH=${var.srk_hash}" \
          --raw \
          --output_file=output/trusted_applet_manifest_unsigned.json
        EOT
      ]
    }
    # Sign the log entry.
    step {
      name = "golang"
      args = [
        "go",
        "run",
        "github.com/transparency-dev/armored-witness/cmd/sign@main",
        "--project_name=$PROJECT_ID",
        "--release=${var.env}",
        "--artefact=applet",
        "--manifest_file=output/trusted_applet_manifest_unsigned.json",
        "--output_file=output/trusted_applet_manifest",
      ]
    }
    # Print the content of the signed manifest.
    step {
      name = "bash"
      script = "cat output/trusted_applet_manifest"
    }
    ### Write the firmware release to the CI transparency log.
    # Copy the signed note to the sequence bucket, preparing to write to log.
    #
    # Use the SHA256 of the manifest as the name of the manifest. This allows
    # multiple triggers to run without colliding.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp output/trusted_applet_manifest \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/trusted_applet_manifest | cut -f1 -d" ")/trusted_applet_manifest
        EOT
      ]
    }
    # Sequence log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call sequence \
        --data="{
          \"entriesDir\": \"${var.entries_dir}/$(sha256sum output/trusted_applet_manifest | cut -f1 -d" ")\",
          \"origin\": \"${var.origin_prefix}/${var.log_shard}\",
          \"bucket\": \"${var.log_name_prefix}-${var.log_shard}\",
          \"kmsKeyName\": \"ft-log-${var.env}\",
          \"kmsKeyRing\": \"firmware-release-${var.env}\",
          \"kmsKeyVersion\": ${var.log_shard},
          \"kmsKeyLocation\": \"global\",
          \"noteKeyName\": \"transparency.dev-aw-ftlog-${var.env}-${var.log_shard}\",
          \"checkpointCacheControl\": \"${var.checkpoint_cache}\"
        }"
        EOT
      ]
    }
    # Integrate log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call integrate \
        --data='{
          "origin": "${var.origin_prefix}/${var.log_shard}",
          "bucket": "${var.log_name_prefix}-${var.log_shard}",
          "kmsKeyName": "ft-log-${var.env}",
          "kmsKeyRing": "firmware-release-${var.env}",
          "kmsKeyVersion": ${var.log_shard},
          "kmsKeyLocation": "global",
          "noteKeyName": "transparency.dev-aw-ftlog-${var.env}-${var.log_shard}",
          "checkpointCacheControl": "${var.checkpoint_cache}"
        }'
        EOT
      ]
    }
    # Clean up the file we added to the _ENTRIES_DIR bucket now that it's been
    # integrated to the log.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage rm \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/trusted_applet_manifest | cut -f1 -d" ")/trusted_applet_manifest
        EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "os_build" {
  name = "os-build-${var.env}"
  location = "global"

  github {
    owner = "transparency-dev"
    name  = "armored-witness-os"

    push {
      branch = var.cloudbuild_trigger_branch != "" ? var.cloudbuild_trigger_branch : null
      tag = var.cloudbuild_trigger_tag != "" ? var.cloudbuild_trigger_tag : null
    }
  }

  build {
    # If the trigger is not based on `tag`, create a fake one.
    #
    # Unfortunately, GCB has no concept of dynamically creating substitutions or
    # passing ENV vars between steps, so the best we can do is to create a file
    # containing our tag in the shared workspace which other steps can inspect.
    step {
      name = "bash"
      script = (
        var.cloudbuild_trigger_tag != "" ?
        "$TAG_NAME > /workspace/git_tag && cat /workspace/git_tag" :
        "date +'0.3.%s-incompatible' > /workspace/git_tag && cat /workspace/git_tag"
      )
    }
    ### Build the Trusted OS and upload it to GCS.
    # Build an image containing the Trusted OS artifacts with the Dockerfile.
    # This step needs to be a bash script in order to read the tag content from file.
    step {
      name = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=$(cat /workspace/git_tag) \
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
    # Prepare a container with a copy of the artifacts.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "create",
        "--name",
        "builder_scratch",
        "builder-image",
      ]
    }
    # Copy the artifacts from the container to the Cloud Build VM.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
       "cp",
       "builder_scratch:/build/bin",
       "output",
      ]
    }
    # List the artifacts.
    step {
      name = "bash"
      script = "ls output"
    }
    # Copy the artifacts from the Cloud Build VM to GCS.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/trusted_os.elf \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/trusted_os.elf | cut -f1 -d" ")
        EOT
      ]
    }
    ### Construct log entry / Claimant Model statement.
    # This step needs to be a bash script in order to read the tag content
    # from file.
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/transparency-dev/armored-witness/cmd/manifest@main \
          create \
          --git_tag=$(cat /workspace/git_tag) \
          --git_commit_fingerprint=$COMMIT_SHA \
          --firmware_file=output/trusted_os.elf \
          --firmware_type=TRUSTED_OS \
          --tamago_version=${var.tamago_version} \
          --build_env="LOG_ORIGIN=${var.origin_prefix}/${var.log_shard}" \
          --build_env="LOG_PUBLIC_KEY=${var.log_public_key}" \
          --build_env="APPLET_PUBLIC_KEY=${var.applet_public_key}" \
          --build_env="OS_PUBLIC_KEY1=${var.os_public_key1}" \
          --build_env="OS_PUBLIC_KEY2=${var.os_public_key2}" \
          --build_env="BEE=${var.bee}" \
          --build_env="DEBUG=${var.debug}" \
          --build_env="SRK_HASH=${var.srk_hash}" \
          --raw \
          --output_file=output/trusted_os_manifest_unsigned.json
        EOT
      ]
    }
    # Sign the log entry.
    step {
      name = "golang"
      args = [
        "go",
        "run",
        "github.com/transparency-dev/armored-witness/cmd/sign@main",
        "--project_name=$PROJECT_ID",
        "--release=${var.env}",
        "--artefact=os1",
        "--manifest_file=output/trusted_os_manifest_unsigned.json",
        "--output_file=output/trusted_os_manifest_transparency_dev",
      ]
    }
    # Countersign the log entry with a second key.
    step {
      name = "golang"
      args = [
        "go",
        "run",
        "github.com/transparency-dev/armored-witness/cmd/sign@main",
        "--project_name=$PROJECT_ID",
        "--release=${var.env}",
        "--artefact=os2",
        "--note_file=output/trusted_os_manifest_transparency_dev",
        "--note_verifier=${var.os_public_key1}",
        "--output_file=output/trusted_os_manifest_both",
      ]
    }
     # Print the content of the signed manifest.
    step {
      name = "bash"
      script = "cat output/trusted_os_manifest_both"
    }
    ### Write the firmware release to the CI transparency log.
    # Copy the signed note to the sequence bucket, preparing to write to log.
    #
    # Use the SHA256 of the manifest as the name of the manifest. This allows
    # multiple triggers to run without colliding.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp output/trusted_os_manifest_both \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/trusted_os_manifest_both | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
    # Sequence log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call sequence \
        --data="{
          \"entriesDir\": \"${var.entries_dir}/$(sha256sum output/trusted_os_manifest_both | cut -f1 -d" ")\",
          \"origin\": \"${var.origin_prefix}/${var.log_shard}\",
          \"bucket\": \"${var.log_name_prefix}-${var.log_shard}\",
          \"kmsKeyName\": \"ft-log-${var.env}\",
          \"kmsKeyRing\": \"firmware-release-${var.env}\",
          \"kmsKeyVersion\": ${var.log_shard},
          \"kmsKeyLocation\": \"global\",
          \"noteKeyName\": \"transparency.dev-aw-ftlog-${var.env}-${var.log_shard}\",
          \"checkpointCacheControl\": \"${var.checkpoint_cache}\"
        }"
        EOT
      ]
    }
    # Integrate log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call integrate \
        --data='{
          "origin": "${var.origin_prefix}/${var.log_shard}",
          "bucket": "${var.log_name_prefix}-${var.log_shard}",
          "kmsKeyName": "ft-log-${var.env}",
          "kmsKeyRing": "firmware-release-${var.env}",
          "kmsKeyVersion": ${var.log_shard},
          "kmsKeyLocation": "global",
          "noteKeyName": "transparency.dev-aw-ftlog-${var.env}-${var.log_shard}",
          "checkpointCacheControl": "${var.checkpoint_cache}"
        }'
        EOT
      ]
    }
    # Clean up the file we added to the _ENTRIES_DIR bucket now that it's been
    # integrated to the log.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage rm \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/trusted_os_manifest_both | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "build_recovery" {
  name = "recovery-build-${var.env}"
  location = "global"

  github {
    owner = "transparency-dev"
    name  = "armored-witness-boot"

    push {
      branch = var.cloudbuild_trigger_branch != "" ? var.cloudbuild_trigger_branch : null
      tag = var.cloudbuild_trigger_tag != "" ? var.cloudbuild_trigger_tag : null
    }
  }

  build {
    # If the trigger is not based on `tag`, create a fake one.
    #
    # Unfortunately, GCB has no concept of dynamically creating substitutions or
    # passing ENV vars between steps, so the best we can do is to create a file
    # containing our tag in the shared workspace which other steps can inspect.
    step {
      name = "bash"
      script = (
        var.cloudbuild_trigger_tag != "" ?
        "$TAG_NAME > /workspace/git_tag && cat /workspace/git_tag" :
        "date +'0.3.%s-incompatible' > /workspace/git_tag && cat /workspace/git_tag"
      )
    }
    ### Build the recovery binary and upload it to GCS.
    # Build an image containing the trusted applet artifacts with the Dockerfile.
    step {
      name = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=ARMORY_UMS_VERSION=${var.armory_ums_version} \
          -t builder-image \
          recovery
       EOT
      ]
    }
    # Prepare a container with a copy of the artifacts.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "create",
        "--name",
        "builder_scratch",
        "builder-image",
      ]
    }
    # Copy the artifacts from the container to the Cloud Build VM.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
       "cp",
       "builder_scratch:/build/armory-ums",
       "output",
      ]
    }
    # List the artifacts.
    step {
      name = "bash"
      script = "ls output"
    }
    # Copy the artifacts from the Cloud Build VM to GCS.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/armory-ums.imx \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/armory-ums.imx | cut -f1 -d" ")
        EOT
      ]
    }
    # HAB: Create SRK table & hash
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
          go run github.com/usbarmory/crucible/cmd/habtool@c77ff4b67b3cd86b4328ecbcad23394d54638ddc \
          -z gcp \
          -1 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[1]} \
          -2 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[2]} \
          -3 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[3]} \
          -4 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[4]} \
          -o output/gcp_hab_srk.hash \
          -t output/gcp_hab_srk.srk
        EOT
      ]
    }
    # Assert SRK hash value
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        if [ -n "${var.srk_hash}"]; then \
          echo "$(od -An -tx1 output/gcp_hab_srk.hash | tr -d ' \n')" >> /workspace/got_srk_hash; \
          if [ "${var.srk_hash}" != $(cat /workspace/got_srk_hash) ]; then \
            echo "Got SRK hash \"$(cat /workspace/got_srk_hash)\""; \
            echo "Expected SRK hash '${var.srk_hash}'"; \
            exit 1; \
          fi; \
        fi
        EOT
      ]
    }
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/usbarmory/crucible/cmd/habtool@c77ff4b67b3cd86b4328ecbcad23394d54638ddc \
          -z gcp \
          -a ${data.terraform_remote_state.hab_pki.outputs.hab_csf_id} \
          -A ${data.terraform_remote_state.hab_pki.outputs.hab_csf_key} \
          -b ${data.terraform_remote_state.hab_pki.outputs.hab_img_id} \
          -B ${data.terraform_remote_state.hab_pki.outputs.hab_img_key} \
          -x 1 \
          -s \
          -t output/gcp_hab_srk.srk \
          -i output/armory-ums.imx \
          -o output/armory-ums.csf
        EOT
      ]
    }
    # Copy the HAB signature into the CAS
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/armory-ums.csf \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/armory-ums.csf | cut -f1 -d" ")
        EOT
      ]
    }
    ### Construct log entry / Claimant Model statement.
    # This step needs to be a bash script in order to read the tag content
    # from file.
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/transparency-dev/armored-witness/cmd/manifest@main \
          create \
          --git_tag=$(cat /workspace/git_tag) \
          --git_commit_fingerprint=${var.armory_ums_version} \
          --firmware_file=output/armory-ums.imx \
          --firmware_type=RECOVERY \
          --tamago_version=${var.tamago_version} \
          --hab_signature_file=output/armory-ums.csf \
          --hab_target=${var.env} \
          --raw \
          --output_file=output/recovery_manifest_unsigned.json
        EOT
      ]
    }
    # Sign the log entry.
    step {
      name = "golang"
      args = [
        "go",
        "run",
        "github.com/transparency-dev/armored-witness/cmd/sign@main",
        "--project_name=$PROJECT_ID",
        "--release=${var.env}",
        "--artefact=os1",
        "--manifest_file=output/recovery_manifest_unsigned.json",
        "--output_file=output/recovery_manifest",
      ]
    }
     # Print the content of the signed manifest.
    step {
      name = "bash"
      script = "cat output/recovery_manifest"
    }
    ### Write the firmware release to the CI transparency log.
    # Copy the signed note to the sequence bucket, preparing to write to log.
    #
    # Use the SHA256 of the manifest as the name of the manifest. This allows
    # multiple triggers to run without colliding.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp output/recovery_manifest \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/recovery_manifest | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
    # Sequence log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call sequence \
        --data="{
          \"entriesDir\": \"${var.entries_dir}/$(sha256sum output/recovery_manifest | cut -f1 -d" ")\",
          \"origin\": \"${var.origin_prefix}/${var.log_shard}\",
          \"bucket\": \"${var.log_name_prefix}-${var.log_shard}\",
          \"kmsKeyName\": \"ft-log-${var.env}\",
          \"kmsKeyRing\": \"firmware-release-${var.env}\",
          \"kmsKeyVersion\": ${var.log_shard},
          \"kmsKeyLocation\": \"global\",
          \"noteKeyName\": \"transparency.dev-aw-ftlog-${var.env}-${var.log_shard}\",
          \"checkpointCacheControl\": \"${var.checkpoint_cache}\"
        }"
        EOT
      ]
    }
    # Integrate log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call integrate \
        --data='{
          "origin": "${var.origin_prefix}/${var.log_shard}",
          "bucket": "${var.log_name_prefix}-${var.log_shard}",
          "kmsKeyName": "ft-log-${var.env}",
          "kmsKeyRing": "firmware-release-${var.env}",
          "kmsKeyVersion": ${var.log_shard},
          "kmsKeyLocation": "global",
          "noteKeyName": "transparency.dev-aw-ftlog-${var.env}-${var.log_shard}",
          "checkpointCacheControl": "${var.checkpoint_cache}"
        }'
        EOT
      ]
    }
    # Clean up the file we added to the _ENTRIES_DIR bucket now that it's been
    # integrated to the log.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage rm \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/recovery_manifest | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
  }
}

resource "google_cloudbuild_trigger" "build_boot" {
  name = "boot-build-${var.env}"
  location = "global"

  github {
    owner = "transparency-dev"
    name  = "armored-witness-boot"

    push {
      branch = var.cloudbuild_trigger_branch != "" ? var.cloudbuild_trigger_branch : null
      tag = var.cloudbuild_trigger_tag != "" ? var.cloudbuild_trigger_tag : null
    }
  }

  build {
    # If the trigger is not based on `tag`, create a fake one.
    #
    # Unfortunately, GCB has no concept of dynamically creating substitutions or
    # passing ENV vars between steps, so the best we can do is to create a file
    # containing our tag in the shared workspace which other steps can inspect.
    step {
      name = "bash"
      script = (
        var.cloudbuild_trigger_tag != "" ?
        "$TAG_NAME > /workspace/git_tag && cat /workspace/git_tag" :
        "date +'0.0.%s-incompatible' > /workspace/git_tag && cat /workspace/git_tag"
      )
    }
    ### Build the bootloader binary and upload it to GCS.
    # Use the dockerfile to build an image containing the bootloader artifact.
    step {
      name = "gcr.io/cloud-builders/docker"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        docker build \
          --build-arg=TAMAGO_VERSION=${var.tamago_version} \
          --build-arg=GIT_SEMVER_TAG=$(cat /workspace/fake_tag) \
          --build-arg=LOG_ORIGIN=${var.origin_prefix}/${var.log_shard} \
          --build-arg=LOG_PUBLIC_KEY=${var.log_public_key} \
          --build-arg=OS_PUBLIC_KEY1=${var.os_public_key1} \
          --build-arg=OS_PUBLIC_KEY2=${var.os_public_key2} \
          --build-arg=BEE=${var.bee} \
          --build-arg=CONSOLE=${var.console} \
          -t builder-image \
          .
        EOT
      ]
    }
    # Prepare a container with a copy of the artifacts.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
        "create",
        "--name",
        "builder_scratch",
        "builder-image",
      ]
    }
    # Copy the artifacts from the container to the Cloud Build VM.
    step {
      name = "gcr.io/cloud-builders/docker"
      args = [
       "cp",
       "builder_scratch:/build",
       "output",
      ]
    }
    # List the artifacts.
    step {
      name = "bash"
      script = "ls output"
    }
    # Copy the artifacts from the Cloud Build VM to GCS.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/armored-witness-boot.imx \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/armored-witness-boot.imx | cut -f1 -d" ")
        EOT
      ]
    }
    # HAB: Create SRK table & hash
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
          go run github.com/usbarmory/crucible/cmd/habtool@c77ff4b67b3cd86b4328ecbcad23394d54638ddc \
          -z gcp \
          -1 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[1]} \
          -2 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[2]} \
          -3 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[3]} \
          -4 ${data.terraform_remote_state.hab_pki.outputs.hab_srk_ca_ids[4]} \
          -o output/gcp_hab_srk.hash \
          -t output/gcp_hab_srk.srk
        EOT
      ]
    }
    # Assert SRK hash value
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        if [ -n "${var.srk_hash}"]; then \
          echo "$(od -An -tx1 output/gcp_hab_srk.hash | tr -d ' \n')" >> /workspace/got_srk_hash; \
          if [ "${var.srk_hash}" != $(cat /workspace/got_srk_hash) ]; then \
            echo "Got SRK hash \"$(cat /workspace/got_srk_hash)\""; \
            echo "Expected SRK hash '${var.srk_hash}'"; \
            exit 1; \
          fi; \
        fi
        EOT
      ]
    }
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/usbarmory/crucible/cmd/habtool@c77ff4b67b3cd86b4328ecbcad23394d54638ddc \
          -z gcp \
          -a ${data.terraform_remote_state.hab_pki.outputs.hab_csf_id} \
          -A ${data.terraform_remote_state.hab_pki.outputs.hab_csf_key} \
          -b ${data.terraform_remote_state.hab_pki.outputs.hab_img_id} \
          -B ${data.terraform_remote_state.hab_pki.outputs.hab_img_key} \
          -x 1 \
          -s \
          -t output/gcp_hab_srk.srk \
          -i output/armored-witness-boot.imx \
          -o output/armored-witness-boot.csf
        EOT
      ]
    }
    # Copy the HAB signature into the CAS
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp \
          output/armored-witness-boot.csf \
          gs://${var.firmware_bucket_prefix}-${var.log_shard}/$(sha256sum output/armored-witness-boot.csf | cut -f1 -d" ")
        EOT
      ]
    }
    ### Construct log entry / Claimant Model statement.
    # This step needs to be a bash script in order to read the tag content
    # from file.
    step {
      name = "golang"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        go run github.com/transparency-dev/armored-witness/cmd/manifest@main \
          create \
          --git_tag=$(cat /workspace/git_tag) \
          --git_commit_fingerprint=$COMMIT_SHA \
          --firmware_file=output/armored-witness-boot.imx \
          --firmware_type=BOOTLOADER \
          --tamago_version=${var.tamago_version} \
          --build_env="LOG_ORIGIN=${var.origin_prefix}/${var.log_shard}" \
          --build_env="LOG_PUBLIC_KEY=${var.log_public_key}" \
          --build_env="OS_PUBLIC_KEY1=${var.os_public_key1}" \
          --build_env="OS_PUBLIC_KEY2=${var.os_public_key2}" \
          --build_env="BEE=${var.bee}" \
          --build_env="CONSOLE=${var.console}" \
          --hab_signature_file=output/armored-witness-boot.csf \
          --hab_target=${var.env} \
          --raw \
          --output_file=output/boot_manifest_unsigned.json
        EOT
      ]
    }
    # Sign the log entry.
    step {
      name = "golang"
      args = [
        "go",
        "run",
        "github.com/transparency-dev/armored-witness/cmd/sign@main",
        "--project_name=$PROJECT_ID",
        "--release=${var.env}",
        "--artefact=boot",
        "--manifest_file=output/boot_manifest_unsigned.json",
        "--output_file=output/boot_manifest",
      ]
    }
     # Print the content of the signed manifest.
    step {
      name = "bash"
      script = "cat output/boot_manifest"
    }
    ### Write the firmware release to the CI transparency log.
    # Copy the signed note to the sequence bucket, preparing to write to log.
    #
    # Use the SHA256 of the manifest as the name of the manifest. This allows
    # multiple triggers to run without colliding.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage cp output/boot_manifest \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/boot_manifest | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
    # Sequence log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call sequence \
        --data="{
          \"entriesDir\": \"${var.entries_dir}/$(sha256sum output/boot_manifest | cut -f1 -d" ")\",
          \"origin\": \"${var.origin_prefix}/${var.log_shard}\",
          \"bucket\": \"${var.log_name_prefix}-${var.log_shard}\",
          \"kmsKeyName\": \"ft-log-${var.env}\",
          \"kmsKeyRing\": \"firmware-release-${var.env}\",
          \"kmsKeyVersion\": ${var.log_shard},
          \"kmsKeyLocation\": \"global\",
          \"noteKeyName\": \"transparency.dev-aw-ftlog-${var.env}-${var.log_shard}\",
          \"checkpointCacheControl\": \"${var.checkpoint_cache}\"
        }"
        EOT
      ]
    }
    # Integrate log entry.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud functions call integrate \
        --data='{
          "origin": "${var.origin_prefix}/${var.log_shard}",
          "bucket": "${var.log_name_prefix}-${var.log_shard}",
          "kmsKeyName": "ft-log-${var.env}",
          "kmsKeyRing": "firmware-release-${var.env}",
          "kmsKeyVersion": ${var.log_shard},
          "kmsKeyLocation": "global",
          "noteKeyName": "transparency.dev-aw-ftlog-${var.env}-${var.log_shard}",
          "checkpointCacheControl": "${var.checkpoint_cache}"
        }'
        EOT
      ]
    }
    # Clean up the file we added to the _ENTRIES_DIR bucket now that it's been
    # integrated to the log.
    step {
      name = "gcr.io/cloud-builders/gcloud"
      entrypoint = "bash"
      args = [
        "-c",
        <<-EOT
        gcloud storage rm \
        gs://${var.log_name_prefix}-${var.log_shard}/${var.entries_dir}/$(sha256sum output/boot_manifest | cut -f1 -d" ")/trusted_os_manifest_both
        EOT
      ]
    }
  }
}

# TODO(jayhou): add GCF stuff.
