# 5 API Definitions 

## 5.1 Nadrf_DataManagement Service API 

### 5.1.7 Error Handling 

#### 5.1.7.1 General 

For the `Nadrf_DataManagement` API, HTTP error responses shall be supported as specified in clause 4.8 of 3GPP TS 29.501 [5]. Protocol errors and application errors specified in table 5.2.7.2-1 of 3GPP TS 29.500 [4] shall be supported for an HTTP method if the corresponding HTTP status codes are specified as mandatory for that HTTP method in table 5.2.7.1-1 of 3GPP TS 29.500 [4]. In addition, the requirements in the following clauses are applicable for the `Nadrf_DataManagement` API.

#### 5.1.7.2 Protocol Errors 

No specific procedures for the `Nadrf_DataManagement` service are specified.

#### 5.1.7.3 Application Errors 

The application errors defined for the `Nadrf_DataManagement` service are listed in Table 5.1.7.3-1.

**Table 5.1.7.3-1: Application errors** 

| Application Error | HTTP status code | Description |
| :--- | :--- | :--- |
| | | | 

### 5.1.8 Feature negotiation 

The optional features in table 5.1.8-1 are defined for the `Nadrf_DataManagement` API. They shall be negotiated using the extensibility mechanism defined in clause 6.6 of 3GPP TS 29.500 [4].

**Table 5.1.8-1: Supported Features** 

| Feature number | Feature Name | Description |
| :--- | :--- | :--- |
| 1 | `MultiProcessingInstruction` | Indicates the support of multiple processing instructions.  |
| 2 | `UpEvents` | Indicates the support of UPF events.  |
| 3 | `EnhDataMgmt` | Indicates the support of enhanced data management mechanisms.  |
| 4 | `LocEvents` | This feature indicates the support of location events.  |
| 5 | `LmfEvents` | This feature indicates the support of LMF data exposure events.  |
| 6 | `PcfEvents` | This feature indicates the support of PCF event exposure events.  |

### 5.1.9 Security 

As indicated in 3GPP TS 33.501 [8] and 3GPP TS 29.500 [4], the access to the `Nadrf_DataManagement` API may be authorized by means of the OAuth2 protocol (see IETF RFC 6749 [9]), based on local configuration, using the "Client Credentials" authorization grant, where the NRF (see 3GPP TS 29.510 [10]) plays the role of the authorization server. If OAuth2 is used, an NF Service Consumer, prior to consuming services offered by the `Nadrf_DataManagement` API, shall obtain a "token" from the authorization server, by invoking the Access Token Request service, as described in 3GPP TS 29.510 [10], clause 5.4.2.2.

> **NOTE:** When multiple NRFs are deployed in a network, the NRF used as authorization server is the same NRF that the NF Service Consumer used for discovering the `Nadrf_DataManagement` service. 

The `Nadrf_DataManagement` API defines the following scopes for OAuth2 authorization as described in 3GPP TS 29.501 [5], clause 4.10.

**Table 5.1.9-1: OAuth2 scopes defined in Nadrf_DataManagement API** 

| Scope | Description |
| :--- | :--- |
| `nadrf-datamanagement` | Access to the `Nadrf_DataManagement` API  |
| `nadrf-datamanagement:storage-read-delete-subs` | Access to service operations applying to `Nadrf_DataManagement_StorageSubscriptionRequest`, `Nadrf_DataManagement_StorageSubscriptionRemoval`, `Nadrf_DataManagement_RetrievalRequest`, `Nadrf_DataManagement_RetrievalSubscribe`, `Nadrf_DataManagement_RetrievalUnsubscribe`, `Nadrf_DataManagement_RetrievalNotify`, `Nadrf_DataManagement_Delete` service operations.  |

---

# A.1 General 

This Annex specifies the formal definition of the API(s) defined in the present specification. It consists of OpenAPI 3.0.0 specifications in YAML format. This Annex takes precedence when being discrepant to other parts of the specification with respect to the encoding of information elements and methods within the API(s).

> **NOTE 1:** The semantics and procedures, as well as conditions, e.g. for the applicability and allowed combinations of attributes or values, not expressed in the OpenAPI definitions but defined in other parts of the specification also apply. 

Informative copies of the OpenAPI specification files contained in this 3GPP Technical Specification are available on a Git-based repository that uses the GitLab software version control system (see 3GPP TS 29.501 [5] clause 5.3.1 and 3GPP TR 21.900 [7] clause 5B).

# A.2 Nadrf_DataManagement API 

```yaml
openapi: 3.0.0 
info: 
  version: 1.2.1 
  title: Nadrf_DataManagement 
  description: | 
    ADRF Data Management Service. 
    © 2026, 3GPP Organizational Partners (ARIB, ATIS, CCSA, ETSI, TSDSI, TTA, TTC). 
    All rights reserved. 
externalDocs: 
  description: 3GPP TS 29.575 V19.6.0; 5G System; Analytics Data Repository Services; Stage 3. 
  url: 'https://www.3gpp.org/ftp/Specs/archive/29_series/29.575/' 
#
servers: 
  - url: '{apiRoot}/nadrf-datamanagement/v1' 
    variables: 
      apiRoot: 
        default: https://example.com 
        description: apiRoot as defined in clause 4.4 of 3GPP TS 29.501. 
#
security: 
  - oAuth2ClientCredentials: 
    - nadrf-datamanagement 
  - {} 
#
paths: 
  /data-store-records: 
    post: 
      summary: Creates a new Individual Data Store Record resource. 
      operationId: CreateADRFDataStoreRecord 
      tags: 
        - ADRF Data Store Records (Collection) 
      requestBody: 
        content: 
          application/json: 
            schema: 
              $ref: '#/components/schemas/NadrfDataStoreRecord' 
        required: true 
        description: ADRF data store record to be stored. 
      responses: 
        '201': 
          description: Successful creation of new Individual ADRF Data Store Record resource. 
          headers: 
            Location: 
              description: > 
                Contains the URI of the newly created resource, according to the structure 
                {apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId} 
              required: true 
              schema: 
                type: string 
          content: 
            application/json: 
              schema: 
                $ref: '#/components/schemas/NadrfDataStoreRecord' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '411': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
        '413': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
        '415': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
      callbacks: 
        storageAlertNotification: 
          '{$request.body#/delNotifUri}': 
            post: 
              requestBody: 
                required: true 
                content: 
                  application/json: 
                    schema: 
                      $ref: '#/components/schemas/NadrfAlertNotification' 
              responses: 
                '200': 
                  description: The alert receipt is acknowledged and a planned action is provided. 
                  content: 
                    application/json: 
                      schema: 
                        $ref: '#/components/schemas/NadrfAlertNotificationResponse' 
                '204': 
                  description: The alert receipt is acknowledged. 
                '307': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
                '308': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
                '400': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
                '401': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
                '403': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
                '404': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
                '411': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
                '413': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
                '415': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
                '429': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
                '500': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
                '502': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
                '503': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
                default: 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
    get: 
      summary: Retrieves existing Individual ADRF Data Store Records. 
      operationId: GetAdrfDataStoreRecords 
      tags: 
        - ADRF Data Store Records (Collection) 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      parameters: 
        - name: store-trans-id 
          description: A storage transaction identifier of a data store record in ADRF. 
          in: query 
          required: false 
          schema: 
            type: string 
        - name: fetch-correlation-ids 
          description: Fetch correlation identifiers received as part of fetch instruction. 
          in: query 
          required: false 
          style: form 
          explode: false 
          schema: 
            type: array 
            items: 
              type: string 
            minItems: 1 
        - name: data-set-id 
          description: The data set identifier. 
          in: query 
          required: false 
          schema: 
            type: string 
      responses: 
        '200': 
          description: Data store records are returned. 
          content: 
            application/json: 
              schema: 
                $ref: '#/components/schemas/NadrfDataStoreRecord' 
        '204': 
          description: No matching ADRF data were found. 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '406': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/406' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /data-store-records/{storeTransId}: 
    delete: 
      summary: Delete an existing Individual ADRF Data Store Record. 
      operationId: DeleteADRFDataStoreRecord 
      tags: 
        - Individual ADRF Data Store Record (Document) 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      parameters: 
        - name: storeTransId 
          in: path 
          description: String identifying a Data Store Record in ADRF. 
          required: true 
          schema: 
            type: string 
      responses: 
        '204': 
          description: > 
            No Content. The Individual ADRF Data Store Record resource matching the 
            storeTransId was deleted. 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /data-retrieval-subscriptions: 
    post: 
      summary: Creates a new Individual ADRF Data Retrieval Subscription resource. 
      operationId: CreateADRFDataRetrievalSubscription 
      tags: 
        - ADRF Data Retrieval Subscriptions (Collection) 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      requestBody: 
        content: 
          application/json: 
            schema: 
              $ref: '#/components/schemas/NadrfDataRetrievalSubscription' 
        required: true 
        description: Individual ADRF Data Retrieval Subscription resource to be created. 
      responses: 
        '201': 
          description: Created a new Individual ADRF Data Retrieval Subscription resource. 
          headers: 
            Location: 
              description: > 
                Contains the URI of the newly created resource, according to the structure 
                {apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId} 
              required: true 
              schema: 
                type: string 
          content: 
            application/json: 
              schema: 
                $ref: '#/components/schemas/NadrfDataRetrievalSubscription' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '411': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
        '413': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
        '415': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
      callbacks: 
        adrfDataRetrievalNotification: 
          '{$request.body#/notificationURI}': 
            post: 
              requestBody: 
                required: true 
                content: 
                  application/json: 
                    schema: 
                      $ref: '#/components/schemas/NadrfDataRetrievalNotification' 
              responses: 
                '204': 
                  description: The receipt of the Notification is acknowledged. 
                '307': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
                '308': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
                '400': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
                '401': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
                '403': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
                '404': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
                '411': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
                '413': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
                '415': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
                '429': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
                '500': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
                '502': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
                '503': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
                default: 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /data-retrieval-subscriptions/{subscriptionId}: 
    delete: 
      summary: Delete an existing Individual ADRF Data Retrieval Subscription resource. 
      operationId: DeleteADRFDataRetrievalSubscription 
      tags: 
        - Individual ADRF Data Retrieval Subscription (Document) 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      parameters: 
        - name: subscriptionId 
          in: path 
          description: > 
            String identifying a data retrieval subscription to the Nadrf_DataManagement 
            Service. 
          required: true 
          schema: 
            type: string 
      responses: 
        '204': 
          description: > 
            No Content. The Individual ADRF Data Retrieval Subscription resource matching 
            the subscriptionId was deleted. 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /request-storage-sub: 
    post: 
      summary: Triggers the creation of a new ADRF Storage Subscription. 
      operationId: CreateADRFStorageSubscription 
      tags: 
        - ADRF Storage Subscriptions 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      requestBody: 
        content: 
          application/json: 
            schema: 
              $ref: '#/components/schemas/NadrfDataStoreSubscription' 
        required: true 
      responses: 
        '200': 
          description: > 
            Successful response with reference used to identify the subscription at the ADRF. 
          content: 
            application/json: 
              schema: 
                $ref: '#/components/schemas/NadrfDataStoreSubscriptionRef' 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '411': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
        '413': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
        '415': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
      callbacks: 
        storageSubAlertNotification: 
          '{$request.body#/delNotifUri}': 
            post: 
              requestBody: 
                required: true 
                content: 
                  application/json: 
                    schema: 
                      $ref: '#/components/schemas/NadrfAlertNotification' 
              responses: 
                '200': 
                  description: The alert receipt is acknowledged and a planned action is provided. 
                  content: 
                    application/json: 
                      schema: 
                        $ref: '#/components/schemas/NadrfAlertNotificationResponse' 
                '204': 
                  description: The alert receipt is acknowledged. 
                '307': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
                '308': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
                '400': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
                '401': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
                '403': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
                '404': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
                '411': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
                '413': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
                '415': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
                '429': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
                '500': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
                '502': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
                '503': 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
                default: 
                  $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /request-storage-sub-removal: 
    post: 
      summary: Triggers the removal of ADRF storage subscription. 
      operationId: DeleteADRFStorageSubscription 
      tags: 
        - ADRF Storage Subscriptions 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      requestBody: 
        content: 
          application/json: 
            schema: 
              $ref: '#/components/schemas/NadrfDataStoreSubscriptionRef' 
        required: true 
      responses: 
        '204': 
          description: > 
            No Content. The ADRF Storage Subscription matching the provided reference was deleted. 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '411': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
        '413': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
        '415': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
  /remove-stored-data-analytics: 
    post: 
      summary: Remove ADRF data based on data or analytics specification. 
      operationId: DeleteADRFData 
      tags: 
        - ADRF Stored Data 
      security: 
        - {} 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
        - oAuth2ClientCredentials: 
          - nadrf-datamanagement 
          - nadrf-datamanagement:storage-read-delete-subs 
      requestBody: 
        content: 
          application/json: 
            schema: 
              $ref: '#/components/schemas/NadrfStoredDataSpec' 
        required: true 
      responses: 
        '204': 
          description: No Content. The ADRF data matching the provided specification is deleted. 
        '307': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/307' 
        '308': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/308' 
        '400': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/400' 
        '401': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/401' 
        '403': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/403' 
        '404': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/404' 
        '411': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/411' 
        '413': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/413' 
        '415': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/415' 
        '429': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/429' 
        '500': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/500' 
        '502': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/502' 
        '503': 
          $ref: 'TS29571_CommonData.yaml#/components/responses/503' 
        default: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/default' 
#
components: 
  securitySchemes: 
    oAuth2ClientCredentials: 
      type: oauth2 
      flows: 
        clientCredentials: 
          tokenUrl: '{nrfApiRoot}/oauth2/token' 
          scopes: 
            nadrf-datamanagement: Access to the nadrf-datamanagement API 
            nadrf-datamanagement:storage-read-delete-subs: > 
              Access to service operations applying to 
              Nadrf_DataManagement_StorageRequest Nadrf_DataManagement_StorageSubscriptionRequest, 
              Nadrf_DataManagement_StorageSubscriptionRemoval, 
              Nadrf_DataManagement_RetrievalRequest, 
              Nadrf_DataManagement_RetrievalSubscribe, Nadrf_DataManagement_RetrievalUnsubscribe, 
              Nadrf_DataManagement_RetrievalNotify, Nadrf_DataManagement_Delete service operations 
#
  schemas: 
#
    NadrfDataStoreRecord: 
      description: Represents an Individual ADRF Data Store Record. 
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
          items: 
            $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscriptionNotification' 
          minItems: 1 
          description: List of analytics subscription notifications. 
        anaSub: 
          type: array 
          items: 
            $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscription' 
          minItems: 1 
          description: > 
            Represents the subscription information of the corresponding analytics notification. 
        dataSub: 
          type: array 
          items: 
            $ref: '#/components/schemas/DataSubscription' 
          minItems: 1 
          description: > 
            Represents the subscription information of the corresponding data notification. 
        storeHandl: 
          $ref: '#/components/schemas/StorageHandlingInfo' 
        dataSetTag: 
          $ref: '#/components/schemas/DataSetTag' 
        dsc: 
          type: string 
          description: Data synthesis and compression information. 
        suppFeat: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/SupportedFeatures' 
#
    NadrfDataStoreSubscription: 
      description: > 
        Contains information to be used by the ADRF to create a Data or Analytics subscription. 
      type: object 
      allOf: 
        - oneOf: 
          - required: [anaSub] 
          - required: [dataSub] 
        - oneOf: 
          - required: [targetNfId] 
          - required: [targetNfSetId] 
      properties: 
        anaSub: 
          $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscription' 
        dataSetTag: 
          $ref: '#/components/schemas/DataSetTag' 
        dataSub: 
          $ref: '#/components/schemas/DataSubscription' 
        targetNfId: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/NfInstanceId' 
        targetNfSetId: 
          $ref: 'TS29571_CommonData.yaml#/components/responses/NfSetId' 
        formatInstruct: 
          $ref: 'TS29574_Ndccf_DataManagement.yaml#/components/schemas/FormattingInstruction' 
        procInstruct: 
          $ref: 'TS29574_Ndccf_DataManagement.yaml#/components/schemas/ProcessingInstruction' 
        multiProcInstructs: 
          type: array 
          items: 
            $ref: 'TS29574_Ndccf_DataManagement.yaml#/components/schemas/ProcessingInstruction' 
          minItems: 1 
          description: Processing instructions to be used for sending event notifications. 
        storeHandl: 
          $ref: '#/components/schemas/StorageHandlingInfo' 
        suppFeat: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/SupportedFeatures' 
#
    NadrfDataRetrievalSubscription: 
      description: Represents an Individual ADRF Data Retrieval Subscription. 
      type: object 
      required: 
        - notifCorrId 
        - notificationURI 
        - timePeriod 
      oneOf: 
        - required: [anaSub] 
        - required: [dataSub] 
        - required: [dataSetId] 
      properties: 
        anaSub: 
          $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscription' 
        dataSetId: 
          type: string 
          description: data set identifier of the data or analytics that are subscribed. 
        dataSub: 
          $ref: '#/components/schemas/DataSubscription' 
        notificationURI: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/Uri' 
        timePeriod: 
          $ref: 'TS29122_CommonData.yaml#/components/schemas/TimeWindow' 
        notifCorrId: 
          type: string 
          description: Notification correlation identifier. 
        consTrigNotif: 
          type: boolean 
          description: > 
            It indicates that notifications shall be buffered (sending only fetch instructions 
            to the NF service consumer) until the NF service consumer requests their delivery 
            using Nadrf_DataManagement Service. 
        suppFeat: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/SupportedFeatures' 
#
    NadrfDataRetrievalNotification: 
      description: > 
        Represents a notification that corresponds with an Individual ADRF Data 
        Retrieval Subscription. 
      type: object 
      required: 
        - notifCorrId 
        - timeStamp 
      oneOf: 
        - required: [anaNotifications] 
        - required: [dataNotif] 
        - required: [fetchInstruct] 
      properties: 
        notifCorrId: 
          type: string 
          description: Notification correlation identifier. 
        anaNotifications: 
          type: array 
          items: 
            $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscriptionNotification' 
          minItems: 1 
          description: List of analytics subscription notifications. 
        dataNotif: 
          $ref: '#/components/schemas/DataNotification' 
        fetchInstruct: 
          $ref: 'TS29576_Nmfaf_3caDataManagement.yaml#/components/schemas/FetchInstruction' 
        terminationReq: 
          type: boolean 
          description: > 
            It indicates the termination of the data management subscription that requested by the 
            ADRF. 
        dsc: 
          type: string 
          description: Data synthesis and compression information. 
        timeStamp: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/DateTime' 
#
    NadrfDataStoreSubscriptionRef: 
      description: Contains a reference to a request for a Data or Analytics subscription. 
      type: object 
      oneOf: 
        - required: [transRefId] 
        - required: [dataSetId] 
      properties: 
        transRefId: 
          type: string 
          description: Transaction reference identifier. 
        dataSetId: 
          type: string 
          description: data set identifier of data or analytics. 
#
    NadrfStoredDataSpec: 
      description: Contains information about Data or Analytics specification. 
      type: object 
      required: 
        - timePeriod 
      oneOf: 
        - required: [dataSpec] 
        - required: [anaSpec] 
        - required: [dataSetId] 
      properties: 
        dataSpec: 
          $ref: '#/components/schemas/DataSubscription' 
        anaSpec: 
          $ref: 'TS29520_Nnwdaf_EventsSubscription.yaml#/components/schemas/NnwdafEventsSubscription' 
        timePeriod: 
          $ref: 'TS29122_CommonData.yaml#/components/schemas/TimeWindow' 
        dataSetId: 
          type: string 
          description: Data set identifier of stored data or analytics records. 
#
    DataSubscription: 
      description: Contains a data specification. 
      type: object 
      oneOf: 
        - required: [amfDataSub] 
        - required: [smfDataSub] 
        - required: [udmDataSub] 
        - required: [nefDataSub] 
        - required: [afDataSub] 
        - required: [nrfDataSub] 
        - required: [nsacfDataSub] 
        - required: [upfDataSub] 
        - required: [gmlcDataSub] 
        - required: [lmfDataSub] 
        - required: [pcfDataSub] 
      properties: 
        amfDataSub: 
          $ref: 'TS29518_Namf_EventExposure.yaml#/components/schemas/AmfEventSubscription' 
        smfDataSub: 
          $ref: 'TS29508_Nsmf_EventExposure.yaml#/components/schemas/NsmfEventExposure' 
        udmDataSub: 
          $ref: 'TS29503_Nudm_EE.yaml#/components/schemas/EeSubscription' 
        afDataSub: 
          $ref: 'TS29517_Naf_EventExposure.yaml#/components/schemas/AfEventExposureSubsc' 
        nefDataSub: 
          $ref: 'TS29591_Nnef_EventExposure.yaml#/components/schemas/NefEventExposureSubsc' 
        nrfDataSub: 
          $ref: 'TS29510_Nnrf_NFManagement.yaml#/components/schemas/SubscriptionData' 
        nsacfDataSub: 
          $ref: 'TS29536_Nnsacf_SliceEventExposure.yaml#/components/schemas/SACEventSubscription' 
        upfDataSub: 
          $ref: 'TS29564_Nupf_EventExposure.yaml#/components/schemas/UpfEventSubscription' 
        gmlcDataSub: 
          $ref: 'TS29515_Ngmlc_Location.yaml#/components/schemas/InputData' 
        lmfDataSub: 
          $ref: 'TS29572_Nlmf_DataExposure.yaml#/components/schemas/LmfDataExposureSubscription' 
        pcfDataSub: 
          $ref: 'TS29523_Npcf_EventExposure.yaml#/components/schemas/PcEventExposureSubsc' 
#
    DataNotification: 
      description: Represents a Data Subscription Notification. 
      type: object 
      oneOf: 
        - required: [amfEventNotifs] 
        - required: [smfEventNotifs] 
        - required: [udmEventNotifs] 
        - required: [nefEventNotifs] 
        - required: [afEventNotifs] 
        - required: [nrfEventNotifs] 
        - required: [nsacfEventNotifs] 
        - required: [upfEventNotifs] 
        - required: [gmlcEventNotifs] 
        - required: [lmfEventNotifs] 
        - required: [pcfEventNotifs] 
      properties: 
        amfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29518_Namf_EventExposure.yaml#/components/schemas/AmfEventNotification' 
          minItems: 1 
          description: List of notifications of AMF events. 
        smfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29508_Nsmf_EventExposure.yaml#/components/schemas/NsmfEventExposureNotification' 
          minItems: 1 
          description: List of notifications of SMF events. 
        udmEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29503_Nudm_EE.yaml#/components/schemas/MonitoringReport' 
          minItems: 1 
          description: List of notifications of UDM events. 
        nefEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29591_Nnef_EventExposure.yaml#/components/schemas/NefEventExposureNotif' 
          minItems: 1 
          description: List of notifications of NEF events. 
        afEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29517_Naf_EventExposure.yaml#/components/schemas/AfEventExposureNotif' 
          minItems: 1 
          description: List of notifications of AF events. 
        nrfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29510_Nnrf_NFManagement.yaml#/components/schemas/NotificationData' 
          minItems: 1 
          description: List of notifications of NRF events. 
        nsacfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29536_Nnsacf_SliceEventExposure.yaml#/components/schemas/SACEventReport' 
          minItems: 1 
          description: List of notifications of NSACF events. 
        upfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29564_Nupf_EventExposure.yaml#/components/schemas/NotificationData' 
          minItems: 1 
          description: List of notifications of UPF events. 
        gmlcEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29515_Ngmlc_Location.yaml#/components/schemas/EventNotifyData' 
          minItems: 1 
          description: List of notifications of GMLC events. 
        lmfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29572_Nlmf_DataExposure.yaml#/components/schemas/LmfDataExposureNotification' 
          minItems: 1 
          description: List of notifications of LMF events. 
        pcfEventNotifs: 
          type: array 
          items: 
            $ref: 'TS29523_Npcf_EventExposure.yaml#/components/schemas/PcEventExposureNotif' 
          minItems: 1 
          description: List of notifications of PCF events. 
        timeStamp: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/DateTime' 
#
    StorageHandlingInfo: 
      description: Contains storage handling information about data or analytics. 
      type: object 
      properties: 
        lifetime: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/DurationSec' 
        delNotifUri: 
          $ref: 'TS29571_CommonData.yaml#/components/schemas/Uri' 
        delNotifCorrId: 
          type: string 
          description: Notification correlation identifier for deletion alerts. 
#
    NadrfAlertNotification: 
      description: Contains information about data or analytics that are about to be deleted. 
      type: object 
      properties: 
        alertStorTransId: 
          type: string 
          description: > 
            Storage transaction identifier that can be used to retrieve data or analytics. 
        delNotifCorrId: 
          type: string 
          description: Notification correlation identifier. 
      required: 
        - alertStorTransId 
        - delNotifCorrId 
#
    NadrfAlertNotificationResponse: 
      description: > 
        Contains information about planned actions related to data or analytics 
        that are about to be deleted. 
      type: object 
      properties: 
        retrievalInd: 
          type: boolean 
          description: > 
            Indicates if the NF service consumer has determined to retrieve the data 
            or analytics that are about to be deleted. 
      required: 
        - retrievalInd 
#
    DataSetTag: 
      description: Contains an identifier and a description of associated records. 
      type: object 
      required: 
        - dataSetId 
      properties: 
        dataSetId: 
          type: string 
          description: Data set identifier of data or analytics records. 
        dataSetDesc: 
          type: string 
          description: Data set description of data or analytics records. 
#
