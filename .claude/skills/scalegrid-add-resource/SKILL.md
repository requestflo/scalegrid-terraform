---
name: scalegrid-add-resource
description: Step-by-step for adding a new resource or data source to terraform-provider-scalegrid (client method, schema, registration, docs, tests). Use when implementing a new scalegrid_* resource/data source or exposing a new API capability in Terraform.
---

# Adding a resource or data source

Mirror the existing files — don't invent new patterns. Good templates:
`internal/provider/backup_resource.go`, `backup_schedule_resource.go`, the four
`*_cluster_resource.go`, and `database_versions_data_source.go`.

## 1. Client method (`internal/client/`)

Add the API call to the relevant file (or a new one). Keep it Terraform-free.

- Return `(…, actionID, err)` for async mutations and let the resource call
  `WaitForAction`. Read endpoints return the decoded struct.
- **Decode leniently.** Any id/flag/threshold/timestamp on a response struct must tolerate
  number-or-string (see the `scalegrid-api` skill: `rawID`, `rawBool`, `flexID`, and the
  embedded-alias `UnmarshalJSON` pattern + the matching-tag gotcha). Add a `httptest` unit
  test that feeds the numeric form.
- Mind per-engine endpoint/path casing (`PathPrefix` vs `listPrefix`, version casing, Redis
  lowercase variants).

## 2. Resource/data source (`internal/provider/`)

Implement `Metadata`, `Schema`, `Configure`, and CRUD (`Create/Read/Update/Delete`) or `Read`.

- `Metadata`: `resp.TypeName = req.ProviderTypeName + "_<name>"`.
- Reuse helpers: `clientFromProviderData`, `parseDBTypeDiag`, `stringValue`, `optionalString`,
  `stringsFromList`, plan modifiers `reqReplaceStr()/reqReplaceInt()`, `boolRequiresReplace()`,
  `listRequiresReplace()`, `UseStateForUnknown` for computed `id`.
- Validators: `stringvalidator.OneOf(...)`, `int64validator.OneOf/Between/AtLeast`.
- Persist the id early for async creates (`persistIDEarly`) before waiting.
- If the API has **no read endpoint** for the thing, make `Read` a no-op that preserves state
  and document the resource as write/assert (no drift detection / no import) — see
  `backup_schedule_resource.go`.

## 3. Register it

Add the constructor to `internal/provider/provider.go` — `Resources()` or `DataSources()`.

## 4. Update the count assertion ⚠️

`internal/provider/provider_test.go` asserts the exact number of resources
(`TestResourceSchemasValid`) / data sources. Bump it, or the suite fails.

## 5. Example + docs

- Add `examples/resources/scalegrid_<name>/resource.tf` (or `data-sources/...`). The example is
  rendered into the generated doc's "Example Usage".
- Regenerate docs: `CHECKPOINT_DISABLE=1 tfplugindocs generate --provider-name scalegrid`
  (see the `scalegrid-release` skill).

## 6. Verify before committing

```bash
gofmt -l internal/ && go build ./... && go vet ./... && go test ./...
```

Then branch off latest `main`, commit with a Conventional Commit (`feat:` for a new
resource), push, and open a PR. New resources bump the **minor** version on release.
