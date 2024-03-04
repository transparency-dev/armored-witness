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
    env = "ci"
    bucket_count = 3

    cloudbuild_trigger_branch = "^main$"
    build_components = {
      applet = {
        repo = "armored-witness-applet"
        cloudbuild_path = "release/cloudbuild_ci.yaml"
      }
    }

    log_shard = 2
    log_name_prefix = "armored-witness-firmware-log-ci"
    firmware_bucket_prefix = "armored-witness-firmware-ci"
    origin_prefix = "transparency.dev/armored-witness/firmware_transparency/ci"

    tamago_version = "1.21.5"
    entries_dir = "firmware-log-sequence"
    log_public_key = "transparency.dev-aw-ftlog-ci-2+f77c6276+AZXqiaARpwF4MoNOxx46kuiIRjrML0PDTm+c7BLaAMt6"
    applet_public_key = "transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3"
    os_public_key1 = "transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ"
    os_public_key2 = "transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh"
    bee = "1"
    debug = "1"
    checkpoint_cache = "public, max-age=30"

    # HAB-related
    srk_hash = "b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a"
    hab_key_version = 1

    # Pinned at tag [v20231018](https://github.com/usbarmory/armory-ums/releases/tag/v20231018)
    armory_ums_version: "850baf54809bd29548d6f817933240043400a4e1"
  }
)
