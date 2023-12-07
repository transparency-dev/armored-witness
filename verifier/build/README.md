# Reproducible Build Verifier

This continuously monitors the log to look for claims about builds being published.
The log properties are checked to ensure the log is consistent with any previous
view, and that all claims are verifiably committed to by the log.

For each manifest claim that it hasn't seen before, the following steps are taken:
 1. The source repository is cloned at the release tag
 2. The git revision at the tag is checked against the manifest
 3. The imx file is compiled from source
 4. The hash for the imx in the manifest is compared against the locally built version

## Running

In order to control the environment in which the code will be built,
a Dockerfile is supplied which will create a compatible environment.

This image can be built and executed using the following commands
from the root of the repository:

```bash
docker build . -t armored-witness-build-verifier -f ./verifier/build/Dockerfile
docker run armored-witness-build-verifier -v=1
```

