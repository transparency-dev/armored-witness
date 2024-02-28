include {
  path = find_in_parent_folders()
}

terraform {
  source = "${get_path_to_repo_root()}/deployment/build_and_release/modules/release"
}

locals {
  common_vars = read_terragrunt_config(find_in_parent_folders())
}

inputs = merge(
  local.common_vars.locals,
  {
    env = "prod"
    bucket_count = 2

    cloudbuild_trigger_tag = ".*"
    build_components = {
      applet = {
        repo = "armored-witness-applet"
        cloudbuild_path = "release/cloudbuild.yaml"
      }
    }

    build_substitutions = {
      log_name = "armored-witness-firmware-log-prod-1"
      firmware_bucket = "armored-witness-firmware-prod-1"
      tamago_version = "1.21.5"
      entries_dir = "firmware-log-sequence"
      key_version = 1
      origin = "transparency.dev/armored-witness/firmware_transparency/prod/1"
      log_public_key = "transparency.dev-aw-ftlog-prod+72b0da75+Aa3qdhefd2cc/98jV3blslJT2L+iFR8WKHeGcgFmyjnt"
      applet_public_key = "transparency.dev-aw-applet-prod+d45f2a0d+AZSnFa8GxH+jHV6ahELk6peqVObbPKrYAdYyMjrzNF35"
      os_public_key1 = "transparency.dev-aw-os-prod+c31218b7+AV7mmRamQp6VC9CutzSXzqtNhYNyNmQQRcLX07F6qlC1"
      os_public_key2 = "transparency.dev-aw-os-prod-wave0+fee4bbcc+AQF1ml5TrXJkhnrJRJz5QsCZAYuCj9oOD5VpUdghWOiQ"
      bee = "1"
      debug = "1"
      checkpoint_cache = "public, max-age=30"
      srk_hash = "TODO"
    }
  }
)
