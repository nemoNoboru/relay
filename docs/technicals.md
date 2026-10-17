# Relay Technicals

## How It Works and How to Deploy It

### How it works

Relay is a **single Go binary**; every node runs the same code, and the only difference is whether a node is public. Every site lives at a **path** under one host — one hostname, one TLS certificate, no wildcards:

```
https://relay.example.org/<node>/<site>/
```

A machine behind NAT cannot accept inbound connections, so a private node opens one **outbound WebSocket** to a public friend and registers its path prefixes. The public node routes each request by its first path segment down the matching tunnel, and the answer travels back up the same wire. No port forwarding, no static IP, no dynamic DNS.

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

All node state — uploaded sites, the deploy-key store, the admin token — lives in **one data directory** you own. Mount it on a volume; without it, the node forgets everything.

**Built today:** static hosting, per-user deploy keys, atomic deploys with rollback, the operator console, native HTTPS. **On the road:** NAT tunnels, one SQLite file per site, opt-in login and stats, sandboxed functions. The full design review is in [`REVIEW.md`](./REVIEW.md).

### The stack

| Concern | Choice | Why |
|---|---|---|
| Language | **Go** | one static binary, cheap concurrency |
| Addressing | **paths under one host** | one cert, identical code public and private |
| Tunnels | **outbound WebSocket** (yamux) | no inbound ports on private nodes |
| Storage | **files now, SQLite per site next** | isolation, export, migration as a single file |
| Keys | **Ed25519** | signed challenges; the secret never travels |

### Deploy a site

Uploading is one POST of a `tar` archive with a deploy key — the flow is on the [landing page](./index.html#deploy), in the [user's guide](./users-guide.md), and hosted on every node at `/_relay/deploy.html`.

### Run a node

Relay is one binary, so it runs anywhere a process or a container runs. **Public nodes** need a stable address and ports 80/443; **private nodes** are outbound-only and need no inbound ports at all. On first start the node writes its **admin token** to `<data>/admin-token` (and prints it once to the log) — it unlocks the operator console at `/_relay/admin.html`.

#### Binary

Install the prebuilt binary (Linux or macOS, amd64/arm64):

```sh
curl --proto '=https' --tlsv1.2 -sSfL \
  https://raw.githubusercontent.com/nemoNoboru/relay/main/install.sh | sh
```

Then run a node:

```sh
relay node --data /var/lib/relay --listen :8080
```

#### Docker

The image is published to GitHub Packages, runs unprivileged, and keeps all state under `/data`:

```sh
docker run -d --name relay --restart unless-stopped \
  -p 8080:8080 \
  -v relay-data:/data \
  ghcr.io/nemonoboru/relay:latest
```

#### Kubernetes

One command installs a `relay` namespace, a one-replica StatefulSet with a persistent volume, and a Service:

```sh
kubectl apply -f https://raw.githubusercontent.com/nemoNoboru/relay/main/deploy/kubernetes.yaml
```

The node serves HTTP on `:8080` with probes on `/healthz`. Expose it with TLS at the cluster edge by filling in your hostname in [`deploy/ingress.yaml`](./deploy/ingress.yaml) and applying it. Because Relay uses path-based addressing, the Ingress needs no wildcard host and no per-site certificate. Read the admin token with either of:

```sh
kubectl logs -n relay relay-0
kubectl exec -n relay relay-0 -- cat /data/admin-token
```

#### HTTPS

Pass `--tls-cert` and `--tls-key` to serve HTTPS directly, or terminate TLS at a reverse proxy (Caddy, Nginx) in front of the node.

### Further reading

- [`REVIEW.md`](./REVIEW.md) — the consolidated design review: decisions, open questions, what is not finished
- [`docs/api-contract.md`](./docs/api-contract.md) and [`docs/architecture.md`](./docs/architecture.md) — the wire contract and the backend design
- [`docs/keys.md`](./docs/keys.md) — ownership keys: the deed to an address
