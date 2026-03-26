# ADRF Implementation Protocol

## 1. Purpose

This protocol defines the mandatory workflow for every ADRF implementation round.
The goal is to keep execution spec-aligned, traceable, and reviewable.

## 2. Mandatory Rule

For every implementation round:

1. Read the required documents before coding.
2. Fill in a `Pre-Check` section before making changes.
3. Fill in a `Round Report` section after changes.
4. Execute validation commands and record outcomes.
5. If a command fails, record the failure and next action.
6. Apply rolling consolidation to keep iteration logs compact:
   - When starting round `Rx`, consolidate `R(x-1)` and its follow-ups into one summary block.
   - The consolidated block must keep only: key decisions, aggregated file changes, validation results, and risks/open items.
   - Keep the current in-progress round detailed; keep older rounds summarized.

A round is not considered complete unless all items above are recorded.

## 3. Required Documents to Re-Read (each round)

1. `docs/contract/adrf1.md`
2. `docs/impl/adrf-free5gc-alignment-guide.md`
3. ADRF Stage-3 spec files used by the round scope (yaml/md), in `docs/spec/`

## 4. Hard Constraints (must always hold)

1. Follow free5gc-style project structure and logger pattern.
2. Keep comments in English only.
3. Add detailed comments for key design logic:
   - state transitions
   - idempotency
   - snapshot filtering
   - error mapping
   - retry/termination behavior
4. Comments must be maintainer-facing:
   - Do reference external standards/specs when relevant (e.g., 3GPP behavior rationale).
   - Do **not** reference internal round labels, internal process docs, or team-only planning context.
5. Logging must be both free5gc-style and operationally informative:
   - Keep category-based logger usage (`Init`, `SBI`, `Proc`, `Store`, etc.).
   - Add enough logs at important state transitions so maintainers can trace key actions and failures.
   - Avoid logs that are too sparse to explain what critical step is running or what failed.
   - For testbed/demo default operation (`logger.level: info`), high-frequency success-path logs must be downgraded to `debug` to avoid log flooding.
   - Keep `info` for lifecycle and low-frequency state transitions (bootstrap, subscription create/complete, shutdown, etc.).
   - Keep `warn/error` for validation failures, dependency failures, callback failures, and unexpected paths.
   - Never dump full data payloads (`dataSub`, `dataNotif`, notification item arrays) in logs on hot paths.
   - When logging correlation IDs / subscription IDs / SUPI in `info` or above, prefer masked/shortened forms.
6. Keep API behavior aligned with agreed ADRF V0 contract.

## 5. Standard Validation Commands

Run these commands in ADRF repo root:

```bash
go fmt ./...
docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...
go vet ./...
go build ./...
go test ./...
```

## 6. Required Round Template

Use this exact structure in `iteration-log.md` for each round.

```md
## Round RXX - YYYY-MM-DD HH:mm (UTC+8)

### Pre-Check (before coding)
- Goal:
- Scope (in):
- Scope (out):
- Required docs re-read:
  - docs/contract/adrf1.md (sections: ...)
  - docs/impl/adrf-free5gc-alignment-guide.md (sections: ...)
  - spec yaml/md (sections: ...)
- Constraints confirmed:
  - free5gc-style logger
  - English comments (detailed on key logic)
  - validation command set

### Implementation
- Files changed:
  - path: what changed
- Key design decisions:
- Tradeoffs:

### Validation
- Commands:
  - go fmt ./...
  - docker run --rm -v "$PWD":/app -w /app golangci/golangci-lint:latest golangci-lint run ./...
  - go vet ./...
  - go build ./...
  - go test ./...
- Result summary:
- Failures/Warnings (if any):

### Risks / Open Items
-

### Next Round Plan
-
```

## 7. Definition of Done (per round)

A round is done only if:

1. Pre-Check is completed.
2. Code and docs for the round are updated.
3. Validation results are recorded.
4. Risks and next plan are documented.
