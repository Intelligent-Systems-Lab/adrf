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
