# 5 API Definitions

## 5.1 Nadrf_DataManagement Service API

### 5.1.1 Introduction

The Nadrf_DataManagement service shall use the Nadrf_DataManagement API.

The API URI of the Nadrf_DataManagement API shall be:

`{apiRoot}/<apiName>/<apiVersion>`

The request URIs used in HTTP requests from the NF service consumer towards the NF service producer shall have the Resource URI structure defined in clause 4.4.1 of 3GPP TS 29.501 [5], i.e.:

`{apiRoot}/<apiName>/<apiVersion>/<apiSpecificResourceUriPart>`

with the following components:

* The `{apiRoot}` shall be set as described in 3GPP TS 29.501 [5].
* The `<apiName>` shall be "nadrf-datamanagement".
* The `<apiVersion>` shall be "v1".
* The `<apiSpecificResourceUriPart>` shall be set as described in clause 5.1.3.

### 5.1.2 Usage of HTTP

#### 5.1.2.1 General

HTTP/2, IETF RFC 9113 [11], shall be used as specified in clause 5 of 3GPP TS 29.500 [4].

HTTP/2 shall be transported as specified in clause 5.3 of 3GPP TS 29.500 [4].

The OpenAPI [6] specification of HTTP messages and content bodies for the Nadrf_DataManagement API is contained in Annex A.

#### 5.1.2.2 HTTP standard headers

##### 5.1.2.2.1 General

See clause 5.2.2 of 3GPP TS 29.500 [4] for the usage of HTTP standard headers.

##### 5.1.2.2.2 Content type

JSON, IETF RFC 8259 [12], shall be used as content type of the HTTP bodies specified in the present specification as specified in clause 5.4 of 3GPP TS 29.500 [4].

The use of the JSON format shall be signalled by the content type "application/json".

"Problem Details" JSON object shall be used to indicate additional details of the error in a HTTP response body and shall be signalled by the content type "application/problem+json", as defined in IETF RFC 9457 [13].

#### 5.1.2.3 HTTP custom headers

The mandatory HTTP custom header fields specified in clause 5.2.3.2 of 3GPP TS 29.500 [4] shall be supported, and the optional HTTP custom header fields specified in clause 5.2.3.3 of 3GPP TS 29.500 [4] may be supported.

### 5.1.3 Resources

#### 5.1.3.1 Overview

This clause describes the structure for the Resource URIs, the resources and methods used for the service.

Figure 5.1.3.1-1 depicts the resource URIs structure for the Nadrf_DataManagement API.

**Figure 5.1.3.1-1: Resource URI structure of the Nadrf_DataManagement API**

Table 5.1.3.1-1 provides an overview of the resources and applicable HTTP methods.

**Table 5.1.3.1-1: Resources and methods overview**

| Resource name | Resource URI | HTTP method or custom operation | Description |
| :--- | :--- | :--- | :--- |
| ADRF Data Store Records | `/data-store-records` | GET | Retrieve the stored data or analytics |
| | | POST | Create a new Individual Data Store Record resource. |
| Individual ADRF Data Store Record | `/data-store-records/{storeTransId}` | DELETE | Delete an individual ADRF Data Store Record identified by `{storeTransId}`. |
| ADRF Data Retrieval Subscriptions | `/data-retrieval-subscriptions` | POST | Create a new Individual ADRF Data Retrieval Subscription resource. |
| Individual ADRF Data Retrieval Subscription | `/data-retrieval-subscriptions/{subscriptionId}` | DELETE | Delete an individual ADRF Data Retrieval Subscription identified by `{subscriptionId}`. |

#### 5.1.3.2 Resource: ADRF Data Store Records

##### 5.1.3.2.1 Description

The ADRF Data Store Records resource represents all data storage records to the Nadrf_DataManagement Service at a given ADRF.

The resource allows an NF service consumer to create a new Individual ADRF Data Store Record resource and to retrieve Individual ADRF Data Store Record resources that fulfil certain criteria.

##### 5.1.3.2.2 Resource Definition

Resource URI: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records`

The `<apiVersion>` shall be set as described in clause 5.1.1.

This resource shall support the resource URI variables defined in table 5.1.3.2.2-1.

**Table 5.1.3.2.2-1: Resource URI variables for this resource**

| Name | Data type | Definition |
| :--- | :--- | :--- |
| `apiRoot` | string | See clause 5.1.1 |

##### 5.1.3.2.3 Resource Standard Methods

###### 5.1.3.2.3.1 POST

This method shall support the URI query parameters specified in table 5.1.3.2.3.1-1.

**Table 5.1.3.2.3.1-1: URI query parameters supported by the POST method on this resource**

| Name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| n/a | | | | | |

This method shall support the request data structures specified in table 5.1.3.2.3.1-2 and the response data structures and response codes specified in table 5.1.3.2.3.1-3.

**Table 5.1.3.2.3.1-2: Data structures supported by the POST Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfDataStoreRecord` | M | 1 | New individual Data Store Record to be created |

**Table 5.1.3.2.3.1-3: Data structures supported by the POST Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NadrfDataStoreRecord` | M | 1 | 201 Created | The creation of an Individual Data Store Record resource is confirmed, and a representation of that resource is returned. |

> **NOTE:** The mandatory HTTP error status code for the HTTP POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.

**Table 5.1.3.2.3.1-4: Headers supported by the 201 response code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains the URI of the newly created resource, according to the structure: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId}` |

###### 5.1.3.2.3.2 GET

This method shall support the URI query parameters specified in table 5.1.3.2.3.2-1.

**Table 5.1.3.2.3.2-1: URI query parameters supported by the GET method on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `store-trans-id` | string | O | 0..1 | Identifies the "Storage Transaction Identifier" of data store record in ADRF. (NOTE) |
| `fetch-correlation-ids` | array(string) | O | 1..N | Identifies fetch correlation identifiers received as part of fetch instruction. (NOTE) |
| `data-set-id` | string | O | 0..1 | Identifies a data set, i.e. the data or analytics records that contain the same value in the data set identifier of their data set tag. It may only be provided if the "EnhDataMgmt" feature is supported. |

> **NOTE:** Exactly one of "store-trans-id", "fetch-correlation-ids", and "data-set-id" shall be provided.

This method shall support the request data structures specified in table 5.1.3.2.3.1-2 and the response data structures and response codes specified in table 5.1.3.2.3.1-3.

**Table 5.1.3.2.3.1-2: Data structures supported by the GET Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| n/a | | | |

**Table 5.1.3.2.3.1-3: Data structures supported by the GET Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NadrfDataStoreRecord` | M | 1 | 200 OK | Data Store record. |
| n/a | | | 204 No Content | If the request ADRF Data Store Record does not exist, the ADRF shall respond with "204 No Content". |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status codes for the HTTP GET method listed in table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.3.2.3.1-4: Headers supported by the 307 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

**Table 5.1.3.2.3.1-5: Headers supported by the 308 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

##### 5.1.3.2.4 Resource Custom Operations

None.

#### 5.1.3.3 Resource: Individual ADRF Data Store Record

##### 5.1.3.3.1 Description

The Individual ADRF Data Store Record resource represents data or analytics stored via the Nadrf_DataManagement_StorageRequest in ADRF.

##### 5.1.3.3.2 Resource Definition

Resource URI: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId}`

The `<apiVersion>` shall be set as described in clause 5.1.1.

This resource shall support the resource URI variables defined in table 5.1.3.3.2-1.

**Table 5.1.3.3.2-1: Resource URI variables for this resource**

| Name | Data type | Definition |
| :--- | :--- | :--- |
| `apiRoot` | string | See clause 5.1.1. |
| `storeTransId` | string | Identifies an individual data store record. |

##### 5.1.3.3.3 Resource Standard Methods

###### 5.1.3.3.3.1 DELETE

This method shall support the URI query parameters specified in table 5.1.3.3.3.1-1.

**Table 5.1.3.3.3.1-1: URI query parameters supported by the DELETE method on this resource**

| Name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| n/a | | | | | |

This method shall support the request data structures specified in table 5.1.3.3.3.1-2 and the response data structures and response codes specified in table 5.1.3.3.3.1-3.

**Table 5.1.3.3.3.1-2: Data structures supported by the DELETE Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| n/a | | | |

**Table 5.1.3.3.3.1-3: Data structures supported by the DELETE Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| n/a | | | 204 No Content | The Individual ADRF Data Store Record resource was deleted successfully. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection, during Individual ADRF Data Store Record deletion. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection, during Individual ADRF Data Store Record deletion. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status code for the HTTP DELETE method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.3.3.3.1-4: Headers supported by the 307 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target ADRF (service) instance towards which the request is redirected. |

**Table 5.1.3.3.3.1-5: Headers supported by the 308 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target ADRF (service) instance towards which the request is redirected. |

##### 5.1.3.3.4 Resource Custom Operations

None in this release of the specification.

#### 5.1.3.4 Resource: ADRF Data Retrieval Subscriptions

##### 5.1.3.4.1 Description

The ADRF Data Retrieval Subscriptions resource represents all data retrieval subscriptions to the Nadrf_DataManagement Service at a given ADRF.

The resource allows an NF service consumer to create a new Individual ADRF Data Retrieval Subscription resource.

##### 5.1.3.4.2 Resource Definition

Resource URI: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions`

The `<apiVersion>` shall be set as described in clause 5.1.1.

This resource shall support the resource URI variables defined in table 5.1.3.4.2-1.

**Table 5.1.3.4.2-1: Resource URI variables for this resource**

| Name | Data type | Definition |
| :--- | :--- | :--- |
| `apiRoot` | string | See clause 5.1.1 |

##### 5.1.3.4.3 Resource Standard Methods

###### 5.1.3.4.3.1 POST

This method shall support the URI query parameters specified in table 5.1.3.4.3.1-1.

**Table 5.1.3.4.3.1-1: URI query parameters supported by the POST method on this resource**

| Name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| n/a | | | | | |

This method shall support the request data structures specified in table 5.1.3.4.3.1-2 and the response data structures and response codes specified in table 5.1.3.4.3.1-3.

**Table 5.1.3.4.3.1-2: Data structures supported by the POST Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfDataRetrievalSubscription` | M | 1 | Individual ADRF Data Retrieval Subscription resource to be created. |

**Table 5.1.3.4.3.1-3: Data structures supported by the POST Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NadrfDataRetrievalSubscription` | M | 1 | 201 Created | The creation of an Individual ADRF Data Retrieval Subscription resource is confirmed and a representation of that resource is returned. |

> **NOTE:** The mandatory HTTP error status code for the HTTP POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.

**Table 5.1.3.4.3.1-4: Headers supported by the 201 response code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains the URI of the newly created resource, according to the structure: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId}` |

##### 5.1.3.4.4 Resource Custom Operations

None in this release of the specification.

#### 5.1.3.5 Resource: Individual ADRF Data Retrieval Subscription

##### 5.1.3.5.1 Description

The Individual ADRF Data Retrieval Subscription resource represents single ADRF data retrieval subscription to the Nadrf_DataManagement Service at a given ADRF.

The resource allows an NF service consumer to delete Individual ADRF Data Retrieval Subscription resource.

##### 5.1.3.5.2 Resource Definition

Resource URI: `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId}`

This resource shall support the resource URI variables defined in table 5.1.3.5.2-1.

**Table 5.1.3.5.2-1: Resource URI variables for this resource**

| Name | Data type | Definition |
| :--- | :--- | :--- |
| `apiRoot` | string | See clause 5.1.1 |
| `subscriptionId` | string | Identifies a subscription to the Nadrf_DataManagement service. |

##### 5.1.3.5.3 Resource Standard Methods

###### 5.1.3.5.3.1 DELETE

This method shall support the URI query parameters specified in table 5.1.3.5.3.1-1.

**Table 5.1.3.5.3.1-1: URI query parameters supported by the DELETE method on this resource**

| Name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| n/a | | | | | |

This method shall support the request data structures specified in table 5.1.3.5.3.1-2 and the response data structures and response codes specified in table 5.1.3.5.3.1-3.

**Table 5.1.3.5.3.1-2: Data structures supported by the DELETE Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| n/a | | | |

**Table 5.1.3.5.3.1-3: Data structures supported by the DELETE Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| n/a | | | 204 No Content | The Individual ADRF Data Retrieval Subscription resource was deleted successfully. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection, during Individual ADRF Data Retrieval Subscription deletion. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection, during Individual ADRF Data Retrieval Subscription deletion. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status code for the HTTP DELETE method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.3.5.3.1-4: Headers supported by the 307 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target ADRF (service) instance towards which the request is redirected. |

**Table 5.1.3.5.3.1-5: Headers supported by the 308 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative ADRF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target ADRF (service) instance towards which the request is redirected. |

##### 5.1.3.5.4 Resource Custom Operations

None in this release of the specification.

### 5.1.4 Custom Operations without associated resources

#### 5.1.4.1 Overview

The structure of the custom operation URIs of the Nadrf_DataManagement service is shown in Figure 5.1.4.1-1.

**Figure 5.1.4.1-1: Custom operation URI structure of the Nadrf_DataManagement API**

Table 5.1.4.1-1 provides an overview of the custom operations and applicable HTTP methods.

**Table 5.1.4.1-1: Custom operations without associated resources**

| Custom operation URI | Mapped HTTP method | Description |
| :--- | :--- | :--- |
| `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub` | POST | Request the ADRF to create a subscription for data or analytics and then store the received data or analytics in the ADRF. |
| `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub-removal` | POST | Request the ADRF to remove a subscription for data or analytics. |
| `{apiRoot}/nadrf-datamanagement/<apiVersion>/remove-stored-data-analytics` | POST | Request the ADRF to remove already stored data or analytics. |

#### 5.1.4.2 Operation: request-storage-sub

##### 5.1.4.2.1 Description

The operation is used by the NF service consumer to request the ADRF to create a subscription for data or analytics and then store the received data or analytics in the ADRF.

##### 5.1.4.2.2 Operation Definition

This operation shall support the request data structures shown in Table 5.1.4.2.2-1 and the response data structures and error codes specified in Tables 5.1.4.2.2-2.

**Table 5.1.4.2.2-1: Data structures supported by the POST Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfDataStoreSubscription` | M | 1 | Information about the storage subscription that the ADRF shall create. |

**Table 5.1.4.2.2-2: Data structures supported by the POST Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NadrfDataStoreSubscriptionRef` | M | 1 | 200 OK | Successful request to trigger the creation of a subscription for data or analytics at the ADRF. A reference is provided. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status code for the HTTP POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.4.2.2-3: Headers supported by the 307 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

**Table 5.1.4.2.2-4: Headers supported by the 308 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

#### 5.1.4.3 Operation: request-storage-sub-removal

##### 5.1.4.3.1 Description

The operation is used by the NF service consumer to request the ADRF to remove a subscription for data or analytics which was used to store the received data or analytics in the ADRF.

##### 5.1.4.3.2 Operation Definition

This operation shall support the request data structures shown in Table 5.1.4.3.2-1 and the response data structures and error codes specified in Tables 5.1.4.3.2-2.

**Table 5.1.4.3.2-1: Data structures supported by the POST Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfDataStoreSubscriptionRef` | M | 1 | Reference used to identify the subscription that the ADRF shall remove. |

**Table 5.1.4.3.2-2: Data structures supported by the POST Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| n/a | | | 204 No Content | Successful request to trigger the removal of a subscription for data or analytics at the ADRF. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status code for the HTTP POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.4.3.2-3: Headers supported by the 307 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

**Table 5.1.4.3.2-4: Headers supported by the 308 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

#### 5.1.4.4 Operation: remove-stored-data-analytics

##### 5.1.4.4.1 Description

The operation is used by the NF service consumer to request the ADRF to remove stored data or analytics based on a data or analytics specification.

##### 5.1.4.4.2 Operation Definition

This operation shall support the request data structures shown in Table 5.1.4.4.2-1 and the response data structures and error codes specified in Tables 5.1.4.4.2-2.

**Table 5.1.4.4.2-1: Data structures supported by the POST Request Body on this resource**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfStoredDataSpec` | M | 1 | Information about the specification of data or analytics stored in ADRF. |

**Table 5.1.4.4.2-2: Data structures supported by the POST Response Body on this resource**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| n/a | | | 204 No Content | Successful request to remove data or analytics at the ADRF based on a data or analytics specification. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status code for the HTTP POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] shall also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.4.4.2-3: Headers supported by the 307 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

**Table 5.1.4.4.2-4: Headers supported by the 308 Response Code on this custom operation**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative target URI located in an alternative ADRF (service) instance towards which the request should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Contains the identifier of the target ADRF (service) instance towards which the request should be redirected. |

### 5.1.5 Notifications

#### 5.1.5.1 General

Notifications shall comply to clause 6.2 of 3GPP TS 29.500 [4] and clause 4.6.2.3 of 3GPP TS 29.501 [5].

**Table 5.1.5.1-1: Notifications overview**

| Notification | Callback URI | HTTP method or custom operation | Description (service operation) |
| :--- | :--- | :--- | :--- |
| Retrieval Notification | `{notificationURI}` | POST | Report data or analytics from ADRF. |
| ADRF Alert Notification | `{delNotifUri}` | POST | Notify about data or analytics that are about to be deleted. |

#### 5.1.5.2 Retrieval Notification

##### 5.1.5.2.1 Description

The Retrieval Notification is used by the NF service producer to report one or several data retrieval events to an NF service consumer that has subscribed to such notifications.

##### 5.1.5.2.2 Target URI

The Callback URI "{notificationURI}" shall be used with the callback URI variables defined in table 5.1.5.2.2-1.

**Table 5.1.5.2.2-1: Callback URI variables**

| Name | Definition |
| :--- | :--- |
| `notificationURI` | String formatted as URI with the Callback Uri |

##### 5.1.5.2.3 Standard Methods

###### 5.1.5.2.3.1 POST

This method shall support the request data structures specified in table 5.1.5.2.3.1-1 and the response data structures and response codes specified in table 5.1.5.2.3.1-2.

**Table 5.1.5.2.3.1-1: Data structures supported by the POST Request Body**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfDataRetrievalNotification` | M | 1 | Provides information about observed data or analytics. |

**Table 5.1.5.2.3.1-2: Data structures supported by the POST Response Body**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| n/a | | | 204 No Content | The receipt of the Notification is acknowledged. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection, during the retrieval notification. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection, during the retrieval notification. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status codes for the POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.5.2.3.1-3: Headers supported by the 307 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI representing the end point of an alternative NF consumer (service) instance towards which the notification should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target NF service consumer (service) instance towards which the notification request is redirected. |

**Table 5.1.5.2.3.1-4: Headers supported by the 308 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI representing the end point of an alternative NF consumer (service) instance towards which the notification should be redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target NF service consumer (service) instance towards which the notification request is redirected. |

#### 5.1.5.3 ADRF Alert Notification

##### 5.1.5.3.1 Description

The ADRF Alert Notification is used by the NF service producer to notify an NF service consumer about data or analytics that are about to be deleted.

##### 5.1.5.3.2 Target URI

The Callback URI "{delNotifUri}" shall be used with the callback URI variables defined in table 5.1.5.3.2-1.

**Table 5.1.5.3.2-1: Callback URI variables**

| Name | Definition |
| :--- | :--- |
| `delNotifUri` | String formatted as URI with the Callback Uri |

##### 5.1.5.3.3 Standard Methods

###### 5.1.5.3.3.1 POST

This method shall support the request data structures specified in table 5.1.5.3.3.1-1 and the response data structures and response codes specified in table 5.1.5.3.3.1-2.

**Table 5.1.5.3.3.1-1: Data structures supported by the POST Request Body**

| Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- |
| `NadrfAlertNotification` | M | 1 | Contains information about data or analytics that are about to be deleted. |

**Table 5.1.5.3.3.1-2: Data structures supported by the POST Response Body**

| Data type | P | Cardinality | Response codes | Description |
| :--- | :--- | :--- | :--- | :--- |
| `NadrfAlertNotificationResponse` | M | 1 | 200 OK | The receipt of the Notification is acknowledged and a response with information about the planned action is provided. |
| n/a | | | 204 No Content | The receipt of the Notification is acknowledged. |
| `RedirectResponse` | O | 0..1 | 307 Temporary Redirect | Temporary redirection. (NOTE 2) |
| `RedirectResponse` | O | 0..1 | 308 Permanent Redirect | Permanent redirection. (NOTE 2) |

> **NOTE 1:** The mandatory HTTP error status codes for the POST method listed in Table 5.2.7.1-1 of 3GPP TS 29.500 [4] also apply.
> **NOTE 2:** The `RedirectResponse` data structure may be provided by an SCP (cf. clause 6.10.9.1 of 3GPP TS 29.500 [4]).

**Table 5.1.5.3.3.1-3: Headers supported by the 307 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative NF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target NF (service) instance towards which the notification request is redirected. |

**Table 5.1.5.3.3.1-4: Headers supported by the 308 Response Code on this resource**

| Name | Data type | P | Cardinality | Description |
| :--- | :--- | :--- | :--- | :--- |
| `Location` | string | M | 1 | Contains an alternative URI of the resource located in an alternative NF (service) instance towards which the request is redirected. For the case where the request is redirected to the same target via a different SCP, refer to clause 6.10.9.1 of 3GPP TS 29.500 [4]. |
| `3gpp-Sbi-Target-Nf-Id` | string | O | 0..1 | Identifier of the target NF (service) instance towards which the notification request is redirected. |

### 5.1.6 Data Model

#### 5.1.6.1 General

This clause specifies the application data model supported by the Nadrf_DataManagement API.

Table 5.1.6.1-1 specifies the data types defined for the Nadrf_DataManagement service based interface protocol.

**Table 5.1.6.1-1: Nadrf_DataManagement specific Data Types**

| Data type | Clause defined | Description | Applicability |
| :--- | :--- | :--- | :--- |
| `DataNotification` | 5.1.6.2.9 | Represents a data subscription notification of one of various possible data sources. | |
| `DataSetTag` | 5.1.6.2.13 | Contains an identifier and a description of associated data or analytics records. | EnhDataMgmt |
| `DataSubscription` | 5.1.6.2.8 | Contains information about Data specification. | |
| `NadrfAlertNotification` | 5.1.6.2.11 | Contains information about data or analytics that are about to be deleted. | EnhDataMgmt |
| `NadrfAlertNotificationResponse` | 5.1.6.2.12 | Contains information about the planned action upon receiving an alert. | EnhDataMgmt |
| `NadrfDataRetrievalNotification` | 5.1.6.2.5 | Represents a notification that corresponds with an Individual ADRF Data Retrieval Subscription resource. | |
| `NadrfDataRetrievalSubscription` | 5.1.6.2.4 | Represents an Individual ADRF Data Retrieval Subscription resource. | |
| `NadrfDataStoreRecord` | 5.1.6.2.2 | Represents an Individual ADRF Data Store Record resource. | |
| `NadrfDataStoreSubscription` | 5.1.6.2.3 | Contains information to be used by the ADRF to create a Data or Analytics subscription. | |
| `NadrfDataStoreSubscriptionRef` | 5.1.6.2.6 | Contains a reference to a request for a Data or Analytics subscription. | |
| `NadrfStoredDataSpec` | 5.1.6.2.7 | Contains information about Data or Analytics specification. | |
| `StorageHandlingInfo` | 5.1.6.2.10 | Contains storage handling information for data or analytics. | EnhDataMgmt |

Table 5.1.6.1-2 specifies data types re-used by the Nadrf_DataManagement service based interface protocol from other specifications, including a reference to their respective specifications and when needed, a short description of their use within the Nadrf_DataManagement service based interface.

**Table 5.1.6.1-2: Nadrf_DataManagement re-used Data Types**

| Data type | Reference | Comments | Applicability |
| :--- | :--- | :--- | :--- |
| `AfEventExposureNotif` | 3GPP TS 29.517 [20] | Represents notifications on AF event(s) that occurred for an Individual AF Event Subscription resource. | |
| `AfEventExposureSubsc` | 3GPP TS 29.517 [20] | Represents AF event subscription. | |
| `AmfEventNotification` | 3GPP TS 29.518 [18] | Represents notifications on AMF event(s) that occurred for an Individual AMF Event Subscription resource. | |
| `AmfEventSubscription` | 3GPP TS 29.518 [18] | Represents AMF event subscription. | |
| `DurationSec` | 3GPP TS 29.571 [16] | Unsigned integer identifying a period of time in units of seconds. | EnhDataMgmt |
| `DateTime` | 3GPP TS 29.571 [16] | Identifies the time. | |
| `EeSubscription` | 3GPP TS 29.503 [19] | Represents UDM event subscription. | |
| `EventNotifyData` | 3GPP TS 29.515 [27] | Represents GMLC event notification. | LocEvents |
| `FetchInstruction` | 3GPP TS 29.576 [24] | The fetch instruction indicates that the data or analytics can be fetched by the consumer. | |
| `FormattingInstruction` | 3GPP TS 29.574 [23] | DCCF formatting Instructions. | |
| `InputData` | 3GPP TS 29.515 [27] | Represents GMLC event subscription. | LocEvents |
| `LmfDataExposureNotification` | 3GPP TS 29.572 [29] | Represents an LMF data exposure event notification. | LmfEvents |
| `LmfDataExposureSubscription` | 3GPP TS 29.572 [29] | Represents an LMF data exposure subscription. | LmfEvents |
| `MonitoringReport` | 3GPP TS 29.503 [19] | UDM Monitoring Report. | |
| `NefEventExposureNotif` | 3GPP TS 29.591 [21] | Represents notifications on network exposure event(s) that occurred for an Individual Network Exposure Event Subscription resource. | |
| `NefEventExposureSubsc` | 3GPP TS 29.591 [21] | Represents NEF event subscription. | |
| `NfInstanceId` | 3GPP TS 29.571 [16] | NF instance identifier. | |
| `NfSetId` | 3GPP TS 29.571 [16] | NF set identifier. | |
| `NnwdafEventsSubscription` | 3GPP TS 29.520 [15] | Represents an NWDAF analytics subscription. | |
| `NnwdafEventsSubscriptionNotification` | 3GPP TS 29.520 [15] | Represents an NWDAF analytics subscription notification. | |
| `NotificationData` (NRF) | 3GPP TS 29.510 [10] | Represents an NRF event notification. | |
| `NotificationData` (UPF) | 3GPP TS 29.564 [26] | Represents a UPF event notification. | UpEvents |
| `NsmfEventExposure` | 3GPP TS 29.508 [17] | Represents an SMF event subscription. | |
| `NsmfEventExposureNotification` | 3GPP TS 29.508 [17] | Represents an SMF event notification. | |
| `PcEventExposureNotif` | 3GPP TS 29.523 [30] | Represents a PCF event notification. | PcfEvents |
| `PcEventExposureSubsc` | 3GPP TS 29.523 [30] | Represents a PCF event subscription. | PcfEvents |
| `ProcessingInstruction` | 3GPP TS 29.574 [23] | DCCF processing Instructions. | |
| `SACEventReport` | 3GPP TS 29.536 [25] | Represents an NSACF event notification. | |
| `SACEventSubscription` | 3GPP TS 29.536 [25] | Represents an NSACF event subscription. | |
| `SubscriptionData` | 3GPP TS 29.510 [10] | Represents an NRF event subscription. | |
| `SupportedFeatures` | 3GPP TS 29.571 [16] | Used to negotiate the applicability of the optional features defined in table 5.1.8-1. | |
| `TimeWindow` | 3GPP TS 29.122 [22] | Represents a time window. | |
| `UpfEventSubscription` | 3GPP TS 29.564 [26] | Represents a UPF event subscription. | UpEvents |
| `Uri` | 3GPP TS 29.571 [16] | URI. | |

#### 5.1.6.2 Structured data types

##### 5.1.6.2.1 Introduction

This clause defines the structures to be used in resource representations.

##### 5.1.6.2.2 Type: NadrfDataStoreRecord

**Table 5.1.6.2.2-1: Definition of type NadrfDataStoreRecord**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `dataNotif` | `DataNotification` | C | 0..1 | Data subscription notification. (NOTE 1) (NOTE 4) | |
| `anaNotifications` | array(`NnwdafEventsSubscriptionNotification`) | C | 1..N | List of analytics subscription notifications. (NOTE 1) (NOTE 4) | |
| `anaSub` | array(`NnwdafEventsSubscription`) | C | 1..N | Represents the subscription information of the corresponding analytics notification. Shall be present if the "anaNotifications" attribute is provided. (NOTE 2) (NOTE 4) | |
| `dataSetTag` | `DataSetTag` | O | 0..1 | Data set tag of the stored data or analytics. | EnhDataMgmt |
| `dsc` | string | O | 0..1 | Data synthesis and compression information. (NOTE 3) | EnhDataMgmt |
| `dataSub` | array(`DataSubscription`) | C | 1..N | Represents the subscription information of the corresponding data notification. Shall be present if the "dataNotif" attribute is provided. (NOTE 2) (NOTE 4) | |
| `storeHandl` | `StorageHandlingInfo` | O | 0..1 | Contains storage handling information for the data or analytics that are being stored. | EnhDataMgmt |
| `suppFeat` | `SupportedFeatures` | C | 0..1 | Contains the list of supported features among the ones defined in clause 5.1.8. This attribute shall be provided only when feature negotiation needs to take place. | |

> **NOTE 1:** Exactly one of the attributes "anaNotifications" and "dataNotif" shall be provided.
> **NOTE 2:** Exactly one of the attributes "anaSub" and "dataSub" shall be provided.
> **NOTE 3:** The format and the contents of this attribute are up to the implementation.
> **NOTE 4:** When the "dataSub" attribute is provided, only one element should be included to map with the "dataNotif" attribute which shall also be provided. When the "anaSub" attribute is provided, the array elements of the "anaSub" attribute match in sequence with the array elements in the "anaNotifications" attribute which shall also be provided.

##### 5.1.6.2.3 Type: NadrfDataStoreSubscription

**Table 5.1.6.2.3-1: Definition of type NadrfDataStoreSubscription**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `anaSub` | `NnwdafEventsSubscription` | C | 0..1 | Subscribed analytics events. (NOTE 1) | |
| `dataSetTag` | `DataSetTag` | O | 0..1 | Data set tag of the stored data or analytics. | EnhDataMgmt |
| `dataSub` | `DataSubscription` | C | 0..1 | Represents requested Events subscription. (NOTE 1) | |
| `targetNfId` | `NfInstanceId` | C | 0..1 | DCCF or NWDAF NF instance identifier to which the ADRF shall create the requested subscription. (NOTE 2) | |
| `targetNfSetId` | `NfSetId` | C | 0..1 | DCCF or NWDAF NF set identifier to which the ADRF shall create the requested subscription. (NOTE 2) | |
| `formatInstruct` | `FormattingInstruction` | O | 0..1 | Formatting instructions to be used for sending event notifications. | |
| `procInstruct` | `ProcessingInstruction` | O | 0..1 | Processing instructions to be used for sending event notifications. (NOTE 3) | |
| `multiProcInstructs` | array(`ProcessingInstruction`) | O | 1..N | Processing instructions to be used for sending event notifications. (NOTE 3) | MultiProcessingInstruction |
| `storeHandl` | `StorageHandlingInfo` | O | 0..1 | Contains storage handling information for the data or analytics that will be collected and stored based on the requested subscription. | EnhDataMgmt |
| `suppFeat` | `SupportedFeatures` | C | 0..1 | Contains the list of supported features among the ones defined in clause 5.1.8. This attribute shall be provided only when feature negotiation needs to take place. | |

> **NOTE 1:** Exactly one of these attributes shall be provided.
> **NOTE 2:** One of "targetNfId" and "targetNfSetId" shall be provided.
> **NOTE 3:** The "multiProcInstructs" attribute shall be used instead of the "procInstruct" attribute when the "MultiProcessingInstruction" feature is supported.

##### 5.1.6.2.4 Type: NadrfDataRetrievalSubscription

**Table 5.1.6.2.4-1: Definition of type NadrfDataRetrievalSubscription**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `anaSub` | `NnwdafEventsSubscription` | C | 0..1 | Subscribed analytics events. (NOTE 1) | |
| `dataSetId` | string | C | 0..1 | Data set identifier of stored data or analytics records. (NOTE 1) | EnhDataMgmt |
| `dataSub` | `DataSubscription` | C | 0..1 | Represents requested Events subscription. (NOTE 1) | |
| `notificationURI` | `Uri` | M | 1 | Notification target address. | |
| `timePeriod` | `TimeWindow` | M | 1 | Represents a start time and a stop time during which the requested data is collected and/or will be collected at the data source. | |
| `notifCorrId` | string | M | 1 | Notification correlation identifier provided by the NF service consumer to be used later by the ADRF in the notifications that correspond with this subscription. The value of this attribute shall be unique per subscription for a given NF service consumer. | |
| `consTrigNotif` | boolean | O | 0..1 | If provided and set to "true", it indicates that notifications shall be buffered (sending only fetch instructions to the NF service consumer) until the NF service consumer requests their delivery using Nadrf_DataManagement Service. The default value is "false". | |
| `suppFeat` | `SupportedFeatures` | C | 0..1 | Contains the list of supported features among the ones defined in clause 5.1.8. This attribute shall be provided only when feature negotiation needs to take place. | |

> **NOTE 1:** Exactly one of these attributes shall be provided.

##### 5.1.6.2.5 Type: NadrfDataRetrievalNotification

**Table 5.1.6.2.5-1: Definition of type NadrfDataRetrievalNotification**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `notifCorrId` | string | M | 1 | This attribute indicates the notification correlation identifier provided by the NF service consumer during the data retrieval subscription. This parameter can be useful if the NF service consumer uses a common callback URI for multiple subscriptions. | |
| `anaNotifications` | array(`NnwdafEventsSubscriptionNotification`) | C | 1..N | List of analytics subscription notifications. (NOTE 1) | |
| `dataNotif` | `DataNotification` | C | 0..1 | Data subscription notification. (NOTE 1) | |
| `fetchInstruct` | `FetchInstruction` | C | 0..1 | The fetch instruction indicates that the data or analytics can be fetched by the consumer. (NOTE 1) | |
| `terminationReq` | boolean | O | 0..1 | If provided and set to "true", it indicates that the subscription is requested to be terminated, i.e. no further notifications related to this subscription will be provided. The default value is "false". | |
| `dsc` | string | O | 0..1 | Data synthesis and compression information. (NOTE 2) | EnhDataMgmt |
| `timeStamp` | `DateTime` | M | 1 | It represents the time when ADRF completed preparation of the requested data or analytics. | |

> **NOTE 1:** Exactly one of these attributes shall be provided.
> **NOTE 2:** The format and the contents of this attribute are up to the implementation.

##### 5.1.6.2.6 Type: NadrfDataStoreSubscriptionRef

**Table 5.1.6.2.6-1: Definition of type NadrfDataStoreSubscriptionRef**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `transRefId` | string | C | 0..1 | Transaction reference identifier. (NOTE) | |
| `dataSetId` | string | C | 0..1 | Data set identifier associated with a storage subscription. (NOTE) | EnhDataMgmt |
| `suppFeat` | `SupportedFeatures` | C | 0..1 | Contains the list of supported features among the ones defined in clause 5.1.8. This attribute shall be provided only when feature negotiation needs to take place. | |

> **NOTE:** One of the "transRefId" and "dataSetId" attributes shall be provided.

##### 5.1.6.2.7 Type: NadrfStoredDataSpec

**Table 5.1.6.2.7-1: Definition of type NadrfStoredDataSpec**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `dataSpec` | `DataSubscription` | C | 0..1 | Represents data specification. (NOTE) | |
| `anaSpec` | `NnwdafEventsSubscription` | C | 0..1 | Represents analytics specification. (NOTE) | |
| `timePeriod` | `TimeWindow` | M | 1 | Represents a start time and a stop time during which the requested data to be removed was collected at the data source. | |
| `dataSetId` | string | C | 0..1 | Data set identifier of stored data or analytics records. (NOTE) | EnhDataMgmt |

> **NOTE:** Exactly one of these attributes shall be provided.

##### 5.1.6.2.8 Type: DataSubscription

**Table 5.1.6.2.8-1: Definition of type DataSubscription**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `amfDataSub` | `AmfEventSubscription` | C | 0..1 | Represents requested AMF Events subscription. (NOTE) | |
| `smfDataSub` | `NsmfEventExposure` | C | 0..1 | Represents requested SMF Events subscription. (NOTE) | |
| `udmDataSub` | `EeSubscription` | C | 0..1 | Represents requested UDM Events subscription. (NOTE) | |
| `nefDataSub` | `NefEventExposureSubsc` | C | 0..1 | Represents requested NEF Events subscription. (NOTE) | |
| `afDataSub` | `AfEventExposureSubsc` | C | 0..1 | Represents requested AF Events subscription. (NOTE) | |
| `nrfDataSub` | `SubscriptionData` | C | 0..1 | Represents requested NRF Events subscription. (NOTE) | |
| `nsacfDataSub` | `SACEventSubscription` | C | 0..1 | Represents requested NSACF Events subscription. (NOTE) | |
| `upfDataSub` | `UpfEventSubscription` | C | 0..1 | Represents requested UPF Events subscription. (NOTE) | UpEvents |
| `gmlcDataSub` | `InputData` | C | 0..1 | Represents requested GMLC Events subscription. (NOTE) | LocEvents |
| `lmfDataSub` | `LmfDataExposureSubscription` | C | 0..1 | Represents requested LMF Events subscription. (NOTE) | LmfEvents |
| `pcfDataSub` | `PcEventExposureSubsc` | C | 0..1 | Represents requested PCF Events subscription. (NOTE) | PcfEvents |

> **NOTE:** Exactly one of these attributes shall be provided.

##### 5.1.6.2.9 Type: DataNotification

**Table 5.1.6.2.9-1: Definition of type DataNotification**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `afEventNotifs` | array(`AfEventExposureNotif`) | C | 1..N | List of notifications on AF event(s). (NOTE 1) | |
| `amfEventNotifs` | array(`AmfEventNotification`) | C | 1..N | List of notifications on AMF event(s). (NOTE 1) | |
| `gmlcEventNotifs` | array(`EventNotifyData`) | C | 1..N | List of notifications on GMLC event(s). (NOTE 1) | LocEvents |
| `smfEventNotifs` | array(`NsmfEventExposureNotification`) | C | 1..N | List of notifications on SMF event(s). (NOTE 1) | |
| `udmEventNotifs` | array(`MonitoringReport`) | C | 1..N | List of monitoring reports containing information about UDM event(s). (NOTE 1) | |
| `nefEventNotifs` | array(`NefEventExposureNotif`) | C | 1..N | List of notifications on network exposure event(s). (NOTE 1) | |
| `nrfEventNotifs` | array(`NotificationData`) | C | 1..N | List of notifications on NRF event(s) as defined in 3GPP TS 29.510 [10]. (NOTE 1) | |
| `nsacfEventNotifs` | array(`SACEventReport`) | C | 1..N | List of notifications on NSACF event(s). (NOTE 1) | |
| `upfEventNotifs` | array(`NotificationData`) | C | 1..N | List of notifications on UPF event(s) as defined in 3GPP TS 29.564 [26]. (NOTE 1) | UpEvents |
| `lmfEventNotifs` | array(`LmfDataExposureNotification`) | C | 1..N | List of notifications on LMF event(s). (NOTE 1) | LmfEvents |
| `pcfEventNotifs` | array(`PcEventExposureNotif`) | C | 1..N | List of notifications on PCF event(s). (NOTE 1) | PcfEvents |
| `timeStamp` | `DateTime` | O | 0..1 | Indicates the timestamp for the event(s). (NOTE 2) | |

> **NOTE 1:** Exactly one of these attributes shall be provided.
> **NOTE 2:** The "timeStamp" attribute within the DataNotification data type may be provided in the case of NRF events or if any of the "timeStamp" attribute within AfEventNotification contained in the AfEventExposureNotif, or within AmfEventReport contained in the AmfEventNotification, or within EventNotification contained in the NsmfEventExposureNotification, or within MonitoringReport, or within NefEventNotification contained in the NefEventExposureNotif, or within SACEventReportItem contained in the SACEventReport, or, if the UpEvents feature is supported, within the NotificationItem contained in the NotificationData data type, or, if the PcfEvents feature is supported, within the PcEventNotification contained in the PcEventExposureNotif data type, or, if the LocEvents feature is supported, the "timestampOfLocationEstimate" attribute within EventNotifyData, or, if the LmfEvents feature is supported, the "timeStamp" attribute within the LmfDataExposureNotification, is not provided.

##### 5.1.6.2.10 Type: StorageHandlingInfo

**Table 5.1.6.2.10-1: Definition of type StorageHandlingInfo**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `lifetime` | `DurationSec` | O | 0..1 | Indicates the lifetime of the provided data or analytics as a duration in seconds. | |
| `delNotifUri` | `Uri` | O | 0..1 | Notification Uri for receiving deletion alerts. Its presence indicates that the NF service consumer wants to receive such notifications. | |
| `delNotifCorrId` | string | C | 0..1 | Notification correlation identifier for the deletion alerts. It shall be provided if the "delNotifUri" attribute is provided. | |

##### 5.1.6.2.11 Type: NadrfAlertNotification

**Table 5.1.6.2.11-1: Definition of type NadrfAlertNotification**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `alertStorTransId` | string | M | 1 | Storage transaction identifier that can be used to retrieve the data or analytics that are about to be deleted. | |
| `delNotifCorrId` | string | M | 1 | Notification correlation identifier for the alert. | |

##### 5.1.6.2.12 Type: NadrfAlertNotificationResponse

**Table 5.1.6.2.12-1: Definition of type NadrfAlertNotificationResponse**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `retrievalInd` | boolean | M | 1 | If the NF service consumer had received a Data or Analytics Deletion Alert in the notification and determined to retrieve stored data or analytics prior to deletion, it shall be set to "true". Otherwise, it shall be set to "false". | |

##### 5.1.6.2.13 Type: DataSetTag

**Table 5.1.6.2.13-1: Definition of type DataSetTag**

| Attribute name | Data type | P | Cardinality | Description | Applicability |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `dataSetId` | string | M | 1 | Data set identifier of data or analytics records. | |
| `dataSetDesc` | string | O | 0..1 | Human-readable data set description of data or analytics records. | |

#### 5.1.6.3 Simple data types and enumerations

None.

#### 5.1.6.4 Data types describing alternative data types or combinations of data types

None.