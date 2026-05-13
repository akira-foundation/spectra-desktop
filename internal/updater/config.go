package updater

// ManifestURL is the public CDN URL for the release manifest.
const ManifestURL = "https://akira-spectra.nyc3.cdn.digitaloceanspaces.com/releases/latest.json"

// MinisignPublicKey is the minisign ed25519 public key used to verify update artifacts.
// Replace with the real pubkey generated via `minisign -G`.
const MinisignPublicKey = "RWRGgIZ0jLok7dGc+b1K1DX/zTRPDvlfR2i22XcbiTIdp4O3zlv0gJkC"
