# Mirrorverse

Mirrorverse is a Kubernetes controller that keeps Secrets and ConfigMaps perfectly mirrored across dimensions (namespaces).

Powered by label-driven syncing and conflict-resolution strategies, it ensures your configs are always in sync — even in multiverse chaos.

**Built for multi-tenant clusters. Inspired by Marvel. Written in Go.**

---

## Architecture

```
                                +-------------------------+
                                | Kubernetes API Server   |
                                +-----------+-------------+
                                            |
                                Watch Secrets & ConfigMaps
                                            |
                                +-----------v-------------+
                                |    Syncer Controller    |
                                |                         |
                                | - Detect labeled source |
                                | - Parse labels          |
                                | - Reconcile logic       |
                                +-----------+-------------+
                                            |
                                Create/Update Resources
                                            |
                                +-----------v-------------+
                                |     Target Namespaces    |
                                | (e.g., staging, prod)    |
                                +--------------------------+
```

---

## Key Labels & Annotations

| Resource | Key Label                                   | Purpose                        |
|----------|---------------------------------------------|--------------------------------|
| Source   | `mirrorverse.dev/sync-source: "true"`       | Opt-in to syncing              |
| Source   | `mirrorverse.dev/targets: "staging-prod"`   | Where to sync                  |
| Source   | `mirrorverse.dev/strategy: "override/patch"`| Syncing mode                   |
| Replica  | `mirrorverse.dev/sync-replica: "true"`      | Marked as managed by you       |
| Replica  | `mirrorverse.dev/sync-source-ref: "dev-db"` | Tracks origin                  |
| Replica  | `mirrorverse.dev/stale: "true"`             | Marked as stale if source gone |

---

## How It Works

### 1. Watch Layer (Secrets/ConfigMaps)
- Watches all namespaces for add/update/delete events.

### 2. Sync Logic
- For each source:
  - Parse target list (`mirrorverse.dev/targets`)
  - Skip excluded namespaces (`mirrorverse.dev/exclude`)
  - For each valid target namespace:
    - If not exists → **Create**
    - If exists & strategy is `override` → **Replace**
    - If strategy is `patch` → **Selective Patch**

### 3. Reconciler Loop
- Periodically:
  - Find all `mirrorverse.dev/sync-replica: "true"`
  - Fetch its `sync-source-ref`
  - Compare data, type, etc.
  - If drifted → **Sync again**
  - If source deleted, either:
    - Remove all replicas with that `sync-source-ref`
    - Or label them `mirrorverse.dev/stale: "true"` and skip further syncing

---

## Usage

1. Label your source Secret/ConfigMap with `mirrorverse.dev/sync-source: "true"` and specify targets.
2. Deploy the controller in your cluster.
3. Watch your resources sync across namespaces!

---

## Suggestions & Best Practices

- Use clear and unique names for your source resources to avoid confusion in target namespaces.
- Regularly review and clean up stale replicas (those labeled with `mirrorverse.dev/stale: "true"`).
- Use the `mirrorverse.dev/strategy` label to control how updates are propagated (choose `override` for full replacement or `patch` for selective updates).
- Always test your sync configuration in a staging environment before rolling out to production.
