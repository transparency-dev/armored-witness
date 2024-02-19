include {
  path = find_in_parent_folders()
}

terraform {
  source = "."
}

locals {
  common_vars = read_terragrunt_config(find_in_parent_folders())
}

inputs = merge(
  local.common_vars.locals,
  {
    env = "presubmit"
  }
)
