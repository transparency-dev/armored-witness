# Log Rotation

This doc outlines how to rotate the logs detailed in [transparency.md](transparency.md).

## Steps

1. [Rotate](https://cloud.google.com/kms/docs/rotate-key#manual) the log key to create a new key version. Note: this can be auotmated in Terraform, but we do not currently manage the key resources in Terraform.

1. Increment the `bucket_count` to create new firmware artefact and log buckets in GCS, `log_shard` to start writing to the new buckets, and `log_public_key` to match the new keys.
    * Example [PR](https://github.com/transparency-dev/armored-witness/pull/182/)
    * The `log_public_key` needs to be generated from the [public key](https://cloud.google.com/kms/docs/retrieve-public-key) of the new key version.
    * After applying the [`build_and_release`](/deployment/build_and_release) Terraform changes, the output values will be used for new paths at https://api.transparency.dev (step #4).

1. Initialize the log bucket:
    ```
    gcloud functions call integrate \
    --data '{
        "initialise": true,
        "origin": "transparency.dev/armored-witness/firmware_transparency/$ENV/$LOG_SHARD",
        "bucket": "armored-witness-firmware-log-$ENV-$LOG_SHARD",
        "kmsKeyName": "ft-log-$ENV",
        "kmsKeyRing": "firmware-release-$ENV",
        "kmsKeyVersion": $LOG_SHARD,
        "kmsKeyLocation": "global",
        "noteKeyName": "transparency.dev-aw-ftlog-$ENV-$LOG_SHARD"
    }'
    ```

1. Apply [`api_transparency_dev`](/deployment/api_transparency_dev) Terraform to create new https://api.transparency.dev paths for the firmware artefacts and log.

1. Populate the buckets by running the build triggers.
    * For CI:
      ```
      gcloud builds triggers run applet-build-ci --branch=main
      gcloud builds triggers run os-build-ci --branch=main
      gcloud builds triggers run boot-build-ci --branch=main
      gcloud builds triggers run recovery-build-ci --branch=main
      ```
    * For prod, create a new release on the Github repo.

### Update dependencies
1. Update the template used by `verify` and `provision` tools. Example [PR](https://github.com/transparency-dev/armored-witness/pull/186).
1. Add the log to the omniwitness config. Example [PR](https://github.com/transparency-dev/witness/pull/175).
1. Add the log to the distributor config. Example [PR](https://github.com/transparency-dev/distributor/pull/131).