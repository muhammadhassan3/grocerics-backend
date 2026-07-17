# Linker: stop double-paying QC credits, and log raw responses for ML

**Date:** 2026-07-17
**Branch:** `feature/linker-raw-response-cache`
**Status:** Approved, implementing inline (no separate plan)

Two independent changes that happen to touch the same path.

| | Change 1 — search cache | Change 2 — raw response sink |
|---|---|---|
| Purpose | Halve credit spend on linking | Give the ML person data |
| Read by the app? | Yes (`ConfirmLink`) | **Never** |
| Storage | In-memory, lost on restart | Postgres, kept forever |
| If it breaks | Falls back to `GetItem` | QC call proceeds regardless |

They ship together only because both sit on the QuickCommerce path. Neither depends on the other.

---

## Change 1 — Seed prices from the search response

### Problem

Linking a variant to a platform costs two QC credits for one piece of information.

1. `GET /v1/inventory-management/link/search` runs a groupsearch — 1 credit per platform. The
   response carries, per match: name, brand, quantity, multipack, price, MRP, availability,
   inventory, stock label, images, deep link.
2. `POST /v1/inventory-management/variants/:variant_id/links` (`LinkingService.ConfirmLink`)
   throws that away and calls `GetItem` for the chosen item — another credit per platform — purely
   to seed `platform_prices`.

The second call cannot return anything the first did not. Per
`internal/integration/quickcommerce/types.go`, `ItemDetail` is a **strict subset** of `Product`:

| | `Product` (groupsearch) | `ItemDetail` (GetItem) |
|---|---|---|
| ID, Name, Quantity, Available, PricePaise, MRPPaise, Inventory, StockLabel, DeepLink | yes | yes |
| Brand, Multipack, Rating, Images | yes | **no** |

`GetItem` returns fewer fields, later. "groupsearch = discovery, GetItem = truth" is therefore
about **recency only** — there is no field-level justification for it.

Cost per variant linked across 6 platforms: 6 (search) + 6 (confirm) = **12 credits**. The trial
key is credit-limited, so this halves usable throughput for nothing.

### Decision

`SearchCandidates` caches every `Product` it already fetched, keyed by `qc_item_id`. `ConfirmLink`
seeds from that cache; on a miss it calls `GetItem` exactly as today.

Cost after: 6 + 0 = **6 credits**.

```
/link/search → groupsearch (6 credits, unchanged)
             → cache[qc_item_id] = Product   (TTL 30m)

ConfirmLink(qc_item_id)
   hit  → productToPrice(...)   0 credits
   miss → GetItem               1 credit  (today's path)
```

`itemToPrice(..., *ItemDetail)` gains a sibling `productToPrice(..., Product)`. Both emit
`*domain.PlatformPrice` with `Source: domain.PriceSourceAPI`. The deep-link fallback reads from
whichever source served the seed.

**Why in-memory rather than a table:** the cache is a latency/credit optimisation with a correct
fallback, not a source of truth. Losing it on restart costs one `GetItem`. It does not deserve a
migration. `// ponytail: in-memory map, single container. redis if it ever scales out.`

**Why server-side rather than the client posting the candidate back:** `platform_prices.source`
distinguishes `api` from `manual`. If the client supplied the numbers and we recorded
`source: api`, the provenance would be a lie — we would be attesting we fetched values we did not.
`manual-price` already exists for admin-supplied prices and labels them honestly.

### API contract

**Unchanged.** Same route, same `ConfirmLinkRequest`, same response. The frontend needs no changes;
it simply stops burning credits. Pure internal change.

### Error handling

A cache miss is not an error — it is the current code path. Restart, expired TTL, stale tab, or
confirming without searching all fall through to `GetItem`. The cache can only make linking
cheaper, never more broken. A failing `GetItem` still returns `QC_GET_ITEM_FAILED`.

### Staleness (accepted)

A hit writes `source: api` with `last_updated_at = now`, so a 29-minute-old price records as fresh.
The TTL is the only bound. Accepted: `POST /refresh` re-pulls live prices and corrects drift, and
admins search and link within minutes. If 30 minutes proves long, it is one constant.

---

## Change 2 — `qc_raw_responses`, a write-only ML sink

### Purpose

The ML person wants raw QuickCommerce data. This table exists **solely** to give it to them. No
application code reads it — not the linker, not the cache, not the consumer API. It is a sink.

This also lands the `QC_DEBUG` raw-response logging noted as started-and-abandoned in the handoff.

### Capture point

`httpClient.doGet` (`internal/integration/quickcommerce/http.go:38`). Every QC call — groupsearch,
item, search, eta, groupeta, credits, supported-platforms — funnels through it, so one writer
covers all callers with no per-callsite work. Today it decodes straight off `resp.Body`; it will
read the body to `[]byte` first, record, then unmarshal from those bytes.

### Wiring

The integration package must not import `repository` (layering). So `Config` gains a callback —
a plain func, not an interface with one implementation:

```go
type RawCall struct {
    Endpoint   string        // "/groupsearch"
    Params     url.Values    // query params sent
    StatusCode int           // 0 when the request never completed
    Body       []byte        // complete unmodified response body
    Err        string        // transport/decode error, empty on success
    DurationMs int
}

type Config struct {
    APIKey  string
    BaseURL string
    Record  func(RawCall)   // nil = don't record
}
```

`app.go` wires `Record` to a repository-backed writer. Nil-safe, so tests and the mock client are
unaffected.

### Table (migration `00014`)

```sql
CREATE TABLE qc_raw_responses (
    id           uuid PRIMARY KEY,
    endpoint     text        NOT NULL,
    params       jsonb       NOT NULL DEFAULT '{}'::jsonb,
    status_code  int         NOT NULL,
    response     jsonb,              -- null when the body wasn't valid JSON
    response_text text,              -- raw text fallback when response is null
    error        text,
    duration_ms  int         NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_qc_raw_responses_endpoint_created ON qc_raw_responses (endpoint, created_at DESC);
```

First `jsonb` in the schema — no precedent in `internal/domain`, so the model uses
`datatypes.JSON` (gorm.io/datatypes, already an indirect dep via gorm).

Failures are recorded too (non-2xx, transport errors): the ML person wants the real distribution,
and a table of only-successes misrepresents it.

### Non-negotiable: never break a QC call

The sink is fire-and-forget. If the insert fails, log at warn and carry on — an admin's search must
never fail because an ML table is full or locked. Insert is synchronous (a few ms, admin-only path)
rather than a goroutine, to avoid unbounded goroutine growth on a QC burst.
`// ponytail: sync insert, errors swallowed. queue it if it ever shows up in latency.`

### Retention: forever

Explicitly chosen — the ML person wants full history. **Accepted risk:** the table grows unbounded
with admin search volume, and the handoff records the box at ~87% disk. Response bodies are a few
KB, so this is slow, not sudden. Revisit with a retention window or periodic dump-to-S3 if disk
gets tight. `// ponytail: unbounded by design. prune or archive if disk bites.`

---

## Testing

Both in `test/` (package `test`), using the existing `quickcommerce/mock.go`.

**Cache:**
1. **Hit** — search, then confirm. Assert `GetItem` was never called and the persisted
   `platform_prices` row matches the search response's price/availability.
2. **Miss** — confirm an id that was never searched. Assert `GetItem` *was* called and the price
   still lands.

Together they fail if the cache stops being consulted, or is consulted when it should not be.

**Raw sink:**
3. A recorded call produces one row with the endpoint and body intact.
4. A `Record` callback that panics/errors does not fail the QC call — the fire-and-forget guarantee.

## Non-goals

**Creating variants from the search response — rejected.** An earlier draft let the admin create a
missing pack size (spotting 2 L when only 300 ml and 1 L exist) straight from the candidate row,
with `volume_value`/`volume_unit` prefilled by a new quantity-string parser. Rejected as too many
edges for the value:

- No unique constraint on `(product_id, volume_value, volume_unit)` — two admins on the same row
  create twin variants.
- `"2 x 100 g"` is ambiguous (200 g, or a 2-pack of 100 g?) and the variant model has no multipack
  field to express the difference.
- The parser can be confidently wrong (`"1.25 L"` → 1) in a way only an attentive admin catches.

The flow stays: create the variant manually via `POST /inventory-management/variants`, then link it.
This also removes any need for the quantity parser — its only consumer was the prefill.
`Candidate.Quantity` already carries the raw string for the admin to read.

**Also out:** combined create-and-link endpoint; batch multi-platform link (linking is free now, so
the frontend looping is fine); changing the refresh path; the ML person's query patterns or schema
preferences beyond "raw, all of it, forever".
