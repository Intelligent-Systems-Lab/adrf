# Release 18 ADRF Interoperability Profile

## 1. Purpose

This document records the Release 18 interoperability profile used by the
NWDAF, PyAnLF, and PyMTLF deployment. It also distinguishes standardized ADRF
resources from repository-specific extensions.

The normative procedure source for this profile is 3GPP TS 29.575 V18.11.0.
The exact wire schemas used by the two services are:

- `Nadrf_DataManagement`: 3GPP
  `TS29575_Nadrf_DataManagement.yaml` version 1.1.4, associated with
  TS 29.575 V18.11.0;
- `Nadrf_MLModelManagement`: 3GPP
  `TS29575_Nadrf_MLModelManagement.yaml` version 1.0.1, associated with
  TS 29.575 V18.7.0;
- ADRF NRF registration capability: 3GPP
  `TS29510_Nnrf_NFManagement.yaml` version 1.3.4, associated with
  TS 29.510 V18.11.0.

The simplified V19.6.0 YAML bundled in this repository is not the
wire-contract source of truth for this Release 18 profile.

### 1.1 Evidence notation

Normative quotations below identify their 3GPP specification version and
clause. Schema excerpts identify the corresponding 3GPP OpenAPI document and
version. Excerpts omit unrelated responses and referenced definitions, but
preserve the relevant path, field, cardinality, and response shape.

The evidence is separated deliberately:

| Evidence | What it proves |
| --- | --- |
| TS 29.575 procedure text | Required service behavior and HTTP outcome |
| Release 18 OpenAPI | Exact path, method, query name, and JSON schema |
| ADRF Git history | When and by whom the local implementation was introduced |
| Project profile | Which valid standard option this deployment selects |

## 2. Implementation provenance

The following history is important when attributing existing behavior:

| Date and time (UTC+8) | Commit | Author | Change |
| --- | --- | --- | --- |
| 2026-03-26 11:45 | `efa7e0d` | `haha39 <leo0390390@gmail.com>` | Registered the Data Management route skeleton. The collection GET still returned `501 Not Implemented`. |
| 2026-03-26 17:12 | `af401d8` | `haha39 <leo0390390@gmail.com>` | Implemented single-ID retrieval through the standard `GET /data-store-records` collection resource. |
| 2026-07-23 08:59 | `04e0cba` | `hokusai0603 <cgmasterminds.ai14@nycu.edu.tw>` | Added the initial ML Model Management implementation, including both collection-query and individual-path GET handlers. |
| 2026-07-31 00:30 | `b0ea497` | `hokusai0603 <cgmasterminds.ai14@nycu.edu.tw>` | Added exported JSON snapshots and the custom `/data-snapshots/{subscriptionId}/download` endpoint, then advertised that endpoint in `fetchUri`. |

The standard Data Management storage and retrieval vertical therefore existed
before the federated-learning final-model publication work. The current
federated-learning integration did not introduce the ADRF Data Management
routes.

## 3. Data Management profile

### 3.1 Standard resources used by the deployment

The deployment uses these TS 29.575 resources:

```http
POST /nadrf-datamanagement/v1/data-store-records

POST /nadrf-datamanagement/v1/data-retrieval-subscriptions

GET /nadrf-datamanagement/v1/data-store-records
    ?fetch-correlation-ids=<fetch-correlation-id>

DELETE /nadrf-datamanagement/v1/data-retrieval-subscriptions/{subscriptionId}
```

3GPP TS 29.575 V18.11.0 clause 4.2.2.5.2 states:

> “The NF service consumer shall send an HTTP GET request with
> `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records` as Resource
> URI representing the ‘ADRF Data Store Records’ resource.”

The same clause defines the successful response:

> “If the requested data or analytics is found, the ADRF shall respond with
> ‘200 OK’ status code with the message body containing the
> `NadrfDataStoreRecord` data structure.”

3GPP `TS29575_Nadrf_DataManagement.yaml` version 1.1.4, associated with
TS 29.575 V18.11.0, makes the collection query and singular response schema
explicit:

```yaml
/data-store-records:
  get:
    parameters:
      - name: store-trans-id
        in: query
        schema:
          type: string
      - name: fetch-correlation-ids
        in: query
        style: form
        explode: false
        schema:
          type: array
          items:
            type: string
          minItems: 1
      - name: data-set-id
        in: query
        schema:
          type: string
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NadrfDataStoreRecord'
      '204':
        description: No matching ADRF data were found.
```

The response reference is one `NadrfDataStoreRecord`; it is not an array of
records. The word “records” in the resource name and response description does
not change the declared JSON schema.

The record itself may contain lists inside one envelope, which is different
from returning a top-level array of store records:

```yaml
NadrfDataStoreRecord:
  type: object
  oneOf:
    - allOf:
        - required: [anaSub]
        - required: [anaNotifications]
    - allOf:
        - required: [dataSub]
        - required: [dataNotif]
  properties:
    dataNotif:
      $ref: '#/components/schemas/DataNotification'
    anaNotifications:
      type: array
    anaSub:
      type: array
    dataSub:
      type: array
```

The retrieval callback carries `fetchInstruct.fetchCorrIds`. PyMTLF uses each
identifier with the standard `data-store-records` collection GET. A successful
request returns one `NadrfDataStoreRecord`; no matching record returns
`204 No Content`.

Although `fetch-correlation-ids` is an array in the OpenAPI definition, the
current ADRF and PyMTLF profile intentionally sends one identifier per request.
One value is valid under the standard cardinality and avoids inventing a bulk
response envelope where the Release 18 response schema is a single
`NadrfDataStoreRecord`.

### 3.2 How PyMTLF handles `fetchUri`

3GPP TS 29.575 V18.11.0 clause 4.2.2.8 states that the mandatory `fetchUri`
is expected to match the standard RetrievalRequest resource, but may contain
another value because an ADRF consumer can perform retrieval using the fetch
correlation identifiers.

The relevant Release 18 note states:

> “The fetch correlation identifiers included in the fetch instructions …
> can be used to fetch data or analytics using the
> `Nadrf_DataManagement_RetrievalRequest` service operation.”

It then clarifies:

> “The (mandatory) fetch URI … is expected to be in line with the standard
> resource URI … `/data-store-records`, but it can be anything because it is
> actually not needed by the NF service consumer in this case.”

This is the direct normative basis for treating `fetchCorrIds`, rather than
dereferencing an arbitrary `fetchUri`, as the interoperable retrieval input.

The referenced `FetchInstruction` comes from 3GPP
`TS29576_Nmfaf_3caDataManagement.yaml` version 1.1.2, associated with
TS 29.576 V18.7.0, and requires both fields:

```yaml
FetchInstruction:
  type: object
  required:
    - fetchUri
    - fetchCorrIds
  properties:
    fetchUri:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uri'
    fetchCorrIds:
      type: array
      items:
        type: string
      minItems: 1
    expiry:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/DateTime'
```

The project does not discard `fetchUri`; it preserves and validates the
notification shape. It simply does not treat that field as permission to
replace the standard ADRF retrieval schema.

PyMTLF therefore:

1. selects an ADRF through fixed configuration or NRF discovery;
2. retains the callback `fetchUri` as received for route-consistency
   diagnostics;
3. does not dereference that URI as an assumed standard response contract;
4. constructs
   `{selectedApiRoot}/nadrf-datamanagement/v1/data-store-records`;
5. supplies one `fetch-correlation-ids` value and validates the returned
   `NadrfDataStoreRecord`.

This choice prevents an arbitrary callback URI from changing the expected
response schema or redirecting dataset retrieval to a different origin.

### 3.3 Non-standard snapshot extension

The repository also exposes:

```http
GET /nadrf-datamanagement/v1/data-snapshots/{subscriptionId}/download
```

This endpoint returns a JSON array containing multiple stored records. Neither
the resource path nor that bulk array response is defined by the Release 18
`Nadrf_DataManagement` OpenAPI.

The complete Release 18 Data Management `paths` set is:

```text
/data-store-records
/data-store-records/{storeTransId}
/data-retrieval-subscriptions
/data-retrieval-subscriptions/{subscriptionId}
/request-storage-sub
/request-storage-sub-removal
/remove-stored-data-analytics
```

There is no `/data-snapshots/{subscriptionId}/download` resource. Its response
array also cannot be substituted for the single `NadrfDataStoreRecord`
declared for the standard collection GET.

The extension itself may remain available for consumers that explicitly adopt
its private contract. It must not be documented as the standardized
`Nadrf_DataManagement_RetrievalRequest`, and the NWDAF/PyMTLF profile does not
depend on it.

The compatibility failure discovered during integration was:

```text
ADRF callback advertised the custom snapshot URL
  -> PyMTLF dereferenced fetchUri as if it were the standard collection GET
  -> snapshot endpoint returned a JSON array
  -> PyMTLF expected one NadrfDataStoreRecord
  -> schema validation failed
```

The correction is on the PyMTLF consumer side: retrieval is now based on the
selected ADRF API root and `fetchCorrIds`. The ADRF Data Management handlers,
repository, and stored-data schema were not modified by this correction.

## 4. ML Model Management profile

### 4.1 Standard resources used by the deployment

The Release 18 profile uses:

```http
POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records

GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records
    ?store-trans-id=<store-transaction-id>

GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records
    ?model-unique-ids=<numeric-model-id>

PUT /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}

DELETE /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}
```

3GPP TS 29.575 V18.11.0 clause 4.3.2.3.2 defines one collection retrieval
operation:

> “The NF service consumer shall send an HTTP GET request with
> `{apiRoot}/nadrf-mlmodelmanagement/<apiVersion>/mlmodel-store-records` as
> Resource URI representing the ‘ADRF ML Model Store Records’ resource.”

The same sentence says that retrieval is selected by:

> “the storage transaction identifier within the `store-trans-id` query
> parameter or the unique ML model identifier(s) within the
> `model-unique-ids` query parameter.”

Its response remains one record representation:

> “If one or more of the requested ML model(s) are found, the ADRF shall
> respond with ‘200 OK’ status code with the message body containing the
> `NadrfMLModelStoreRecord` data structure.”

3GPP `TS29575_Nadrf_MLModelManagement.yaml` version 1.0.1, associated with
TS 29.575 V18.7.0, expresses the same contract:

```yaml
/mlmodel-store-records:
  get:
    operationId: GetAdrfMLModelStoreRecord
    parameters:
      - name: store-trans-id
        in: query
        schema:
          type: string
      - name: model-unique-ids
        in: query
        schema:
          type: array
          items:
            $ref: 'TS29571_CommonData.yaml#/components/schemas/Uinteger'
    responses:
      '200':
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/NadrfMLModelStoreRecord'
      '204':
        description: No matching ADRF ML Model were found.
```

3GPP TS 29.575 V18.11.0 API table 5.2.3.2.3.2-1 additionally states that
exactly one of `store-trans-id` and `model-unique-ids` shall be provided. Table
5.2.3.2.3.2-3 assigns cardinality `1` to the `NadrfMLModelStoreRecord` in a
`200 OK` response.

Exactly one of `store-trans-id` and `model-unique-ids` is supplied to the
collection GET. A successful retrieval returns one
`NadrfMLModelStoreRecord`; no matching record returns `204 No Content`.

The two identifiers have different ownership:

- `storeTransId` is allocated by ADRF and identifies an ADRF store record.
- `modelUniqueId` is allocated by the model provider and identifies a model
  independently of the ADRF storage transaction.

The project profile stores and retrieves one completed model per request.
`modelUniqueId` is represented as the Release 18 `Uinteger`, not as an
implementation-defined string.

### 4.2 Existing non-standard ML model resources

Commit `04e0cba` introduced both of these GET resources:

```http
GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records
    ?model-unique-ids=...

GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}
```

Only the collection-query form is the Release 18 retrieval operation. The
Release 18 individual resource defines PUT and DELETE, but not GET.

The corresponding path in 3GPP
`TS29575_Nadrf_MLModelManagement.yaml` version 1.0.1 begins:

```yaml
/mlmodel-store-records/{storeTransId}:
  delete:
    operationId: DeleteADRFMLModelStoreRecord
    parameters:
      - name: storeTransId
        in: path
        required: true
  put:
    operationId: UpdateADRFMLModelStoreRecord
    parameters:
      - name: storeTransId
        in: path
        required: true
```

No `get` operation is declared under that path. The local
`GET /mlmodel-store-records/{storeTransId}` handler is therefore an extension
relative to the Release 18 profile, even though using a path parameter for an
individual resource is a conventional REST design.

The repository also exposes:

```http
GET /nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}/model
```

That path is not a standardized ML Model Management resource. It is an
implementation-specific artifact download endpoint. It can be returned as the
model address in `MLModelInfo.mlFileAddr`, because the model file address
identifies where the consumer can download the artifact; it must not be
presented as an additional standardized service operation.

The schema supports an implementation-selected artifact address without
standardizing the artifact-serving resource itself:

```yaml
MLModelInfo:
  required:
    - modelUniqueId
    - mlFileAddr
    - mlStorageSize
  properties:
    modelUniqueId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uinteger'
    mlFileAddr:
      $ref: 'TS29520_Nnwdaf_MLModelProvision.yaml#/components/schemas/MLModelAddr'
    mlStorageSize:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uinteger'
    allowConsumerList:
      type: array
      items:
        $ref: '#/components/schemas/AllowedConsumer'

AllowedConsumer:
  type: object
  properties:
    nfInstanceId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfInstanceId'
    nfSetId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfSetId'
  oneOf:
    - required: [nfInstanceId]
    - required: [nfSetId]

MLModelAddr:
  type: object
  properties:
    mLModelUrl:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uri'
    mlFileFqdn:
      type: string
  oneOf:
    - required: [mLModelUrl]
    - required: [mlFileFqdn]
```

Consequently, an ADRF-hosted `/.../{storeTransId}/model` URL is a valid value
for `mLModelUrl`, but that does not make the URL path a standardized
Nadrf_MLModelManagement operation.

The access check is also explicitly described in 3GPP TS 29.575 V18.11.0
clause 4.3.2.3.2:

> “If the NF Service Consumer is not included in the allowed NF consumer list
> for the ML model and/or is not same as the NF of the NWDAF containing MTLF
> that stored the model, the ADRF shall send an HTTP "403 Forbidden" error
> response including the "cause" attribute set to
> "RETRIEVAL_ML_MODEL_NOT_ALLOWED".”

The current plain-HTTP experiment does not provide ADRF with an authenticated
caller NF identity. ADRF persists `allowConsumerList`, and PyAnLF rejects a
record that does not authorize its containing NWDAF, but this consumer-side
check is not a substitute for the ADRF-side authorization required by the
quoted procedure. ADRF-side `403/RETRIEVAL_ML_MODEL_NOT_ALLOWED` enforcement
remains an explicit security limitation of this experiment profile.

The NWDAF/PyMTLF/PyAnLF integration:

- uses the standard collection GET for record retrieval;
- does not depend on the non-standard individual GET;
- follows the `mlFileAddr` returned inside the validated record to download
  the model artifact.

## 5. Exact changes made for the federated-learning integration

### 5.1 ADRF repository changes

The implementation baseline before this integration was commit `b656b08`.
Compared with that baseline, the production-code changes are limited to the
following files and behaviors:

| Code area | Concrete change |
| --- | --- |
| `internal/context/context.go` | Replaced capability flags under `customInfo` with a compatibility NF profile that serializes top-level Release 18 `adrfInfoList`, advertising both model and data storage. |
| `internal/sbi/consumer/nrf_service.go` | Replaced generated-profile registration with an explicit JSON `PUT /nnrf-nfm/v1/nf-instances/{nfInstanceId}` so `adrfInfoList` is not dropped by the older generated model. |
| `internal/sbi/processor/mlmodel_request.go` | Replaced the private ML model request/response shape with the supported Release 18 URL-backed profile; corrected POST validation, artifact download, response representation, collection-query retrieval, and standard error mapping. |
| `internal/store/mlmodel_repository.go` | Changed persisted model identity from string to integer, added `allowConsumerList` and nested `modelStoreResult`, and made `modelUniqueId` unique. |
| `internal/sbi/processor/processor.go` | Updated the ML model repository interface to query numeric model identifiers. |
| `internal/sbi/processor/mlmodel_request_test.go` | Updated the storage/retrieval lifecycle tests and added assertions for the corrected schema and failure behavior. |

No ADRF Data Management production handler, retrieval-subscription query, data
store repository, or single-ID fetch implementation was changed by this
integration. Data Management changes in this worktree are documentation
corrections that distinguish the standard collection GET from the private
snapshot-download extension.

#### 5.1.1 NRF registration delta

Before this integration, ADRF placed `mlModelStorageInd` and
`dataStorageInd` in `customInfo`. The Release 18 NRF schema defines them under
top-level `adrfInfoList`, so a standard discovery consumer could not rely on
the previous registration shape.

The integration now:

1. builds `adrfInfoList.default.mlModelStorageInd=true`;
2. builds `adrfInfoList.default.dataStorageInd=true`;
3. preserves the existing NF identity, PLMN, S-NSSAI, address, status, and
   service list;
4. serializes the compatibility profile directly into the NRF registration
   request, avoiding loss of `adrfInfoList` in the pinned generated model.

Heartbeat update and deregistration behavior were not redesigned.

#### 5.1.2 ML model wire-schema delta

The previous implementation used:

```json
{
  "modelUniqueId": "uuid-like-string",
  "mlFileAddr": "http://source/model",
  "storeResult": "..."
}
```

The supported store request now uses:

```json
{
  "nfInstanceId": "nwdaf-c-instance",
  "mlModelInfo": [
    {
      "modelUniqueId": 1720000000001,
      "mlFileAddr": {
        "mLModelUrl": "http://source/model"
      },
      "mlStorageSize": 4096,
      "allowConsumerList": [
        {
          "nfInstanceId": "nwdaf-a-instance"
        }
      ]
    }
  ]
}
```

After ADRF has downloaded and stored the artifact, the `201 Created` response
uses the ADRF-hosted artifact address and adds the storage result:

```json
{
  "nfInstanceId": "nwdaf-c-instance",
  "mlModelInfo": [
    {
      "modelUniqueId": 1720000000001,
      "mlFileAddr": {
        "mLModelUrl": "http://adrf.example/nadrf-mlmodelmanagement/v1/mlmodel-store-records/store-001/model"
      },
      "mlStorageSize": 4096,
      "allowConsumerList": [
        {
          "nfInstanceId": "nwdaf-a-instance"
        }
      ]
    }
  ],
  "modelStoreResult": {
    "modelUniqueId": 1720000000001,
    "storeResult": "ML_MODEL_FILE_STORED_IN_ADRF"
  }
}
```

The concrete corrections are:

- `modelUniqueId`: string to non-negative integer;
- `mlFileAddr`: string to `MLModelAddr` object;
- `storeResult`: private top-level string to standard `modelStoreResult`;
- `allowConsumerList`: accepted, persisted, and returned;
- owner: exactly one of `nfInstanceId` and `nfSetId`;
- address: exactly one of `mLModelUrl` and `mlFileFqdn`.

#### 5.1.3 ML model storage behavior delta

For `POST /mlmodel-store-records`, ADRF now:

1. validates the supported Release 18 representation before creating state;
2. accepts exactly one URL-backed `mlModelInfo` entry;
3. downloads the model from `mLModelUrl` with a five-minute timeout;
4. verifies the downloaded byte count equals `mlStorageSize`;
5. deletes a partial file and creates no retrievable MongoDB record when
   download, size validation, or persistence fails;
6. stores the completed file under its ADRF-assigned `storeTransId`;
7. returns `201 Created`, a `Location` for the created store record, and a
   corrected `NadrfMLModelStoreRecord` representation.

A missing source URL maps to
`404/ML_MODEL_FILE_ADDRESS_NOT_FOUND`. Transport, download, and size failures
map to `500/ML_MODEL_FILE_DOWNLOAD_FAILED`.

#### 5.1.4 ML model retrieval behavior delta

For the standard collection GET, ADRF now:

1. requires exactly one query selector: `store-trans-id` or
   `model-unique-ids`;
2. parses `model-unique-ids` as non-negative integers;
3. returns one `NadrfMLModelStoreRecord`, rather than the previous private
   top-level array;
4. returns `204 No Content` when no record matches;
5. treats multiple matched records as an ambiguous server-side failure
   because the supported response envelope is singular.

The existing `GET /mlmodel-store-records/{storeTransId}` remains a repository
extension. The integration consumers use the standard collection GET and do
not depend on that extension.

#### 5.1.5 Persistence and test delta

MongoDB now stores numeric `modelUniqueId`, structured `mlModelInfo`,
`allowConsumerList`, and structured `modelStoreResult`. A unique index on
`modelUniqueId` prevents an ambiguous duplicate during publication recovery.

The updated processor tests cover:

- successful URL-backed store, record retrieval, artifact download, update,
  and delete;
- corrected integer identifiers and response representation;
- source `404` without creation of a retrievable record;
- owner, address, model-entry, and allowed-consumer validation.

#### 5.1.6 Deliberate restrictions

The integration does not implement every Release 18 storage alternative. It
intentionally:

- accepts one URL-backed model per store request;
- rejects inline `mlModels`;
- rejects FQDN-backed storage;
- uses a project-selected five-minute download timeout;
- requires one unambiguous record from a collection retrieval.

These restrictions are narrower than the complete Release 18 feature set.
The standard permits multiple model entries and does not mandate the unique
MongoDB index or timeout value.

#### 5.1.7 Standards basis

The wire-contract corrections follow the storage procedure in 3GPP
TS 29.575 V18.11.0 clause 4.3.2.2.2:

> “The `NadrfMLModelStoreRecord` data structure provided in the request body
> shall include either the `MLModelInfo` data structure in the `mlModelInfo`
> attribute or the `MLModel` data structure in the `mlModels` attribute.”

For URL-backed storage, the same clause requires `modelUniqueId`,
`mlFileAddr`, and `mlStorageSize`, and permits `allowConsumerList`. It further
requires ADRF to assign a `storeTransId`, download the model if needed, store
it, and return `201 Created` with a `Location` identifying the created record.

The record schema supporting these checks is defined by 3GPP
`TS29575_Nadrf_MLModelManagement.yaml` version 1.0.1:

```yaml
NadrfMLModelStoreRecord:
  allOf:
    - oneOf:
        - required: [nfInstanceId]
        - required: [nfSetId]
    - anyOf:
        - required: [mlModelInfo]
        - required: [mlModels]
  properties:
    nfInstanceId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfInstanceId'
    nfSetId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfSetId'
    mlModelInfo:
      type: array
      items:
        $ref: '#/components/schemas/MLModelInfo'
      minItems: 1
    mlModels:
      type: array
      items:
        $ref: '#/components/schemas/MLModel'
      minItems: 1
    modelStoreResult:
      $ref: '#/components/schemas/ModelStoreResult'

ModelStoreResult:
  required:
    - modelUniqueId
    - storeResult
  properties:
    modelUniqueId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uinteger'
    storeResult:
      $ref: '#/components/schemas/StoreResult'
```

The error mapping is also procedure evidence, not a project invention. Clause
4.3.2.2.2 requires:

> “If the ML model file address(es) was/were not found, the ADRF shall send an
> HTTP "404 Not Found" status code with the response body containing a
> ProblemDetails data structure with the "cause" attribute including the
> "ML_MODEL_FILE_ADDRESS_NOT_FOUND" application error response.”

It requires:

> “If the ML model file(s) download failed, the ADRF shall send an HTTP "500
> Internal Server Error" status code with the response body containing a
> ProblemDetails data structure with the "cause" attribute including the
> "ML_MODEL_FILE_DOWNLOAD_FAILED" application error response.”

The NRF correction adds the Release 18 `adrfInfoList` with
`mlModelStorageInd` and `dataStorageInd`. A compatibility wrapper and full JSON
registration request preserve those fields when the pinned free5GC generated
NF profile does not contain them.

3GPP `TS29510_Nnrf_NFManagement.yaml` version 1.3.4, associated with
TS 29.510 V18.11.0, defines that capability shape:

```yaml
adrfInfoList:
  type: object
  additionalProperties:
    $ref: '#/components/schemas/AdrfInfo'
  minProperties: 1

AdrfInfo:
  type: object
  properties:
    mlModelStorageInd:
      type: boolean
      default: false
    dataStorageInd:
      type: boolean
      default: false
```

This is why placing only similarly named values in `customInfo` was not
equivalent to advertising the standardized ADRF capability.

### 5.2 NWDAF, PyMTLF, and PyAnLF usage

The integration work outside ADRF is separated by owner:

| Component | Implemented responsibility |
| --- | --- |
| Go NWDAF | Added Release 18-compatible ADRF ML model record types and standard-shaped store/retrieval forwarding. It validates and forwards SBI messages but does not proxy model bytes or decide publication. |
| PyMTLF | After final validation, publishes the completed model by URL, validates ADRF's `201 + Location + NadrfMLModelStoreRecord`, records ADRF identity and transaction metadata, and provisions the ADRF reference to PyAnLF. |
| PyAnLF | Resolves the referenced ADRF, retrieves the store record through the standard collection GET, validates owner/consumer/model metadata, downloads `mlFileAddr`, and activates the model only after artifact validation succeeds. |
| ADRF | Downloads and persists the completed model, exposes the standard store-record representation, and serves the stored artifact through the address carried in `mlFileAddr`. |

The final-model flow is:

```text
PyMTLF final validation succeeds
  -> PyMTLF asks its containing Go NWDAF to send the standard ADRF store request
  -> ADRF downloads and stores the completed model
  -> ADRF returns 201, Location, and NadrfMLModelStoreRecord
  -> PyMTLF records modelUniqueId, storeTransId, ADRF identity, and publication state
  -> Model Provision notifies PyAnLF with mLModelAdrf
  -> PyAnLF asks its containing Go NWDAF to use the standard collection GET
  -> PyAnLF validates owner, allowed consumer, model ID, size, and artifact
  -> PyAnLF activates the new model and completes monitor cutover
```

Go remains the standardized communication boundary for ML model store and
record retrieval. PyMTLF owns publication state and model identity; PyAnLF owns
artifact activation. ADRF stores the completed artifact and its access
metadata, but does not decide which model is latest.

The Model Provision reference used for that handoff is defined by 3GPP
`TS29520_Nnwdaf_MLModelProvision.yaml` version 1.1.4, associated with
TS 29.520 V18.13.0, rather than by a project-only field:

```yaml
MLEventNotif:
  properties:
    mLFileAddr:
      $ref: '#/components/schemas/MLModelAddr'
    mLModelAdrf:
      $ref: '#/components/schemas/MLModelAdrf'
    modelUniqueId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/Uinteger'
  allOf:
    - required: [event]
    - oneOf:
        - required: [mLFileAddr]
        - required: [mLModelAdrf]

MLModelAdrf:
  type: object
  properties:
    adrfId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfInstanceId'
    adrfSetId:
      $ref: 'TS29571_CommonData.yaml#/components/schemas/NfSetId'
    storTransId:
      type: string
  oneOf:
    - required: [adrfId]
    - required: [adrfSetId]
```

`storTransId` is optional in `MLModelAdrf`. This is why the deployment retains
the ADRF identity and `modelUniqueId` as authoritative retrieval inputs and
does not make successful reprovision dependent on always recovering the
storage transaction identifier.

## 6. Supported and extension boundaries

| Behavior | Classification | Used by NWDAF/PyMTLF/PyAnLF |
| --- | --- | --- |
| `POST /data-store-records` | Release 18 standard | Yes |
| `GET /data-store-records?fetch-correlation-ids=...` | Release 18 standard | Yes |
| `GET /data-snapshots/{subscriptionId}/download` returning a JSON array | Repository extension | No |
| `POST /mlmodel-store-records` | Release 18 standard | Yes |
| Collection GET with `store-trans-id` or `model-unique-ids` | Release 18 standard | Yes |
| `GET /mlmodel-store-records/{storeTransId}` | Repository extension in the Release 18 profile | No |
| `GET /mlmodel-store-records/{storeTransId}/model` | Artifact-serving implementation detail | Yes, only through `mlFileAddr` |

Future changes must preserve this classification. Adding an extension is not
equivalent to extending the standardized OpenAPI contract, and standard
consumers must not be made dependent on an undocumented extension.
