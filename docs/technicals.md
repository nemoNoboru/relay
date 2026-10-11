# Relay Technicals

## How It Works and How to Deploy It

### How Relay Is Built

Relay is a **single Go binary** that runs as either a public node or a private node — the same code, one flag. It leans on proven parts instead of inventing new ones.

| Concern | Choice | Why |
|---|---|---|
| Language | **Go** | one static binary; Caddy is embeddable; goroutines for many tunnels |
| TLS + HTTP | **Caddy** (embedded) | one certificate, HTTP/2+3, reverse proxy, static files |
| Addressing | **paths under one host** | one cert, no wildcards, identical code public and private |
| Storage | **SQLite**, one file per site | isolation, export, and migration as a single file |
| Compute sandbox | **QuickJS** via Wazero | tiny, memory- and CPU-cappable, speaks JavaScript |
| Keys | **Ed25519** | signed challenges; the secret never travels |
| Federation | a published **registry list** | discovery without an authority |

The reasoning behind each is in [`REVIEW.md`](./REVIEW.md).

### Addressing

Every site lives at a **path** under one host:

```
https://relay.cloudpunk.org/<node>/<site>/
```

- One hostname per public node, **one TLS certificate**, renewed automatically.
- No wildcards, no per-site certificates, no on-demand TLS.
- The router reads the first path segment, finds that node's tunnel, and forwards the rest.
- Reserved paths: `/_relay/` for the control API; `_relay` inside a site for its runtime API.

**The trade-off, stated plainly.** Sites share one browser origin, so the browser does not isolate them from each other. Relay compensates in code: path-scoped cookies (`SameSite=Strict`, `HttpOnly`, `Secure`), a per-site CSRF token on mutations, and a `Referer`-prefix check. Absolute-path assets are handled with a `<base>` prefix or the `X-Forwarded-Prefix` header. The full analysis is in [`REVIEW.md`](./REVIEW.md).

### Reachability

A machine behind NAT cannot accept inbound connections. Relay fixes this with one rule: **every node runs the same code; the only difference is whether the node is public.**

```
        Visitors
            │  relay.cloudpunk.org/raspberry/docs/
            ▼
   ┌──────────────────────────────┐
   │         Public node          │
   │      relay.cloudpunk.org     │
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

- A private node opens an **outbound WebSocket** to a public friend: no port forwarding, no static IP, no dynamic DNS, no firewall rule.
- It registers the **path prefixes** it serves (`/raspberry/`). The public node keeps the list and hands back nothing but a route.
- Requests are relayed by the first path segment down the matching tunnel; responses travel the same way.
- A node may hold several public friends and fail over between them, ordered by round-trip time.

The same tunnel also fronts an app you already run — Relay as a tunnel-only layer. Details are in [`REVIEW.md`](./REVIEW.md).

### Storage

- **One SQLite file per site.** Natural isolation, trivial export, no cluster to run.
- Static files are stored as content-addressed blobs (BLAKE3) and deduplicated on deploy.
- The unit of migration is the file. Moving a site is copying its file and keeping its key.

### Login and Stats

Both are **opt-in and per site**, and both are rows in the site's own database — there is no network-level account.

- Users and sessions live inside the site's SQLite file.
- Passwords are hashed with **Argon2id**; sessions are opaque, hashed at rest, and carried in a path-scoped cookie.
- An optional **Ed25519 key login** replaces passwords for sites that want it.
- **Stats** are counted by the node that served the request: aggregate, cookieless, with a daily-salted visitor count, visible only to the key holder and exportable with the site.

### Compute

- Backend functions run in a **QuickJS** sandbox (via Wazero), deployed beside the site.
- Per-site limits: max memory, max CPU-milliseconds, max calls per minute, and an allow-list of outbound hosts.
- Functions bind to the same per-site database. Which node executes one is an implementation detail.

### Governance

- Each server **publishes its rule set**; small local classifiers score a candidate deploy against it.
- In-rule deploys go live automatically; borderline ones are **held for a human**, never silently erased.
- Rules are files: readable, forkable, and debatable.

### Using Relay

1. **Build** — by hand, with a framework, or by describing what you want. Any stack will do.
2. **Preview** — see the static site locally, exactly as it will be served.
3. **Deploy** — one step: choose a server, get an address and your key.
4. **Store** — connect the site to its database so it can remember what users give it.
5. **Know your people** — turn on site-level login and read your own stats; both stay in the site's own data.
6. **Serve from anywhere** — run it on your own machine and let a public friend relay the world in.
7. **Move** — export your data, point at another server, and take your site with you.

### Technical Foundation

- **Static Hosting** — any folder of files becomes a live website
- **Path-Based Addressing** — one host per public node; each node a path prefix, each site a path under it
- **Outbound-Only WebSocket Tunnels** — no port forwarding, no static IP required
- **One Node, One Codebase** — public and private nodes run identical software
- **Ownership Keys** — a text-file key controls and protects each deployed address
- **Per-Site SQLite** — portable, exportable, one file per site
- **Site-Level Login** — opt-in accounts and sessions in the site's own database
- **Cookieless Metrics** — first-party, aggregate usage stats the owner can export
- **Tunnel-Only Mode** — point Relay at a server you already run and use it purely as reachability
- **Federation** — thousands of independent servers, one open protocol
- **Classifier-Based Rules** — local models, published rule sets, transparent moderation
- **Backend Compute (planned)** — functions sandboxed in QuickJS and rate-limited per site

### Deployment

Relay is one binary, so it deploys anywhere a container or a process runs. **Public nodes** need a stable address and ports 80/443; **private nodes** are outbound-only and need no inbound ports at all.

All node state — keys, site files, per-site SQLite — lives under one data directory. Mount it on a volume; without it, the node forgets everything.

#### As a binary

```sh
relay node --data /var/lib/relay --listen :8080
```

Put Caddy (or any reverse proxy) in front for :80/:443, or let the embedded Caddy handle TLS directly.

#### Docker

A draft Dockerfile:

```dockerfile
FROM golang:1.23-alpine AS build
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /relay ./cmd/relay

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /relay /relay
VOLUME /data
EXPOSE 8080
ENTRYPOINT ["/relay", "node", "--data", "/data", "--listen", ":8080"]
```

```sh
docker build -t relay ./deploy
docker run -d --name relay --restart unless-stopped \
  -p 80:8080 -p 443:8443 \
  -v relay-data:/data \
  relay
```

#### Docker Compose

A Compose file for a single node, with its data on a named volume:

```yaml
services:
  relay:
    build: .
    # image: ghcr.io/relay/relay:latest
    command: ["node", "--data", "/data", "--listen", ":8080"]
    ports:
      - "80:8080"
      - "443:8443"
    volumes:
      - relay-data:/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "/relay", "health"]
      interval: 30s

volumes:
  relay-data:
```

#### Docker Swarm

A Swarm stack — a public node on a manager, plus private nodes that dial out to it:

```yaml
services:
  relay-public:
    image: ghcr.io/relay/relay:latest
    command: ["node", "--public", "--data", "/data", "--listen", ":8080"]
    ports:
      - target: 8080
        published: 80
        protocol: tcp
      - target: 8443
        published: 443
        protocol: tcp
    volumes:
      - relay-data:/data
    deploy:
      replicas: 1
      placement:
        constraints: ["node.role == manager"]

  relay-private:
    image: ghcr.io/relay/relay:latest
    command: ["node", "--friend", "relay-public:8080", "--data", "/data"]
    volumes:
      - relay-data:/data
    deploy:
      replicas: 2

volumes:
  relay-data:
```

The public node needs a stable address; the private nodes only dial out, so they can run on any worker.

#### Kubernetes

A Kubernetes StatefulSet (one node, one data volume) and a Service:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: relay
spec:
  serviceName: relay
  replicas: 1
  selector:
    matchLabels: { app: relay }
  template:
    metadata:
      labels: { app: relay }
    spec:
      containers:
        - name: relay
          image: ghcr.io/relay/relay:latest
          args: ["node", "--data", "/data", "--listen", ":8080"]
          ports:
            - containerPort: 8080
          volumeMounts:
            - name: data
              mountPath: /data
  volumeClaimTemplates:
    - metadata:
        name: data
      spec:
        accessModes: ["ReadWriteOnce"]
        resources:
          requests:
            storage: 10Gi
---
apiVersion: v1
kind: Service
metadata:
  name: relay
spec:
  selector: { app: relay }
  ports:
    - port: 80
      targetPort: 8080
```

An Ingress (or a LoadBalancer Service) terminates TLS at the cluster edge and forwards to the Service. Because Relay uses path-based addressing, the Ingress needs no wildcard host and no per-site certificate — one host and one certificate is enough for every site on the node.

**A node is stateful; the edge is not.** Keep the data on a PersistentVolumeClaim, and keep TLS at the ingress or in the embedded Caddy. Scale private nodes freely; give a public node one stable address.

### Further reading

- [`REVIEW.md`](./REVIEW.md) — the consolidated design review: decisions, open questions, and what is not finished yet
- [`docs/vision.md`](docs/vision.md) — the vision behind Relay
