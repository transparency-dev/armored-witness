variable "project_id" {
  description = "The project ID to host the cluster in"
  type        = string
}

variable "env" {
  description = "Unique identifier for the env, e.g. ci or prod"
  type        = string
}

variable "bucket_count" {
  description = "The number of log and firmware buckets to create (each)."
  type        = number
}

variable "signing_keyring_location" {
  description = "The GCP location to create the signing keyring"
  type        = string
}

variable "build_components" {
  type = map(object({
    repo            = string
    cloudbuild_path = string
  }))
}

variable "cloudbuild_trigger_tag" {
  description = <<EOH
    Specifies how the build will be triggered. Exactly one of cloudbuild_trigger_branch or cloudbuild_trigger_tag should be specified.
    See more: https://cloud.google.com/build/docs/api/reference/rest/v1/projects.locations.triggers#BuildTrigger.PushFilter.
  EOH
  type        = string
  default     = ""
}

variable "cloudbuild_trigger_branch" {
  description = <<EOH
    Specifies how the build will be triggered. Exactly one of cloudbuild_trigger_branch or cloudbuild_trigger_tag should be specified.
    See more: https://cloud.google.com/build/docs/api/reference/rest/v1/projects.locations.triggers#BuildTrigger.PushFilter.
  EOH
  type        = string
  default     = ""
}

variable "build_substitutions" {
  type = object({
    log_name = string
    firmware_bucket = string
    tamago_version = string
    entries_dir = string
    # This must correspond with the trailing number on the firmware_bucket,
    # origin, log_name values.
    key_version = number
    origin = string
    log_public_key = string
    applet_public_key = string
    os_public_key1 = string
    os_public_key2 = string
    bee = string
    debug = string
    checkpoint_cache = string
    # Pinned CI SRK hash
    # This MUST be identical to the _PINNED_SRK_HASH in
    # https://github.com/transparency-dev/armored-witness-boot/blob/main/release/cloudbuild_ci.yaml#L223-L224
    # and MUST NOT be changed unless you know very well what you're doing,
    # otherwise devices will be bricked!
    srk_hash = string
  })
}

