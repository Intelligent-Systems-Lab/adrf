# ADRF (Analytics Data Repository Function)

This repository (`adrf`) implements the **Analytics Data Repository Function
(ADRF)** microservice for 5G Core Networks. It contains a Release 18
interoperability profile and several repository-specific extensions; those
extensions are not part of the standardized TS 29.575 resource set.

The exact boundary, implementation provenance, and NWDAF/PyMTLF/PyAnLF usage
are documented in
[Release 18 ADRF Interoperability Profile](docs/impl/release18-interoperability-profile.md).

---

## 🏛️ Architecture & System Positioning

ADRF serves as the central storage repository for historical network analytics records and Machine Learning (ML) models within the 5G Core architecture. It aligns with the free5gc microservice framework and interacts directly with **NWDAF (`/home/king25158986/NWDAF`)** and **MTLF-subp (`/home/king25158986/MTLF-subp`)**.

```mermaid
graph LR
    subgraph VM ["5GC Virtual Machine (192.168.107.5)"]
        NWDAF["Go NWDAF Core<br/>(AnLF / SBI Gateway)"]
        ADRF["ADRF Service<br/>(Port 9888 / MongoDB)"]
        MTLF["MTLF-subp Gateway<br/>(Python Microservice)"]
    end

    NWDAF -- "1. POST /data-store-records" --> ADRF
    NWDAF -- "2. POST /data-retrieval-subscriptions" --> ADRF
    ADRF -- "3. POST Notification (/collector/retrieval-notify)" --> NWDAF
    NWDAF -- "4. Forward Notification (/api/v1/mtlf/adrf-callback)" --> MTLF
    MTLF -- "5. GET /data-store-records?fetch-correlation-ids=..." --> ADRF
    NWDAF -- "6. DELETE /data-retrieval-subscriptions/{id}" --> ADRF
    NWDAF -- "7. POST /mlmodel-store-records (Register Model)" --> ADRF
```

---

## 🔄 Snapshot Export & ML Model Registration Lifecycle

```mermaid
sequenceDiagram
    autonumber
    participant NWDAF as Go NWDAF Core (:8080)
    participant ADRF as ADRF Service (:9888)
    participant Mongo as MongoDB / File Store
    participant MTLF as MTLF-subp Gateway (:9887)

    NWDAF->>ADRF: POST /data-store-records (Analytics Records)
    ADRF->>Mongo: Store Record in data_store_records
    ADRF-->>NWDAF: 201 Created (Location: /data-store-records/{id})

    NWDAF->>ADRF: POST /data-retrieval-subscriptions (NotifURI: /collector/retrieval-notify)
    ADRF->>Mongo: Query Records & Export JSON Snapshot
    Mongo-->>ADRF: Saved ./storage/snapshots/{subId}.json
    ADRF-->>NWDAF: 201 Created (Location: /data-retrieval-subscriptions/{subId})

    ADRF->>NWDAF: POST /collector/retrieval-notify (fetchInstruct.fetchUri)
    NWDAF-->>ADRF: 204 No Content
    NWDAF->>MTLF: Forward Notification to /api/v1/mtlf/adrf-callback

    MTLF->>ADRF: GET /data-store-records?fetch-correlation-ids={id}
    ADRF->>Mongo: Read one matching store record
    ADRF-->>MTLF: 200 OK (NadrfDataStoreRecord)

    NWDAF->>ADRF: DELETE /data-retrieval-subscriptions/{subId}
    ADRF->>Mongo: Delete ./storage/snapshots/{subId}.json & Clear Sub State
    ADRF-->>NWDAF: 204 No Content

    Note over NWDAF,ADRF: ML Model Registration Phase (TS 29.575 Clause 4.3)
    NWDAF->>ADRF: POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records (Staging Model URL)
    ADRF->>Mongo: Store Model Metadata in mlmodel_store_records
    ADRF-->>NWDAF: 201 Created (Location: /mlmodel-store-records/{storeTransId})
```

---

## 🔑 Supported 3GPP Service Operations (TS 29.575)

### 1. `Nadrf_DataManagement` Service

#### `POST /nadrf-datamanagement/v1/data-store-records`
* Stores analytics data records collected from NWDAF / UPF into MongoDB / Memory store.
* **HTTP Response**: `201 Created` with `Location` header pointing to the created record URI.

#### `POST /nadrf-datamanagement/v1/data-retrieval-subscriptions`
* Subscribes to historical data retrieval.
* Triggers snapshot generation, exporting matched records into a local JSON snapshot file (`./storage/snapshots/{subscriptionId}.json`).
* Asynchronously sends a `NadrfDataRetrievalNotification` callback to the subscriber's notification URI (`http://192.168.107.5:8080/collector/retrieval-notify`) containing fetch instructions.
* **HTTP Response**: `201 Created` with `Location: /nadrf-datamanagement/v1/data-retrieval-subscriptions/{subscriptionId}` and `NadrfDataRetrievalSubscription` response body.

#### `GET /nadrf-datamanagement/v1/data-snapshots/{subscriptionId}/download` (`fetchUri`)
* Provides RESTful HTTP GET dataset download for exported snapshot files.
* Allows consumers (such as `MTLF-subp`) to pull historical analytics data without requiring shared disk mounts.
* **Extension**: this resource and its JSON-array response are not defined by
  the Release 18 `Nadrf_DataManagement` OpenAPI. The NWDAF/PyMTLF profile uses
  `fetchCorrIds` with the standard `/data-store-records` collection GET
  instead.

#### `DELETE /nadrf-datamanagement/v1/data-retrieval-subscriptions/{subscriptionId}` (`RetrievalUnsubscribe`)
* Cancels an active data retrieval subscription.
* **Cleanup Lifecycle**: Removes the subscription from memory/MongoDB and automatically deletes the exported temporary snapshot JSON file (`./storage/snapshots/{subscriptionId}.json`).
* **HTTP Response**: `204 No Content`.

---

### 2. `Nadrf_MLModelManagement` Service

#### `POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records`
* Registers newly trained ML model metadata and download URLs produced by NWDAF / MTLF.
* **HTTP Response**: `201 Created` with `Location` identifying the created
  store-record resource
  (`/nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}`).
* The returned `mlFileAddr` identifies the artifact download URL. This
  implementation may use the non-standard `.../{storeTransId}/model` endpoint
  for that address.

---

## 📑 OpenAPI 3.0 Schema Alignment

The following example shows the repository-specific snapshot extension. It is
not the Release 18 RetrievalRequest response contract:

```json
{
  "notifCorrId": "sub-12345",
  "timeStamp": "2026-07-30T09:00:00Z",
  "fetchInstruct": {
    "fetchUri": "http://192.168.107.5:9888/nadrf-datamanagement/v1/data-snapshots/sub-12345/download",
    "fetchCorrIds": ["store-transaction-1"]
  }
}
```

---

## 🧪 Building & Running

### Building Binary
```bash
go build -o bin/adrf ./cmd/main.go
```

### Running Service
```bash
./bin/adrf --config ./config/adrfcfg.yaml
```

### Running Unit Tests
```bash
go test -v -short ./internal/sbi/processor/...
```
* **Test Status**: `100% PASSED` (0.012s).
