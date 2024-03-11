# Transparency

The ArmoredWitness device is built to help improve the security properties of
ecosystems which rely on [transparency](https://transparency.dev) logs.

It does this by:

* **Observing public transparency logs** \
  Verifying that they're operating in an append-only fashion, and
  counter-signing those checkpoints which it has determined are consistent
  with all previous checkpoints its seen from the same log.
* **Making these counter-signed checkpoints available** \
  Sending them to a
  [distributor](https://github.com/transparency-dev/distributor), which
  then collates counter-signatures for a given checkpoint from one or more
  ArmoredWitness devices, and serves them via a public API.

The benefit of this system comes though removing trust from log operators to
behave honestly, and placing some of that trust in the witnesses. Splitting the
trust across multiple parties in this way means that a larger number of parties
must collude to hide malfeasance, and as other witness implementations/networks
start to appear, the number of parties required to collude increases
correspondingly.

However, we can minimise the amount of trust required to be placed in the
ArmoredWitness by having it be as transparent as possible too.

## Minimising trust

In order to minimise the level of trust required to be placed in the ArmoredWitness,
we have designed & implemented it such that:

* All firmware is opensource, written in [TamaGo](https://github.com/usbarmory/tamago),
  and is build-reproducible by anyone.
* All firmware is logged to a Firmware Transparency log at build and release time.
* The [`provision`](cmd/provision/) tool will only use firmware artefacts discovered
  in the FT log in order to program devices.
* The on-device self-update process requires that updated firmware is hosted in the 
  FT log.
* The boot "chain of trust" requires valid "offline FT proof bundles" to be present
  alongside the firmware at boot time:
  * The bootloader verifies signatures and FT proofs for the secure monitor ("OS"),
    and only launches it if they succeed.
  * The secure monitor ("OS") verifies signatures and FT proofs for the witness
    applet, and only launches it if they succeed.
* The [`verify`](cmd/verify) tool can be used by _custodians_ to inspect the device,
  extract the firmware components from it, and verify that they are present in the FT log.
* The [`verify_build`](cmd/verify_build) command continuously monitors the contents of
  the FT log, and tests that every logged firmware is indeed reproducibly built.

### Firmware Transparency

Every piece of non-ROM firmware for the ArmoredWitness is automatically added to our
FirmwareTransparency logs during the build & release process (e.g.
[applet GCB config](https://github.com/transparency-dev/armored-witness/blob/843b7cf1f703698de40cf82aa7d0cc38fa76859d/deployment/build_and_release/modules/release/main.tf#L245-L465)
). In conjunction with the controls in the firmware self-update, provision, and user
verification tooling mentioned above, this helps to ensure that we, or anybode else,
would find it very difficult to covertly target all, or a subset of the devices,
with malicious firmware.

Rather than writing compiled blobs directly into the log, we store a _manifest_, which
commits to all necessary inputs to the build, and the corresponding firmware output.

Here is a [real manifest from the CI log](https://storage.googleapis.com/armored-witness-firmware-log-ci-2/seq/00/00/00/00/2f):

```
{
  "schema_version": 0,
  "component": "TRUSTED_APPLET",
  "git": {
    "tag_name": "0.3.1709910063-incompatible",
    "commit_fingerprint": "9651fc25839d9937acc041057cf3906f26fc1ae5"
  },
  "build": {
    "tamago_version": "1.22.0",
    "envs": [
      "FT_LOG_URL=https://api.transparency.dev/armored-witness-firmware/ci/log/2",
      "FT_BIN_URL=https://api.transparency.dev/armored-witness-firmware/ci/artefacts/2",
      "LOG_ORIGIN=transparency.dev/armored-witness/firmware_transparency/ci/2",
      "LOG_PUBLIC_KEY=transparency.dev-aw-ftlog-ci-2+f77c6276+AZXqiaARpwF4MoNOxx46kuiIRjrML0PDTm+c7BLaAMt6",
      "APPLET_PUBLIC_KEY=transparency.dev-aw-applet-ci+3ff32e2c+AV1fgxtByjXuPjPfi0/7qTbEBlPGGCyxqr6ZlppoLOz3",
      "OS_PUBLIC_KEY1=transparency.dev-aw-os1-ci+7a0eaef3+AcsqvmrcKIbs21H2Bm2fWb6oFWn/9MmLGNc6NLJty2eQ",
      "OS_PUBLIC_KEY2=transparency.dev-aw-os2-ci+af8e4114+AbBJk5MgxRB+68KhGojhUdSt1ts5GAdRIT1Eq9zEkgQh",
      "REST_DISTRIBUTOR_BASE_URL=https://api.transparency.dev/ci",
      "BEE=1",
      "DEBUG=1",
      "SRK_HASH=b8ba457320663bf006accd3c57e06720e63b21ce5351cb91b4650690bb08d85a"
    ]
  },
  "output": {
    "firmware_digest_sha256": "lLPLT5TO2+Ln71cByKhVvNFyAL47IzOOSGoXNKVSCvU="
  }
}

— transparency.dev-aw-applet-ci P/MuLOfW8473+PNMa58SZA2/rw1aEaIaLTw/aNfdawSiyFEcDjGksYqCTFMnHHGAhhbfnITkkktL1We6UF3VMuHakwU=
```

The full description of this structure is avilable in the source
[here](https://github.com/transparency-dev/armored-witness-common/blob/main/release/firmware/ftlog/log_entries.go#L32),
but, broadly, the `component` and `git` fields tell us that the build is for the `APPLET` firmware
type, and was done:

* at `github.com/transparency-dev/armored-witness-applet@9651fc25839d9937acc041057cf3906f26fc1ae5`
* using TamaGo `v1.22.0`
* with the set of environment variables in `env` declared.

and that the resulting firmware binary has the SHA256 hash in the `firmware_digest_sha256` field.

Since we're staking our reputation on this _claim_ being true, the line starting "— transparency.dev-aw-applet-ci P/Mu..."
is a signature from our CI build robot committing to it.

The actual binary itself is stored in a content addressable store (CAS) served adjacent to the log itself (in fact,
the URL to the root of that CAS is present in the `env` section under the `FT_BIN_URL` variable), keyed by the _hex_
encoded firmware hash (rather than the base64 encoding the JSON stores).

If you wish, you could download that firmware image and verify its hash with the following commands:

```bash
$ SHA256HEX=$(echo -n "lLPLT5TO2+Ln71cByKhVvNFyAL47IzOOSGoXNKVSCvU=" | base64 -d | hexdump -v -e '/1 "%02x" ')
$ echo ${SHA256HEX}
94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5
$ wget https://api.transparency.dev/armored-witness-firmware/ci/artefacts/2/${SHA256HEX}
--2024-03-08 15:53:22--  https://api.transparency.dev/armored-witness-firmware/ci/artefacts/2/94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5
Resolving api.transparency.dev (api.transparency.dev)... 2600:1901:0:499e::, 34.36.253.177
Connecting to api.transparency.dev (api.transparency.dev)|2600:1901:0:499e::|:443... connected.
HTTP request sent, awaiting response... 200 OK
Length: 16961249 (16M) [application/octet-stream]
Saving to: ‘94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5.1’

94b3cb4f94cedbe2e7ef5701c8a855bcd1 100%[==============================================================>]  16.17M  70.1MB/s    in 0.2s

2024-03-08 15:53:22 (70.1 MB/s) - ‘94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5’ saved [16961249/16961249]
$ sha256sum 94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5
94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5  94b3cb4f94cedbe2e7ef5701c8a855bcd17200be3b23338e486a1734a5520af5
```

#### Firmware Transparency verifier

With the information above, you can check out the [applet repo](https://github.com/transparency-dev/armored-witness-applet),
at the particular commit, install the correct version of `TamaGo`, set up the environment variables, run `make imx` and expect it
to create a `trusted_applet.imx` file which should have the hash as above.

We've built a tool to do this for you though, it's in this repo and is called [`verify_build`](cmd/verify_build).
Instructions on running it are in that directory, but in summary it will "tail" the FT log, and for each entry it finds
there it'll attempt the above and report on whether it was able to successfully reproduce the build and get a matching
hash from the output file.

#### Firmware Transparency enforcement

Even with many people running the `verify_build` tool above, it's no use if the device installation & update
process allows firmware which is not in the log to be installed on the device.

##### Initial provisioning

The [`provision`](/cmd/provision) tool, which is what we use to first turn blank devices into locked-down
ArmoredWitnesses, has no capability to fetch firmware (including the
[`recovery`](https://github.com/usbarmory/armory-ums) firmware used as part of the installation process)
from anywhere other than the FT log. Before attempting to install any of the firmware, it also checks that
the firmware hash matches the expected value in the manifest.

##### On-device - self-update

The device has a
[`self-update`](https://github.com/transparency-dev/armored-witness-common/blob/main/release/firmware/update/update.go)
component, which, similarly to the `provision` tool, is built to only use the FT log as the source for updates.

If a newer `APPLET` or `OS` component than is currently installed is found (note that the updater cannot update
the `bootloader`), then the updater builds a
[`ProofBundle`](https://github.com/transparency-dev/armored-witness-common/blob/main/release/firmware/bundle.go#L19-L38)
before submitting it to the `OS` via RPC to be installed.

The `ProofBundle` contains:

* An FT log [`Checkpoint`](https://github.com/transparency-dev/formats/blob/main/log/README.md) which
  commits to the state of the log in which the update was found.
* The `Manifest` for the firmware build found.
* The `Index` of the `Manifest` in the log.
* The `InclusionProof` of the `Manifest` under the `Checkpoint`.
* The `Firmware` binary.

Before flashing the updated firmware onto the device, the OS
[checks whether the proof bundle is valid](https://github.com/transparency-dev/armored-witness-common/blob/main/release/firmware/verify.go#L41-L73):

1. Verify the signature on the `Checkpoint` is correct and from the expected log signer, and that the `Checkpoint` format is correct.
1. Verify that the `InclusionProof` for the `Manifest` at `Index` does reproduce the root hash the `Checkpoint` commits to.
1. Verify that the SHA256 of the `FirmwareImage` matches the `firmware_digest_sha256` value in the `Manifest`.

If any of these checks fail, the update is rejected.

##### On-device - boot

Both the `provision` and `self-update` component store the `ProofBundle` data on the MMC at the same time as the
firmware is written.

The `bootloader` verifiers the `ProofBundle` for the `OS` before launching it. Similarly, the `OS` verifies the
`ProofBundle` for the `Applet` before launching that.

This, coupled with:

* The self-update being unable to upgrade the bootloader,
* The requirement to manually flip a DIP switch on the bottom of the device to boot the `recovery` image
  means that only firmware present in the FT can boot on the device without the custodian being aware that
  something untoward has happened.

##### Manual

Of course, we _could_ have a `SeKrEt EvIl HaX0r` version of the `provision` tool that we didn't opensource,
in which all these checks are defeated, and which will install corresponding `SeKrEt EvIl` firmware versions
that will accept non-FT updates too.

To make sure we'd be caught if we did that, we've provided the [`verify`](/cmd/verify) tool.

This tool uses the [`recovery`](https://github.com/usbarmory/armory-ums) firmware from the FT log to expose
the device's MMC storage as a USB Mass Storage device, and then extracts each of the 3 types of firmware
and their `Manifest`s stored on there, before verifying the firmware hashes against the ones in the `Manifest`s
and finally checking with the FT log that all 3 `Manifest`s are present there.

TODO(al): redo this

![verify](/docs/images/verify.svg)
