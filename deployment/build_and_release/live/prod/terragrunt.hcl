include "root" {
  path   = find_in_parent_folders()
  expose = true
}

terraform {
  source = "${get_repo_root()}/deployment/build_and_release/modules/release"
}

inputs = merge(
  include.root.locals,
  {
    bucket_count = 2

    # Require "v" in front of the regex from
    # https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
    cloudbuild_trigger_tag = "^v(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"

    log_shard = 1
    log_name_prefix = "armored-witness-firmware-log-prod"
    firmware_bucket_prefix = "armored-witness-firmware-prod"
    origin_prefix = "transparency.dev/armored-witness/firmware_transparency/prod"
    
    tamago_version = "1.23.1"
    log_public_key = "transparency.dev-aw-ftlog-prod-1+3e6d87ee+Aa3qdhefd2cc/98jV3blslJT2L+iFR8WKHeGcgFmyjnt"
    applet_public_key = "transparency.dev-aw-applet-prod+d45f2a0d+AZSnFa8GxH+jHV6ahELk6peqVObbPKrYAdYyMjrzNF35"
    os_public_key1 = "transparency.dev-aw-os1-prod+985bdfd2+AV7mmRamQp6VC9CutzSXzqtNhYNyNmQQRcLX07F6qlC1"
    os_public_key2 = "transparency.dev-aw-os2-prod+662add8c+AebLJIKJhx57T3mWmHKe0sasFnXmtIQNTGRaoj2PQLrY"
    bee = "1"
    debug = "1"

    bastion_addr = "bastion.glasklar.is:443"

    # HAB-related
    srk_hash = "77e021cc51b5547fb0c2192fb32710bfa89b4bbaa7dab5f97fc585f673b0b236"

    # Pinned at tag [v20231018](https://github.com/usbarmory/armory-ums/releases/tag/v20231018)
    armory_ums_version: "850baf54809bd29548d6f817933240043400a4e1"

    verify_template = "prod"
  }
)
