# Reproducible Build Verifier

This continuously monitors the log to look for claims about builds being published.
The log properties are checked to ensure the log is consistent with any previous
view, and that all claims are verifiably committed to by the log.

For each manifest claim that it hasn't seen before, the following steps are taken:
 1. The source repository is cloned at the git commit hash
 2. The git revision at the tag is checked against the manifest
 3. The ELF file is compiled from source
 4. The hash for the ELF in the manifest is compared against the locally built version

## Running

In order to control the environment in which the code will be built,
a Dockerfile is supplied which will create a compatible environment.

This image can be built and executed using the following commands
from the root of the repository:

```bash
docker build . -t armored-witness-build-verifier -f ./cmd/verify_build/Dockerfile
docker run armored-witness-build-verifier
```

The output with the default verbosity provides information on what is happening with
information to debug any failed builds:

```
I0219 12:23:00.470332       1 verify.go:162] Leaf index 125: ✅ reproduced build BOOTLOADER@0.0.1707929407-incompatible (373ce9ef15cc7937e1dc024a7288e4d4b1c33eab) => 7ac229b8c166d26c93006586ffb4e46a0f13c31d881fda85b09816f88c1ebc31
E0219 12:23:01.898909       1 verify.go:158] Leaf index 126: ❌ failed to reproduce build RECOVERY@0.0.0 (74060722c9aa92bbdcf3725ed0d0be4ebe8f8687) => (got b7e32298bb284c92f1cdceaffa39fa8840d21b19e4b08167023985a4a60206b2, wanted d59b4eaf94b1a895264b2a11eaee60c2a5a89c4a5b115fe29d78171187fb4df1)
I0219 12:23:01.898934       1 verify.go:94] 🔎 Evidence of failed build: /tmp/armored-witness-build-verify2690175259 (126: RECOVERY@74060722c9aa92bbdcf3725ed0d0be4ebe8f8687)
I0219 12:23:02.815427       1 verify.go:162] Leaf index 127: ✅ reproduced build RECOVERY@0.0.0 (850baf54809bd29548d6f817933240043400a4e1) => b7e32298bb284c92f1cdceaffa39fa8840d21b19e4b08167023985a4a60206b2
I0219 12:23:03.857585       1 verify.go:162] Leaf index 128: ✅ reproduced build BOOTLOADER@0.0.1707998563-incompatible (6062287365c4d7bab79532940d70d1bab846ef78) => 70d2fc2cc57de625f9a04e8923e2e0a99060c7694c3df758def16ca0f030aa4c
```

Note that in the above, the entire directory from the failed build can be obtained by:
 1. Finding the container of the verifier using `docker container ls -a` (refer to this as `DOCKER_CONTAINER`)
 2. Finding the directory name from the `🔎 Evidence of failed build` entry (refer to this as `EVIDENCE_DIR`)
 3. Copying the directory to somewhere it can be easily inspected using `docker cp $DOCKER_CONTAINER:$EVIDENCE_DIR /tmp/evidence`

To find more information about failed builds (e.g. full commandline, env variables), the verbosity can be increased by passing `--v=2` to the `docker run` command.
