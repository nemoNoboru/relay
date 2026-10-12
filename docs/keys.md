# Ownership Keys

**A key is a deed.** An address on Relay — `/<node>/<site>/` — is owned by the
Ed25519 key that was registered for it, not by an account, an email, or a
platform's permission. Hold the secret half of the key and you control the
address; lose it and the address is gone.

This document covers what is built and what is decided. Locks come from
[`REVIEW.md`](../REVIEW.md); the code lives in
[`internal/domain/key`](../internal/domain/key/key.go).

## The two halves

Every key is an Ed25519 keypair:

| Half | Size | Where it lives | Used for |
|---|---|---|---|
| **Public** | 32 bytes | registered on the address, free to share | verifying signatures |
| **Secret** | 32 bytes (the seed) | only with the key holder | making signatures |

Why Ed25519: small keys and signatures, fast on weak hardware (the friend node
is a Raspberry Pi), constant-time in Go's standard library, and no hash
preprocessing to get wrong. The public half can be derived from the secret half
— that is why a key file only ever stores the secret.

## Proving ownership without sending the secret

The secret never travels. When the owner (the deploy tool, the node UI, a
future library) makes an owner request, it:

1. builds a **canonical string** for the request,
2. signs it with the secret half,
3. sends the request, the public key (or its fingerprint), the timestamp, and
   the signature.

The node verifies the signature against the public key registered on the
address. Timestamp + nonce make a captured request un-replayable.

```text
owner                         node
  |  request + ts + nonce        |
  |  sig = sign(canonical string)|
  | ---------------------------> |  verify sig against registered public key
  |                               |  check ts fresh, nonce unused
  |                               |  check public key matches the address
  | <--------------------------- |  accept / reject
```

The domain package signs and verifies the raw bytes
(`Key.Sign`, `key.Verify`); the canonical-string rules belong to the layer that
builds owner requests, so they stay visible and reviewable in one place.

## Fingerprints

A **fingerprint** is the first 16 hex characters of the SHA-256 of the public
half — 8 bytes, enough to tell keys apart without ambiguity in practice, and
stable for the life of the key.

Fingerprints are the *only* part of a key that is safe to log, show in the UI,
or put in a commit. They identify a key; they do not authenticate as one.
Never log, print, or paste the public or secret half itself.

```text
fingerprint 3f199360809efaf9   ← safe to show anywhere
secret PvUde0J...              ← never
```

## The key file

`relay key new` produces a key file, format version 1:

```text
relay key v1
fingerprint 3f199360809efaf9
secret PvUde0JmiH9WlFrrEpF23aHfRpvjbDdyaQG5s3qRZzo=
```

Example above: a real generated key, now discarded — do not use it.

Three lines, and that is the whole format:

| Line | Meaning |
|---|---|
| `relay key v1` | magic + version. A different magic or version is rejected — a key file is never mistaken for another file type, and old files do not silently load as new ones. |
| `fingerprint …` | the fingerprint of the key, in the file so the file is **self-validating**. |
| `secret …` | the 32-byte seed, base64. The seed is the key: it deterministically regenerates the whole keypair. |

Loading checks, in order: shape (exactly 3 lines, exact header), fingerprint
shape (16 lowercase hex), secret shape (base64, exactly 32 bytes), and then the
cross-check — the fingerprint must match the secret. A truncated, corrupted,
or swapped file fails; you get an `invalid key file` error, never a silently
wrong key.

## Using keys

```sh
relay key new                 # print the key file to stdout
relay key new -o ./deed.key   # write it to a file (mode 0600)
```

- Without `-o` the file goes to stdout, so `relay key new > deed.key` works.
- With `-o` the file is written `0600` (owner read/write only) and the
  fingerprint is printed to stderr — the key material never leaks into logs
  that also capture stdout.
- The command refuses to export a key that holds only public material, so a
  key file always contains what its name promises.

Tests cover the round trip (`File` → disk → `Load`), the cross-check
(corrupt fingerprint, wrong version, bad base64, wrong seed length), and
sign/verify (`internal/domain/key/key_test.go`).

## Guarding a key

Treat a key file like a password that is also a property title:

- **Keep it 0600.** `relay key new -o` does this; if you create the file
  yourself, do the same.
- **Never commit it.** Not to git, not to a "backup" repo, not to a paste
  bin. The public half and the fingerprint may go anywhere.
- **Back it up** somewhere you control and can recover — paper, an encrypted
  stash, a password manager. There is no recovery path: no server holds the
  secret, so it cannot be reset.
- **Move a site** by copying the site's files *and* the key
  ([`docs/technicals.md`](./technicals.md)): the unit of migration is the
  folder plus its deed.

## What a key can do

Built now (scaffold):

- generate a keypair,
- sign and verify,
- fingerprint,
- render and load the versioned, self-validating key file.

Decided, coming with the node runtime ([`REVIEW.md`](../REVIEW.md)):

- **claim and hold a name** — the public key registered for `/<node>/<site>/`
  is the address's deed;
- **deploy, roll back, delete** — owner requests, signed as above;
- **key login** — optional Ed25519 login that replaces passwords for a site;
- **tunnel authentication** — friend nodes authenticate to public nodes by key.

## Not solved yet

Key lifecycle is an open question (REVIEW.md §7.14): **rotation, loss, and
release** of a key are not designed yet. Until they are: one key per address,
no recovery, and rotating means claiming a new address with a new key and
migrating the files.
