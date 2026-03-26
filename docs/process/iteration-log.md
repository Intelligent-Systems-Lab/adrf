# ADRF Iteration Log

This log keeps ADRF implementation history compact and reviewable.
It follows `docs/process/implementation-protocol.md` and uses rolling consolidation.

---

## Round R00 (Consolidated) - 2026-03-26

### Pre-Check (before coding)
- Goal: establish process guardrails before feature work.
- Scope (in): process documents only.
- Scope (out): ADRF runtime code.

### Implementation
- Created process baseline:
  - `docs/process/implementation-protocol.md`
  - `docs/process/iteration-log.md`
- Established round template and Definition of Done.

### Validation
- Docs-only round; no code validation commands required.

### Risks / Open Items
- Process quality depends on disciplined use each round.

### Next Round Plan
- Start ADRF skeleton and free5gc-style baseline.

---

## Round R01-R02 (Consolidated) - 2026-03-26

Covers: `R01`, `R01-Followup*`, `R02`, `R02-Followup*`.

### Pre-Check (before coding)
- Goal: build ADRF skeleton and store-path foundation.
- Scope (in):
  - project scaffold, config/logger/server wiring
  - Mongo repository/model/indexes
  - `POST /data-store-records` baseline
  - style/lint alignment and process constraints
- Scope (out): retrieval runtime flow.

### Implementation
- ADRF skeleton:
  - `cmd`, `pkg`, `internal`, `config`, `.golangci.yml`
  - free5gc-style logger categories and SBI route setup
- Store persistence model and repository:
  - `storeTransId`, `ingestedAt`, `supi`
  - required indexes: unique `storeTransId`, `(supi, ingestedAt)`, `ingestedAt`
- `POST /data-store-records`:
  - validation + parsing
  - Mongo insert
  - `201 + Location + body`
- Process hardening:
  - comment/logging conventions
  - rolling consolidation rule in protocol.

### Validation
- Passed:
  - `go fmt ./...`
  - `docker run ... golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`

### Risks / Open Items
- Retrieval APIs still pending.
- V0 data shape intentionally focused on agreed `smfDataSub`-based path.

### Next Round Plan
- Implement strict store error behavior and retrieval path.

---

## Round R03-R05 (Consolidated) - 2026-03-26

Covers: `R03`, `R04`, `R05`.

### Pre-Check (before coding)
- Goal: complete store strictness and retrieval subscribe+notify flow.
- Scope (in):
  - strict `ProblemDetails` for store path
  - retrieval snapshot subscribe (`POST /data-retrieval-subscriptions`)
  - callback batching (`corrIdBatchSize`) with terminal notify
- Scope (out):
  - fetch retrieval API
  - unsubscribe API
  - persistent retry scheduler

### Implementation
- Store strict behavior (`R03`):
  - deterministic header/body validation
  - strict `ProblemDetails` mapping
  - `201 + Location + body` preserved
- Retrieval subscribe (`R04`):
  - snapshot freeze rule implemented:
    - `supi + timePeriod + ingestedAt <= T_sub`
  - deterministic `fetchCorrIds` generation from `storeTransId`
  - in-memory retrieval subscription state
- Retrieval notify callback (`R05`):
  - callback sender abstraction + HTTP sender
  - batch split by `corrIdBatchSize`
  - final batch `terminationReq=true`
  - async dispatch after successful subscribe

### Validation
- Passed all standard commands (`fmt/lint/vet/build/test`) each round.
- Added unit tests for:
  - store success/error mapping
  - snapshot filtering
  - callback batching and dispatch trigger

### Risks / Open Items
- Callback dispatch remained best-effort in-process (no persistent retry queue).
- Subscription state remained process-local.

### Next Round Plan
- Implement fetch API and logging-noise convergence.

---

## Round R06 (Consolidated) - 2026-03-26

Covers: `R06`, `R06-Followup`.

### Pre-Check (before coding)
- Goal: deliver single-ID fetch path and reduce hot-path log noise.
- Scope (in):
  - `GET /data-store-records?fetch-correlation-ids=<id>`
  - `200` one record, `204` no data
  - low-noise operational logging policy
- Scope (out):
  - multi-ID fetch response
  - `store-trans-id` / `data-set-id` retrieval branches

### Implementation
- Retrieval fetch API:
  - lookup by `storeTransId` via fetch-correlation-id
  - strict query validation + `ProblemDetails`
  - `mongo.ErrNoDocuments -> 204`
- Logging convergence:
  - moved high-frequency success logs to `debug`
  - kept lifecycle and failures visible (`info`/`warn`/`error`)
  - masked IDs in logs, no payload dump on hot paths
  - updated protocol logging requirements.

### Validation
- Passed all standard commands (`fmt/lint/vet/build/test`).

### Risks / Open Items
- Still one fetch ID per request by V0 design.
- Deep trace requires temporary debug log level.

### Next Round Plan
- Implement retrieval unsubscribe with cancellation.

---

## Round R07 (Consolidated) - 2026-03-26

### Pre-Check (before coding)
- Goal: implement unsubscribe cleanup and stop future callbacks.
- Scope (in):
  - `DELETE /data-retrieval-subscriptions/{id}`
  - runtime state cleanup
  - cancel in-flight dispatch
- Scope (out):
  - persistent subscription store
  - cross-restart resume

### Implementation
- `DELETE` handler implemented:
  - returns `204`
  - removes subscription state
  - triggers stored dispatch cancel function
- Added subscription-scoped dispatch context/cancel wiring.
- Kept V0 idempotent cleanup behavior for already-removed IDs.

### Validation
- Passed all standard commands (`fmt/lint/vet/build/test`).
- Added targeted tests:
  - delete success + cancel invoked
  - delete no-op
  - invalid path id

### Risks / Open Items
- Cancellation state remains process-local.

### Next Round Plan
- Add E2E flow test and finalize docs sync.

---

## Round R08 (Consolidated) - 2026-03-26

### Pre-Check (before coding)
- Goal: prove ADRF V0 main flow with fake NWDAF/fake callback.
- Scope (in):
  - `store -> subscribe -> notify -> fetch -> unsubscribe` E2E path
  - process log synchronization
- Scope (out):
  - real NWDAF binary integration
  - restart-resume resilience

### Implementation
- Added E2E test:
  - `internal/sbi/processor/e2e_flow_test.go`
  - uses fake in-memory repository + fake callback transport
  - validates:
    - batched notify by `corrIdBatchSize`
    - terminal `terminationReq=true`
    - per-ID fetch success path
    - unsubscribe cleanup result
- Synced test constants/lint cleanup in existing tests.

### Validation
- Passed all standard commands:
  - `go fmt ./...`
  - `docker run ... golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`

### Risks / Open Items
- E2E is single-process and in-memory by design; persistence and restart behavior remain future work.

### Next Round Plan
- Prepare ADRF V0 handoff checklist (release/readiness notes, remaining gaps, and integration guidance).

