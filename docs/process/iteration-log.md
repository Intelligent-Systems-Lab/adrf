# ADRF Iteration Log

This file records ADRF implementation rounds.
Follow `implementation-protocol.md` and apply rolling consolidation.

---

## Round R00 - Initial Setup

### Pre-Check (before coding)
- Goal: Establish mandatory implementation workflow files.
- Scope (in): process docs only.
- Scope (out): ADRF code changes.
- Required docs re-read:
  - docs/contract/adrf1.md (overall scope)
  - docs/impl/adrf-free5gc-alignment-guide.md (constraints)
- Constraints confirmed:
  - free5gc-style logger
  - English comments
  - validation command set

### Implementation
- Files changed:
  - docs/process/implementation-protocol.md
  - docs/process/iteration-log.md
- Key design decisions:
  - Use one protocol doc + one cumulative log.
- Tradeoffs:
  - Slight process overhead for better traceability.

### Validation
- Commands:
  - Not applicable (docs-only round).
- Result summary:
  - Process files created.
- Failures/Warnings (if any):
  - None.

### Risks / Open Items
- Team discipline is required to keep this format consistent.

### Next Round Plan
- Start ADRF V0 code skeleton (R01).

---

## Round R01 (Consolidated) - 2026-03-26

Covers: `R01`, `R01-Followup`, `R01-Followup2`.

### Pre-Check (before coding)
- Goal: Build ADRF skeleton and align baseline style with free5gc.
- Scope (in): project scaffold, lifecycle wiring, logger/config, SBI route skeleton, lint/style convergence.
- Scope (out): store/retrieval business logic.
- Required docs re-read:
  - docs/contract/adrf1.md
  - docs/impl/adrf-free5gc-alignment-guide.md
  - docs/spec/TS29575_Nadrf_DataManagement.yaml
- Constraints confirmed:
  - free5gc-style structure/logger
  - English maintainer-facing comments
  - validation command set

### Implementation
- Aggregated file changes:
  - Initialized ADRF module skeleton (`cmd`, `config`, `pkg`, `internal`).
  - Added app lifecycle, config parsing/defaults, SBI server skeleton, processor/consumer placeholders.
  - Added free5gc-style logger categories and route constants.
  - Aligned `.golangci.yml` with free5gc style.
  - Converged naming/types:
    - `models.ProblemDetails`
    - `Mongodb.Url`
    - `AdrfDataRetrievalSubscriptionsPath`
- Key design decisions:
  - Keep full package boundaries from day one.
  - Return `501` for not-yet-supported APIs.
- Tradeoffs:
  - Upgraded module/toolchain baseline (`go 1.25.5`) due dependency requirements.

### Validation
- Commands:
  - go mod tidy
  - go fmt ./...
  - docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...
  - go vet ./...
  - go build ./...
  - go test ./...
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`).
- Failures/Warnings (if any):
  - Early lint/config/dependency issues were fixed in-round.

### Risks / Open Items
- Functional ADRF data path not implemented yet in this consolidated stage.

### Next Round Plan
- Implement R02 store model, Mongo indexes, and StorageRequest flow.

---

## Round R02 (Consolidated) - 2026-03-26

Covers: `R02`, `R02-Phase2`, `R02-Code-Followup`, `R02-Docs-Followup`.

### Pre-Check (before coding)
- Goal: Deliver store-path baseline and improve maintainability/observability standards.
- Scope (in):
  - persistent model + Mongo repository/indexes
  - `POST /data-store-records` end-to-end path
  - comment/log quality convergence
  - process rule updates for documentation discipline
- Scope (out): retrieval subscribe/request/unsubscribe implementation.
- Required docs re-read:
  - docs/contract/adrf1.md
  - docs/contract/adrf2.md
  - docs/impl/adrf-free5gc-alignment-guide.md
  - docs/spec/TS29575_Nadrf_DataManagement.yaml
  - docs/process/implementation-protocol.md
- Constraints confirmed:
  - free5gc-style logging and API behavior
  - maintainer-facing English comments
  - validation command set

### Implementation
- Aggregated file changes:
  - Added persistent store model and repository:
    - `internal/store/datastore_record.go`
    - `internal/store/datastore_repository.go`
  - Added `storeTransId` generation and required Mongo indexes:
    - unique `storeTransId`
    - compound `(supi, ingestedAt)`
    - `ingestedAt`
  - Wired Mongo bootstrap and repository injection in service startup.
  - Implemented StorageRequest path:
    - parse/validate `dataSub + dataNotif`
    - extract `dataSub[*].smfDataSub.supi`
    - persist to Mongo
    - respond `201 + Location + body`
  - Refined comments/logging to remove internal-round wording and improve key action traceability.
  - Updated process docs with explicit comment/log hard constraints and rolling-consolidation policy.
- Key design decisions:
  - Keep `storeTransId` as external retrieval key independent from Mongo `_id`.
  - Use `ingestedAt` as ADRF accept-time metadata for deterministic snapshot logic.
- Tradeoffs:
  - Current store validation focuses on agreed `smfDataSub`-based data path.
  - Broader TS 29.575 optional variants are deferred.

### Validation
- Commands:
  - go mod tidy
  - go fmt ./...
  - docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...
  - go vet ./...
  - go build ./...
  - go test ./...
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`).
- Failures/Warnings (if any):
  - Minor lint warnings (`shadow`, nil-map check, line length) fixed in-round.

### Risks / Open Items
- Retrieval subscribe/request/unsubscribe path remains unimplemented.
- StorageRequest currently targets agreed NWDAF data-shape path; other optional schema branches pending.

### Next Round Plan
- Implement retrieval subscription persistence model and fetch-queue metadata.

---

## Round R03 - 2026-03-26 15:54 (UTC+8)

### Pre-Check (before coding)
- Goal: Complete `POST /data-store-records` success and strict error behavior.
- Scope (in):
  - `201 Created + Location + body`
  - strict `ProblemDetails` mapping for request/persistence failures
  - handler unit tests for key success/failure paths
- Scope (out):
  - retrieval APIs (`POST/DELETE /data-retrieval-subscriptions`, `GET /data-store-records`)
  - OAuth2 behavior
- Required docs re-read:
  - docs/contract/adrf1.md (store and error behavior sections)
  - docs/impl/adrf-free5gc-alignment-guide.md (response + ProblemDetails style)
  - docs/spec/TS29575_Nadrf_DataManagement.yaml (`/data-store-records` response and error codes)
- Constraints confirmed:
  - free5gc-style logger usage
  - English maintainer-facing comments
  - validation command set

### Implementation
- Files changed:
  - `internal/sbi/processor/datastore_request.go`
    - added strict header checks (`Content-Type`, `Content-Length`/chunked)
    - added structured `ProblemDetails` builders with `invalidParams`
    - refined JSON bind error mapping (`INVALID_JSON`, mandatory body handling)
    - refined persistence error mapping (`500/503` system-failure paths)
    - generated absolute `Location` URI when host/scheme is available
  - `internal/sbi/processor/processor.go`
    - replaced concrete repository dependency with narrow writer interface for better testability
  - `internal/sbi/processor/datastore_request_test.go` (new)
    - success test for `201 + Location + body`
    - failure tests for `415/411/400/503/500` ProblemDetails mapping
- Key design decisions:
  - enforce request preconditions before JSON parsing for deterministic HTTP status behavior.
  - keep response payload as stored request shape (`dataSub + dataNotif`) while generating server-owned `storeTransId` via `Location`.
- Tradeoffs:
  - V0 intentionally validates and extracts SUPI from `dataSub[*].smfDataSub.supi` only (agreed data path).

### Validation
- Commands:
  - `go fmt ./...`
  - `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`; `go test` passed including new processor tests).
- Failures/Warnings (if any):
  - First lint attempt was blocked by Docker socket sandbox permission; rerun with approved escalation succeeded.

### Risks / Open Items
- `anaSub + anaNotifications` branch of `NadrfDataStoreRecord` remains out of V0 scope.
- `ProblemDetails` cause strings are implementation-defined; keep consistent across future handlers.

### Next Round Plan
- Implement retrieval subscription persistence and callback workflow (R04).

---

## Round R04 - 2026-03-26 16:16 (UTC+8)

### Pre-Check (before coding)
- Goal: Implement `POST /data-retrieval-subscriptions` with snapshot freeze and fetch correlation ID generation.
- Scope (in):
  - create retrieval subscription handler (`201 + Location + body`)
  - apply snapshot freeze rule (`supi + timePeriod + ingestedAt <= T_sub`)
  - generate and store `fetchCorrIds` for follow-up callback rounds
  - add unit tests for handler and time-window filter logic
- Scope (out):
  - callback delivery (`RetrievalNotify`) execution
  - fetch GET API and unsubscribe API
- Required docs re-read:
  - docs/contract/adrf1.md (snapshot semantics and retrieval request shape)
  - docs/contract/adrf2.md (flow decomposition and expected behavior)
  - docs/impl/adrf-free5gc-alignment-guide.md (response/error style and logging expectations)
  - docs/spec/TS29575_Nadrf_DataManagement.yaml (`/data-retrieval-subscriptions` and schema requirements)
- Constraints confirmed:
  - free5gc-style logger usage
  - English maintainer-facing comments
  - validation command set

### Implementation
- Files changed:
  - `internal/sbi/processor/processor.go`
    - extended repository interface with snapshot query contract
    - added in-process retrieval subscription state map
    - wired `HandleCreateDataRetrievalSubscription` to concrete implementation
  - `internal/sbi/processor/retrieval_subscription_request.go` (new)
    - implemented create-retrieval-subscription request parsing/validation
    - implemented strict `ProblemDetails` errors for request and snapshot failures
    - implemented snapshot cutoff capture (`T_sub`) and repository snapshot query call
    - generated `subscriptionId`, set `Location`, and returned `201 + body`
  - `internal/sbi/processor/retrieval_subscription_state.go` (new)
    - added runtime state structure for `subscriptionId -> snapshot/fetchCorrIds`
  - `internal/store/datastore_snapshot_query.go` (new)
    - implemented Mongo snapshot candidate query by `supi` + `ingestedAt <= T_sub`
    - added `notificationItems[*].startTime` window filter over `dataNotif.upfEventNotifs`
    - returned deterministic ordered `storeTransId` list as `fetchCorrIds`
  - `internal/sbi/processor/retrieval_subscription_request_test.go` (new)
    - added success and error tests for create-retrieval-subscription handler
  - `internal/store/datastore_snapshot_query_test.go` (new)
    - added unit tests for startTime window matching across timestamp encodings
  - `internal/sbi/processor/datastore_request_test.go`
    - updated stub repository to satisfy extended processor interface
- Key design decisions:
  - keep retrieval snapshot filtering deterministic by freezing `T_sub` once at subscription creation.
  - keep fetch correlation IDs equal to `storeTransId` (no transformation/prefixing).
  - persist retrieval runtime state in-process for immediate use in R05 callback delivery.
- Tradeoffs:
  - snapshot time-window filtering is currently performed in ADRF application logic after indexed Mongo pre-filtering.
  - V0 retrieval path is intentionally restricted to `dataSub.smfDataSub.supi` + `consTrigNotif=true`.

### Validation
- Commands:
  - `go fmt ./...`
  - `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`; processor/store tests green).
- Failures/Warnings (if any):
  - Initial lint run reported `errcheck/shadow/lll` issues; fixed in-round and revalidated.

### Risks / Open Items
- Retrieval subscription state is currently process-local and not yet persisted across ADRF restarts.
- Callback batching and termination notification are pending R05.

### Next Round Plan
- Implement RetrievalNotify callback sender with `corrIdBatchSize` chunking and `terminationReq=true` on the last batch (R05).

---

## Round R05 - 2026-03-26 16:34 (UTC+8)

### Pre-Check (before coding)
- Goal: Implement Retrieval notify callback delivery with corr-id batching and termination signaling.
- Scope (in):
  - callback sender abstraction and HTTP implementation
  - `corrIdBatchSize` batching logic for `fetchInstruct.fetchCorrIds`
  - `terminationReq=true` on the last callback batch
  - async dispatch trigger after successful retrieval-subscription creation
  - unit tests for batching and callback dispatch behavior
- Scope (out):
  - `GET /data-store-records` fetch response implementation
  - retrieval unsubscribe implementation
  - callback retry scheduler/persistence
- Required docs re-read:
  - docs/contract/adrf1.md (retrieval callback batching and termination behavior)
  - docs/contract/adrf2.md (R05 flow ownership and notify sequence)
  - docs/impl/adrf-free5gc-alignment-guide.md (free5gc logger/comment style)
  - docs/spec/TS29575_Nadrf_DataManagement.yaml (`NadrfRetrievalNotify`, `FetchInstruction`)
- Constraints confirmed:
  - free5gc-style logger usage
  - English maintainer-facing comments
  - validation command set

### Implementation
- Files changed:
  - `internal/sbi/processor/processor.go`
    - added `retrievalNotifier` dependency field and default HTTP sender wiring.
  - `internal/sbi/processor/retrieval_notification_sender.go` (new)
    - added callback sender interface + HTTP implementation.
    - implemented corr-id chunking helper with last-batch `terminationReq=true`.
    - implemented callback payload build (`notifCorrId`, `fetchInstruct`, optional `terminationReq`, `timeStamp`).
    - implemented callback transport/error handling and operational logs.
    - added fetch URI builder for `/data-store-records`.
  - `internal/sbi/processor/retrieval_subscription_request.go`
    - after `201 + Location + body`, triggers async callback dispatch with frozen snapshot IDs.
  - `internal/sbi/processor/retrieval_subscription_request_test.go`
    - added notifier stub injection and success assertion that callback dispatch is triggered.
    - added check that `corrIdBatchSize` is read from ADRF config.
  - `internal/sbi/processor/retrieval_notification_sender_test.go` (new)
    - added batching tests and callback payload assertions for termination behavior.
- Key design decisions:
  - keep callback delivery decoupled from create-subscription response latency via asynchronous dispatch.
  - keep callback batch-size control at ADRF config (`retrieval.corrIdBatchSize`), independent from NWDAF config.
  - when no IDs exist, still emit one terminating callback to close workflow deterministically.
- Tradeoffs:
  - current callback dispatch has no retry queue/persistence; failures are logged and left for later rounds.

### Validation
- Commands:
  - `go fmt ./...`
  - `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`; processor tests green with new R05 tests).
- Failures/Warnings (if any):
  - first docker-lint attempt failed due sandbox docker-socket permission; rerun with approved escalation succeeded.

### Risks / Open Items
- Callback dispatch remains best-effort in-process; ADRF restart loses in-flight callback work.
- Retry/backoff and dead-letter handling are not implemented yet.

### Next Round Plan
- Implement `GET /data-store-records` fetch path (single `fetch-correlation-ids` first, 200/204 behavior) for R06.

---

## Round R06 - 2026-03-26 16:51 (UTC+8)

### Pre-Check (before coding)
- Goal: Implement retrieval fetch API (`GET /data-store-records`) for single-ID path.
- Scope (in):
  - support `fetch-correlation-ids` query for one correlation ID per request
  - map existing store record to `NadrfDataStoreRecord` response (`200`)
  - return `204` when requested record is not found
  - add strict query validation and problem mapping for invalid requests
  - add unit tests for `200/204/400/503/500`
- Scope (out):
  - multi-ID fetch response aggregation
  - `store-trans-id` and `data-set-id` retrieval branches
  - delete APIs and unsubscribe APIs
- Required docs re-read:
  - docs/contract/adrf1.md (fetch one-ID behavior and 204 handling)
  - docs/contract/adrf2.md (R06 fetch flow)
  - docs/spec/TS29575_Nadrf_DataManagement.yaml (`GET /data-store-records`, query parameters, 200/204 responses)
  - docs/impl/adrf-free5gc-alignment-guide.md (style and error/logging expectations)
- Constraints confirmed:
  - free5gc-style logger usage
  - English maintainer-facing comments
  - validation command set

### Implementation
- Files changed:
  - `internal/sbi/processor/processor.go`
    - wired `HandleGetDataStoreRecords` to concrete fetch handler.
    - extended store interface with fetch-by-storeTransID method.
  - `internal/sbi/processor/datastore_retrieval_request.go` (new)
    - implemented `GET /data-store-records` handler.
    - added query validation for ADRF V0 fetch mode:
      - exactly one retrieval key
      - fetch-correlation-ids only
      - exactly one fetch correlation ID
    - mapped `mongo.ErrNoDocuments` to `204 No Content`.
    - mapped timeout/cancel/internal failures to `ProblemDetails`.
    - converted persisted BSON payload back to response JSON payload.
  - `internal/store/datastore_fetch_query.go` (new)
    - added Mongo lookup by `storeTransId` for retrieval fetch path.
  - `internal/sbi/processor/datastore_request_test.go`
    - extended repository stub to support fetch method.
  - `internal/sbi/processor/datastore_retrieval_request_test.go` (new)
    - added coverage for `200/204/400/503/500` cases.
- Key design decisions:
  - treat fetch correlation ID as direct `storeTransId` lookup key.
  - keep R06 single-ID behavior strict to match agreed NWDAF one-by-one fetch path.
  - treat `mongo.ErrNoDocuments` as normal no-data signal (204), not error.
- Tradeoffs:
  - `store-trans-id` and `data-set-id` retrieval branches are deferred until later scope.
  - query validation currently returns `MANDATORY_IE_MISSING` cause for unsupported/mixed query shapes in V0.

### Validation
- Commands:
  - `go fmt ./...`
  - `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`; processor/store tests green).
- Failures/Warnings (if any):
  - first docker-lint attempt failed due sandbox docker-socket permission; rerun with approved escalation succeeded.

### Risks / Open Items
- Retrieval fetch currently supports only one `fetch-correlation-id` per request by design.
- `store-trans-id` and `data-set-id` query modes remain unimplemented in ADRF V0.

### Next Round Plan
- Implement `DELETE /data-retrieval-subscriptions/{subscriptionId}` with state cleanup and `204` behavior (R07).

---

## Round R06-Followup - 2026-03-26 17:04 (UTC+8)

### Pre-Check (before coding)
- Goal: Tune ADRF logging for low-noise demo operation while keeping maintainability.
- Scope (in):
  - reduce high-frequency success-path logs from `info` to `debug`
  - keep failures and state transitions visible at `warn/error`/selected `info`
  - document logging policy in process protocol
- Scope (out):
  - API behavior changes
  - retrieval/store schema changes

### Implementation
- Files changed:
  - `internal/sbi/processor/datastore_request.go`
    - downgraded hot-path success logs to debug and masked identifiers.
  - `internal/sbi/processor/datastore_retrieval_request.go`
    - downgraded per-fetch `200/204` logs to debug and masked fetch IDs in logs.
  - `internal/sbi/processor/retrieval_notification_sender.go`
    - downgraded per-batch callback success logs to debug; kept subscription-level lifecycle in info.
    - masked subscription/correlation identifiers in info/error logs.
  - `internal/sbi/processor/retrieval_subscription_request.go`
    - masked identifiers in retrieval-subscription creation info log.
  - `internal/sbi/processor/logging_helpers.go` (new)
    - added identifier summarization helper for operational logs.
  - `internal/store/datastore_repository.go`
    - downgraded per-record insert success log to debug.
  - `docs/process/implementation-protocol.md`
    - added mandatory low-noise logging requirements for demo/testbed operation.
- Key design decisions:
  - never log full retrieval payload content in hot paths.
  - keep info-level logs for low-frequency transitions; move high-frequency success telemetry to debug.
- Tradeoffs:
  - info-level visibility of per-record success events is intentionally reduced; deep tracing now requires debug level.

### Validation
- Commands:
  - `go fmt ./...`
  - `docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...`
  - `go vet ./...`
  - `go build ./...`
  - `go test ./...`
- Result summary:
  - All commands passed (`golangci-lint: 0 issues`).
- Failures/Warnings (if any):
  - first docker-lint attempt failed due sandbox docker-socket permission; rerun with approved escalation succeeded.

### Risks / Open Items
- If production troubleshooting requires per-request traces, operators must temporarily switch logger level to debug.

### Next Round Plan
- Continue R07 unsubscribe path implementation with the same low-noise logging policy.
