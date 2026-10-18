# Relay Technicals

## How It Works and How to Deploy It

### How it works

Relay is a **single Go binary**. Every node runs the same code. Sites live at **paths** under one host, so a node needs one hostname and one TLS certificate, with no wildcards and no per-site certs:

```
https://relay.example.org/<node>/<site>/
```

Uploading a site is one `POST` of a `tar` archive, authenticated with a per-user **deploy key**. The upload is atomic: the new version replaces the old one all at once, and the previous 10 are kept for rollback. Everything a node knows (sites, versions, keys, the admin token) lives in **one data directory**. Back it up and you've backed up the node.

### What's built

| Piece | Choice | Status |
|---|---|---|
| Static hosting | paths under one host | **live** |
| Uploads | `tar` + per-user deploy key | **live** |
| Versions | atomic deploys, last 10 kept, rollback | **live** |
| Operator console | issue, reveal, revoke, rotate keys; sites by user | **live** |
| HTTPS | native (`--tls-cert`/`--tls-key`) or a reverse proxy | **live** |
| Reachability | outbound WebSocket tunnel to a public friend | planned |
| Ownership | Ed25519 key as the deed to an address; signed deploys | planned |
| Site data | one SQLite file per site, opt-in login and stats | planned |

### Reachability (planned — not built yet)

A machine behind NAT can't accept inbound connections. So a private node will dial out: it opens one **outbound WebSocket** to a public friend and registers its path prefix. The public node routes each request by its first path segment down the matching tunnel. No port forwarding, no static IP, no dynamic DNS.

```
        Visitors
            │  relay.example.org/raspberry/docs/
            ▼
   ┌──────────────────────────────┐
   │         Public node          │
   │      relay.example.org       │
   │                              │
   │  friend list:                │
   │    /raspberry/ ──────────────┼──┐
   │    /bookshelf/ ──────────────┼┐ │
   └──────────────────────────────┘│ │
                                   │ │  outbound WebSocket
                                   ▼ ▼
                        ┌───────────┐  ┌───────────┐
                        │ Raspberry │  │ Laptop    │
                        │ Pi behind │  │ behind    │
                        │ NAT       │  │ NAT       │
                        └───────────┘  └───────────┘
```

### Keep a server

Every node needs a **name**, the first path segment of its sites (`--name mynode` serves `/mynode/<site>/`). On first start the node writes an **admin token** to `<data>/admin-token` and prints it once to the log. Use it to sign in to the operator console at `/_relay/admin.html` and issue deploy keys.

#### Binary

Linux or macOS, amd64 or arm64:

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh

relay node --name mynode --data /var/lib/relay --listen :8080
```

#### Docker

The image runs unprivileged and keeps all state in `/data`. Flags after the image name are passed to `relay node`:

```sh
docker run -d --name relay --restart unless-stopped \
  -p 8080:8080 -v relay-data:/data \
  ghcr.io/nemonoboru/relay:latest --name mynode
```

Read the admin token with `docker logs relay`.

#### Kubernetes

One command installs a `relay` namespace, a single-replica StatefulSet with a persistent volume, and a Service. The node is named `relay`, so sites live at `/relay/<site>/`:

```sh
kubectl apply -f https://raw.githubusercontent.com/nemoNoboru/relay/main/deploy/kubernetes.yaml
kubectl exec -n relay relay-0 -- cat /data/admin-token
```

The node serves HTTP on `:8080` with probes on `/healthz`. For TLS, put your hostname in [`deploy/ingress.yaml`](https://github.com/nemoNoboru/relay/blob/main/deploy/ingress.yaml) and apply it. Path-based addressing means one host and one certificate cover every site.

#### HTTPS

Pass `--tls-cert` and `--tls-key` to serve HTTPS directly, or put Caddy or Nginx in front. Testing with a self-signed certificate? Add `--insecure` to `relay deploy` and `relay rollback`.

### Further reading

- [Source on GitHub](https://github.com/nemoNoboru/relay): code, issues, and the [user's guide](https://github.com/nemoNoboru/relay/blob/main/users-guide.md)
- [`REVIEW.md`](https://github.com/nemoNoboru/relay/blob/main/REVIEW.md): the design review, with locked decisions and open questions
- [API contract](https://github.com/nemoNoboru/relay/blob/main/docs/api-contract.md) and [backend architecture](https://github.com/nemoNoboru/relay/blob/main/docs/architecture.md): the wire protocol and how the node is built
- [Ownership keys](https://github.com/nemoNoboru/relay/blob/main/docs/keys.md): how a key becomes the deed to an address
