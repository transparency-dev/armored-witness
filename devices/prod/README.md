# Prod devices

The files contained in this directory hold identity information about _production fused_ witness devices.

End-users are not expected to use these files, they're just being stored here.

## Files

File names are based on the unique and immutable serial number assigned at time of manufacturing of the SoC
which lies at the core of the device.

The `.pub` file contains a public key which represents the _device_ with the corresponding serial number.

The `.witness.0` file contains a signed note, verifiable with the device's `pub` key, which commits to the
initial witness identity used by the device to cosign checkpoints.

The body of this note is formed of 4 lines:

1. A line with the text "ArmoredWitness ID attestation v1".
2. The ASCII encoded HEX representation of the device serial number whose witness public key is below.
3. The ASCII encoded decimal number 0.
4. A note Verifier string representing the witness public key which will be used by this device.

The note is signed by the _device_ key corresponding to the serial number on line 2.
