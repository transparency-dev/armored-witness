include {
  path = find_in_parent_folders()
}

terraform {
  source = "${get_path_to_repo_root()}/deployment/hab_pki/modules/hab_pki"
}

locals {
  common_vars = read_terragrunt_config(find_in_parent_folders())
}

inputs = merge(
  local.common_vars.locals,
  {
    env = "wave0"
    hab_revision = 0
    hab_leaf_revision = 0
  }
)
