# ADRF Architecture

## Scope

This document describes the current `adrf` process only. It covers the
implemented HTTP resources, internal owners, persistence, lifecycle, and known
limits. It does not describe a deployment topology or the behavior of any ADRF
consumer.

## Runtime Structure

```text
cmd/main.go
  -> configuration and logging
  -> pkg/service.AdrfApp
       -> ADRF runtime context
       -> MongoDB repositories
       -> SBI processor
       -> HTTP server
       -> NRF management consumer
       -> TTL worker
```

The main owners are:

| Component | Responsibility |
| --- | --- |
| `pkg/factory` | Loads YAML configuration and applies runtime defaults. |
| `internal/context` | Owns the ADRF NF instance identity, registered addresses, advertised services, PLMN information, locality, and heartbeat interval. |
| `internal/sbi` | Registers Gin routes and maps HTTP requests to the processor. |
| `internal/sbi/processor` | Validates requests, coordinates storage, owns retrieval-subscription state, dispatches callbacks, and maps failures to HTTP responses. |
| `internal/store` | Persists Data Management and ML Model Management metadata in MongoDB. |
| `internal/sbi/consumer` | Registers the ADRF NF profile with the configured NRF, sends heartbeats, and deregisters on shutdown. |
| `internal/service/ttl_worker.go` | Periodically removes expired MongoDB records. |

The process currently serves HTTP. Selecting HTTPS in the configuration is
rejected at server startup because HTTPS serving is not implemented.

## Data Management

### Store records

`POST /nadrf-datamanagement/v1/data-store-records` accepts the implemented
`dataSub` plus `dataNotif` variant. The processor extracts the SUPI from
`dataSub.smfDataSub.supi`, assigns a store transaction identifier and ingestion
time, and persists the record in MongoDB.

The `Location` response header identifies the created record. The current
router does not expose an individual Data Store Record resource at that
location; retrieval uses the collection resource described below.

### Retrieve records

`GET /nadrf-datamanagement/v1/data-store-records` currently accepts exactly one
`fetch-correlation-ids` value. The identifier resolves to the stored
`storeTransId`. The alternative `store-trans-id` and `data-set-id` query forms,
and multi-identifier retrieval, are not implemented.

`POST /nadrf-datamanagement/v1/data-store-records/search` is a repository
extension. It returns paginated records filtered by optional SUPI and ingestion
time bounds.

### Retrieval subscriptions and snapshots

`POST /nadrf-datamanagement/v1/data-retrieval-subscriptions` performs the
following work:

1. It validates the notification correlation, absolute callback URI, time
   window, consumer-triggered notification flag, and SUPI scope.
2. It selects matching store transaction identifiers at a fixed snapshot
   boundary.
3. It writes the selected records to
   `./storage/snapshots/{subscriptionId}.json`.
4. It keeps the active subscription in process memory.
5. It returns `201 Created` and dispatches one or more asynchronous retrieval
   notifications.

Each callback contains the original `notifCorrId`, a batch of `fetchCorrIds`,
and a `fetchUri` pointing to the materialized snapshot. The last callback sets
`terminationReq` to `true`. An empty result still produces one terminating
callback.

`GET /nadrf-datamanagement/v1/data-snapshots/{snapshotId}/download` and the
snapshot representation are repository extensions. Deleting the retrieval
subscription cancels its callback context, removes the in-memory state, and
removes the snapshot file. Deletion is idempotent and returns `204` even when
the subscription is already absent.

Retrieval-subscription state is not persisted. A process restart loses active
subscription state, although previously written snapshot files remain until
explicitly removed.

## ML Model Management

### Create and import

`POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records` currently accepts one
model entry. The processor:

1. validates the metadata and HTTP model URL;
2. allocates a `storeTransId`;
3. downloads the model artifact to
   `{configuration.mlModelStorage.localDirectory}/{storeTransId}.tar.gz`;
4. verifies the downloaded byte count against `mlStorageSize`;
5. stores metadata in MongoDB; and
6. returns the stored representation and a `Location` header.

The returned `mLModelUrl` points to the repository-specific model download
resource owned by this ADRF instance.

### Query, update, and delete

The collection GET accepts exactly one of `store-trans-id` or
`model-unique-ids`. A query must resolve to one unambiguous record.

The item resource supports metadata replacement and deletion. Explicit DELETE
removes both the MongoDB record and the corresponding local artifact. The
repository-specific item GET returns the stored representation, and the
`/model` child resource streams the local `tar.gz` artifact.

`allowConsumerList` is validated and stored as metadata, but the current model
download handler does not authenticate the caller or enforce that list.

## Persistence

MongoDB contains two primary collections:

| Collection | Contents |
| --- | --- |
| `data_store_records` | Stored data subscription context, notification content, SUPI, store transaction identity, ingestion time, and optional expiry time. |
| `mlmodel_store_records` | Model identity, ADRF and source addresses, size, consumer metadata, storage result, timestamps, and optional expiry time. |

Model bytes are stored outside MongoDB in the configured local model directory.
Retrieval snapshots are stored under `./storage/snapshots`.

The TTL worker runs every 60 seconds and deletes expired MongoDB documents.
Explicit ML Model DELETE also removes the artifact file. TTL removal of a model
record does not currently remove the corresponding local artifact file.

## NRF Lifecycle

When `configuration.nrfUri` is non-empty, ADRF builds and registers an NF
profile containing:

- NF type `ADRF`;
- the configured service list and endpoints;
- PLMN, S-NSSAI, locality, and address information; and
- `adrfInfoList` advertising data and ML model storage capability.

The process periodically attempts an NF status heartbeat and deregisters on
graceful shutdown. The current heartbeat request does not serialize the JSON
Patch body prepared by the lifecycle loop. If no NRF URI is configured, ADRF
runs without registration.

## Configuration

The sample file is [`../config/adrfcfg.yaml`](../config/adrfcfg.yaml). The
current runtime uses these main settings:

- NF identity, NRF URI, locality, heartbeat interval, and advertised services;
- SBI scheme, bind address, registered address, and port;
- PLMN and S-NSSAI support;
- MongoDB database name and URI;
- retrieval callback batch size; and
- local ML model storage directory.

The current retrieval path always creates a snapshot. The sample
`retrieval.oneIdPerFetch` and `retrieval.snapshot.enabled` values do not change
that behavior in the current implementation.

## Current Limits

- HTTP serving only.
- MongoDB is required for functional storage operations.
- Only the `dataSub` plus `dataNotif` Data Store Record variant is supported.
- Data retrieval supports one `fetch-correlation-ids` value per GET.
- Retrieval-subscription state is volatile.
- Snapshot search and download are repository extensions.
- Snapshot and model download resources do not authenticate callers.
- ML Model creation supports one model entry per request.
- Model consumer metadata is not enforced as download authorization.
- TTL cleanup removes expired metadata but does not remove expired model files.
- The NRF heartbeat request currently carries no JSON Patch body.
