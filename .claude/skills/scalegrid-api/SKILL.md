---
name: scalegrid-api
description: How the ScaleGrid console API behaves and how this provider's client wraps it. Use when adding or debugging anything in internal/client, decoding API responses, dealing with "cannot unmarshal" errors, async job polling, cloud profiles, sizes/versions, or per-engine endpoint quirks.
---

# ScaleGrid console API & client conventions

The provider talks to the ScaleGrid **console** API at `https://console.scalegrid.io`.
A copy of the OpenAPI spec lives in `scalegrid-openapi.json` at the repo root — but
**it is incomplete**: several endpoints we rely on are not documented there (see
"Reverse-engineered endpoints"). The spec is the starting point, not the source of truth.

## Architecture

- `internal/client` — pure Go API client, **no Terraform dependencies**. All HTTP,
  auth, and JSON decoding lives here.
- `internal/provider` — Terraform Plugin Framework (protocol 6) resources/data sources
  that call the client.

Keep that split. Don't import `terraform-plugin-*` into `internal/client`.

## Auth & response envelope

- Login: `POST /login` with `{username, password}`; sets a session cookie. 2FA returns
  error code `TwoFactorAuthNeeded`. See `client.go`.
- Every response is wrapped: `{"error":{"code":"Success", ...}, ...payload}`. A non-`Success`
  code is an API error; `IsNotFound(err)` recognises the not-found variants.

## Async jobs (THE provisioning model)

Mutating endpoints (create/scale/delete/backup/...) return an **integer** `actionID`
and the cluster/resource id. They are asynchronous — you must poll:

- `WaitForAction(ctx, actionID, pollInterval)` → `GET /actions/{id}`.
- Status is `Running`/`Completed`/`Failed`, and the API returns it **both** at the top
  level **and** sometimes nested under an `"action"` object — `actionResponse` handles both.
- Resources persist the new id to state *before* waiting (`persistIDEarly`) so a timeout
  doesn't orphan the resource.

## ⚠️ JSON decode leniency — the #1 recurring bug

The API is wildly inconsistent about types. The **same field** comes back as a number on
one endpoint and a string on another (and bools as `"true"`/`"false"` strings). Go's
`encoding/json` then fails with `cannot unmarshal number into Go struct field ... of type string`.

**Rule: every identifier/flag on a response struct must be decoded leniently.** Helpers in
`internal/client/types.go`:

- `rawID(json.RawMessage) string` — accepts a number or quoted string → string. Use for all
  ids, clusterIds, machinePoolIds, object_ids, thresholds, and numeric timestamps.
- `rawBool(json.RawMessage) bool` — accepts `true`/`false` or `"true"/"yes"/"1"`.
- `flexID` — a `string` type whose `UnmarshalJSON` calls `rawID`. Use it as a **field type**
  (e.g. `Action.ID flexID`) when you can't add a custom `UnmarshalJSON` to the struct —
  e.g. `Action` is embedded in `actionResponse`, and a custom `Action.UnmarshalJSON` would be
  promoted to the outer struct and break its nested decoding.

Structs already hardened: `Cluster`, `CloudProfile`, `Backup`, `Action`, `AlertRule`.

### The embedded-alias UnmarshalJSON pattern (and its gotcha)

```go
func (b *Backup) UnmarshalJSON(data []byte) error {
    type alias Backup                              // avoid infinite recursion
    aux := struct {
        ID      json.RawMessage `json:"id"`        // shadow numeric fields as RawMessage
        Created json.RawMessage `json:"created"`
        *alias
    }{alias: (*alias)(b)}
    if err := json.Unmarshal(data, &aux); err != nil { return err }
    b.ID = rawID(aux.ID); b.Created = rawID(aux.Created)
    return nil
}
```

**GOTCHA (real bug we shipped and then fixed):** the shadow field's `json:"..."` tag must
match the embedded field's tag **exactly**, or it won't shadow it and the number still hits
the string field. `AlertRule` had `json:"clusterID"` shadowing a field tagged `json:"clusterId"`
— so a numeric `clusterId` still crashed. When the API uses inconsistent casing, capture
**both** spellings (`clusterId` and `clusterID`). Always add a unit test that unmarshals the
field as a number (see `TestLenientDecoding`).

## Reverse-engineered endpoints (NOT in the spec)

These were derived from the live console, not `scalegrid-openapi.json`. Treat their response
shapes as unverified and decode defensively:

- `GET /Clusters/getDatabaseActiveVersions` — returns `versions` as an **object** keyed by
  identifier (`{"V1804":"18.04",...}`), not an array. Decoded by the lenient `versionList` type.
- `setBackupSchedule` / `setClusterBackupSchedule` — no read-back endpoint exists, so the
  backup-schedule resource is write/assert (no drift detection, no import).
- Follower relationship endpoints under `/clusters/{id}/...`.

When wiring a new endpoint, if it's missing from the spec, assume nothing about its response
and tell the user to confirm against the live API (the sandbox can't reach scalegrid.io).

## Per-engine quirks

- `DBType.PathPrefix()` = `"<Engine>Clusters"` (used by create, deletebackup, getBackupSchedule,
  enablePgBouncer). `DBType.listPrefix()` is the same **except PostgreSQL is all-lowercase**
  (`postgresqlclusters`).
- Version casing in create bodies: Mongo & PostgreSQL `ToUpper`, MySQL `ToLower`.
- Endpoint name casing differs per engine: e.g. Redis uses `listbackups` (lowercase) while
  others use `listBackups`; Mongo/Redis use `setClusterBackupSchedule` while MySQL/PostgreSQL
  use `setBackupSchedule`. Check `backups.go` for the existing switches.
- `WireType()` → upper-cased dbType for generic bodies (`MONGODB`, `REDIS`, `MYSQL`, `POSTGRESQL`).

## Cloud profiles (machine pools)

- `/clouds/list` returns each profile with `shared` (bool) and `type` (cloud).
  **`shared: true` = Dedicated / ScaleGrid-hosted; `false` = Bring Your Own Cloud (BYOC).**
- `type` is the wire cloud name: `EC2` (AWS), `AZUREARM` (Azure), `DIGITALOCEAN`, `GCP`, `LINODE`.
  `NormalizeCloudProvider` maps friendly names ↔ these; `ValidCloudProviders` is the accepted set.
- **Every create endpoint requires `machinePoolIDList`** — one machine-pool id per node across
  shards (`shard_count` × nodes-per-shard). Items need not be unique.
- For Dedicated plans the user need not name a profile: `FindSharedCloudProfile(db, provider, region)`
  auto-selects the shared profile and `resolveMachinePools` replicates its id across the nodes.
  Disambiguate with `cloud_provider` + `region` (both optional).

## Sizes & versions

- Sizes are **not enumerated in the spec** and are provider-specific (e.g. `Nano` exists on
  DigitalOcean). `ValidSizes`/`NormalizeSize` is our curated superset — extend it when the API
  accepts a tier we reject, rather than assuming the spec is exhaustive.
- Discover versions with the `scalegrid_database_versions` data source.

## Sandbox limitation

The build/dev sandbox **cannot reach scalegrid.io** (network policy 403). You cannot run
the provider against the real API here — unit-test the client with `httptest`, and have the
user exercise reverse-engineered paths against their live account.
