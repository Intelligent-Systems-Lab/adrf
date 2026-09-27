# ADRF

This repository implements an Analytics Data Repository Function (ADRF) with
Data Management and ML Model Management services. The process stores metadata
in MongoDB, stores model artifacts on the local filesystem, exposes HTTP SBI
routes, and registers its service profile with an NRF when an NRF URI is
configured.

The implementation follows the free5GC application structure, but it also
contains explicitly identified project extensions. See
[`docs/architecture.md`](docs/architecture.md) for the implemented operations,
internal ownership, persistence model, and current limitations.

## Implemented Services

### Data Management

| Method and path | Purpose | Boundary |
| --- | --- | --- |
| `POST /nadrf-datamanagement/v1/data-store-records` | Store one data record | Release 18-shaped operation |
| `GET /nadrf-datamanagement/v1/data-store-records?fetch-correlation-ids=...` | Retrieve one stored record | Release 18-shaped operation with a one-ID implementation limit |
| `POST /nadrf-datamanagement/v1/data-retrieval-subscriptions` | Create a retrieval snapshot and send fetch instructions | Release 18-shaped resource with repository-specific snapshot behavior |
| `DELETE /nadrf-datamanagement/v1/data-retrieval-subscriptions/{subscriptionId}` | Remove retrieval state and its snapshot file | Release 18 operation |
| `POST /nadrf-datamanagement/v1/data-store-records/search` | Search records by SUPI and ingestion time | Repository extension |
| `GET /nadrf-datamanagement/v1/data-snapshots/{snapshotId}/download` | Download a materialized JSON snapshot | Repository extension |

### ML Model Management

| Method and path | Purpose | Boundary |
| --- | --- | --- |
| `POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records` | Import a model artifact and create its metadata record | Release 18 operation |
| `GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records` | Retrieve a record by `store-trans-id` or `model-unique-ids` | Release 18 operation |
| `PUT /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}` | Replace stored model metadata | Release 18 operation |
| `DELETE /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}` | Delete metadata and the local artifact | Release 18 operation |
| `GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}` | Retrieve one record by resource path | Repository extension |
| `GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}/model` | Download the stored model artifact | Repository extension |

## Requirements

- Go 1.25.5
- MongoDB reachable through `configuration.mongodb.url`
- A writable directory for model artifacts
- An NRF only when registration and discovery visibility are required

## Build, Run, and Test

```bash
go build -o bin/adrf ./cmd/main.go
./bin/adrf --config ./config/adrfcfg.yaml
go test ./...
```

The sample configuration binds the service to `127.0.0.1:9888`. Update the
binding address, registered address, NRF URI, PLMN information, MongoDB URI,
and model storage path for the target deployment.
