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
    bucket_env = ""
    cloudbuild_path = "release/cloudbuild.yaml"
    cloudbuild_branch = ""
    cloudbuild_tag = ".*"
  }
)
