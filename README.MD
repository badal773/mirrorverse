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
##  Key Labels (Used by Mirrorverse)

### **Source Resource Labels**

These are applied to the **source Secret/ConfigMap** to enable and control syncing.

| Label                                     | Purpose                                                         | Default            |
| ----------------------------------------- | --------------------------------------------------------------- | ------------------ |
| `mirrorverse.dev/sync-source: "true"`     | Opt-in to syncing. Marks this resource as the source of truth.  | Required           |
| `mirrorverse.dev/targets: "staging_prod"` | Underscore-separated list of target namespaces to sync into.         | Required           |
| `mirrorverse.dev/strategy: "replace"`     | Sync strategy: `replace` fully overwrites, `patch` merges.      | `patch`          |
| `mirrorverse.dev/cleanup: "true"`         | If source is deleted, cleanup replicas automatically.           | `false` (optional) |
| `mirrorverse.dev/exclude: "devops"`       | Underscore-separated list of namespaces to **exclude** from targets. | Optional           |

---

###  **Replica Resource Labels**

These are automatically added by Mirrorverse to track and manage synced replicas.

| Label                                       | Purpose                                                               |
| ------------------------------------------- | --------------------------------------------------------------------- |
| `mirrorverse.dev/sync-replica: "true"`      | Indicates that this resource is a managed replica.                    |
| `mirrorverse.dev/sync-source-ref: "<name>.<namespace>"` | References the name of the source resource it was synced from.        |
| `mirrorverse.dev/stale: "true"`             | Set when the source no longer exists — marks the replica as orphaned. |

---

### Notes

* `mirrorverse.dev/cleanup: "true"` enables **auto-deletion** of replicas when their source is deleted. If omitted, replicas will just be marked as `stale`.
* Sync targets (`mirrorverse.dev/targets`) and excludes (`mirrorverse.dev/exclude`) can both be specified, and exclude takes precedence.
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

## Installation (Helm)

```sh
# 1. Add the Mirrorverse Helm repository
helm repo add mirrorverse https://badal773.github.io/mirrorverse/

# 2. Install the Mirrorverse chart
helm install mirrorverse mirrorverse/mirrorverse  --create-namespace --namespace mirrorverse
```

---

## Suggestions & Best Practices

- Use clear and unique names for your source resources to avoid confusion in target namespaces.
- Regularly review and clean up stale replicas (those labeled with `mirrorverse.dev/stale: "true"`).
- Use the `mirrorverse.dev/strategy` label to control how updates are propagated (choose `override` for full replacement or `patch` for selective updates).
- Always test your sync configuration in a staging environment before rolling out to production.
