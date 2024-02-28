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
    firmware_bucket = string
    tamago_version = string
    entries_dir = string
    key_version = number
    origin = string
    log_name = string
    log_public_key = string
    applet_public_key = string
    os_public_key1 = string
    os_public_key2 = string
    bee = string
    debug = string
    checkpoint_cache = string
    srk_hash = string
  })
}

