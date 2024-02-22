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
    log_count = 1

    cloudbuild_trigger_tag = ".*"
    build_components = {
      applet = {
        repo = "armored-witness-applet"
        cloudbuild_path = "release/cloudbuild.yaml"
      }
    }
  }
)
