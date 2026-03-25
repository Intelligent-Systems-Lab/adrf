3GPP TS 29.575 V19.6.0 (2026-03)
Technical Specification
3rd Generation Partnership Project; Technical Specification Group Core Network and Terminals; 5G System; Analytics Data Repository Services; Stage 3 (Release 19) 

# Foreword 

This Technical Specification has been produced by the 3rd Generation Partnership Project (3GPP). 
The contents of the present document are subject to continuing work within the TSG and may change following formal TSG approval. 
Should the TSG modify the contents of the present document, it will be re-released by the TSG with an identifying change of release date and an increase in version number as follows: 
Version x.y.z 
where: 
* x the first digit: 
  * 1 presented to TSG for information; 
  * 2 presented to TSG for approval; 
  * 3 or greater indicates TSG approved document under change control. 
* y the second digit is incremented for all changes of substance, i.e. technical enhancements, corrections, updates, etc. 
* z the third digit is incremented when editorial only changes have been incorporated in the document. 

In the present document, modal verbs have the following meanings: 
* **shall** indicates a mandatory requirement to do something 
* **shall not** indicates an interdiction (prohibition) to do something 

The constructions "shall" and "shall not" are confined to the context of normative provisions, and do not appear in Technical Reports. 
The constructions "must" and "must not" are not used as substitutes for "shall" and "shall not". 
Their use is avoided insofar as possible, and they are not used in a normative context except in a direct citation from an external, referenced, non-3GPP document, or so as to maintain continuity of style when extending or modifying the provisions of such a referenced document. 

* **should** indicates a recommendation to do something 
* **should not** indicates a recommendation not to do something 
* **may** indicates permission to do something 
* **need not** indicates permission not to do something 

The construction "may not" is ambiguous and is not used in normative elements. 
The unambiguous constructions "might not" or "shall not" are used instead, depending upon the meaning intended. 

* **can** indicates that something is possible 
* **cannot** indicates that something is impossible 

The constructions "can" and "cannot" are not substitutes for "may" and "need not". 

* **will** indicates that something is certain or expected to happen as a result of action taken by an agency the behaviour of which is outside the scope of the present document 
* **will not** indicates that something is certain or expected not to happen as a result of action taken by an agency the behaviour of which is outside the scope of the present document 
* **might** indicates a likelihood that something will happen as a result of action taken by some agency the behaviour of which is outside the scope of the present document 
* **might not** indicates a likelihood that something will not happen as a result of action taken by some agency the behaviour of which is outside the scope of the present document 

In addition: 
* **is** (or any other verb in the indicative mood) indicates a statement of fact 
* **is not** (or any other negative verb in the indicative mood) indicates a statement of fact 

The constructions "is" and "is not" do not indicate requirements. 

# 1 Scope 

The present document specifies the stage 3 protocol and data model for the ADRF Service Based Interface. 
It provides stage 3 protocol definitions and message flows, and specifies the API for each service offered by the ADRF. 
The 5G System stage 2 architecture and procedures are specified in 3GPP TS 23.501 [2] and 3GPP TS 23.502 [3]. 
The stage 2 definition and procedures of store and retrieve the collected data and analytics are contained in 3GPP TS 23.288 [14] and 3GPP TS 23.502 [3]. 
The Technical Realization of the Service Based Architecture and the Principles and Guidelines for Services Definition are specified in 3GPP TS 29.500 [4] and 3GPP TS 29.501 [5]. 

# 3 Definitions, symbols and abbreviations 

## 3.1 Definitions 

For the purposes of the present document, the terms and definitions given in 3GPP TR 21.905 [1] and the following apply. 
A term defined in the present document takes precedence over the definition of the same term, if any, in 3GPP TR 21.905 [1]. 
None. 

## 3.2 Symbols 

For the purposes of the present document, the following symbols apply: 
None. 

## 3.3 Abbreviations 

For the purposes of the present document, the abbreviations given in 3GPP TR 21.905 [1] and the following apply. 
An abbreviation defined in the present document takes precedence over the definition of the same abbreviation, if any, in 3GPP TR 21.905 [1]. 

* **ADRF** Analytics Data Repository Function 
* **AF** Application Function 
* **AMF** Access and Mobility Management Function 
* **DCCF** Data Collection Coordination Function 
* **GMLC** Gateway Mobile Location Centre 
* **LMF** Location Management Function 
* **MFAF** Messaging Framework Adaptor Function 
* **NEF** Network Exposure Function 
* **NF** Network Function 
* **NRF** Network Repository Function 
* **NWDAF** Network Data Analytics Function 
* **NSACF** Network Slice Admission Control Function 
* **PCF** Policy Control Function 
* **SMF** Session Management Function 
* **UDM** Unified Data Management 
* **UPF** User Plane Function 

# 4 Services offered by the ADRF 

## 4.1 Introduction 

The Analytics Data Repository Service is used for the Analytics Data Repository Function (ADRF) to storage and retrieval of data and ML model(s) by e.g.  NF service consumers (e.g. NWDAF) which access the data or ML model(s) using Nadrf service. 
The ADRF offers to NF service consumers the following services: 

**Table 4.1-1: Service provided by ADRF** 

| Service Name | Description | Service Operations | Operation Semantics | Example Consumer(s) |
| :--- | :--- | :--- | :--- | :--- |
| Nadrf_DataManagement | This service enables the NF service consumers to store, retrieve and remove the data or analytics in an ADRF. | StorageRequest | Request / Response | DCCF, NWDAF, MFAF |
| | | StorageSubscriptionRequest | Request / Response | DCCF, NWDAF |
| | | StorageSubscriptionRemoval | Request / Response | DCCF, NWDAF |
| | | RetrievalRequest | Request / Response | DCCF, NWDAF |
| | | RetrievalSubscribe | Subscribe / Notify | DCCF, NWDAF |
| | | RetrievalUnsubscribe | Subscribe / Notify | DCCF, NWDAF |
| | | RetrievalNotify | Subscribe / Notify | DCCF, NWDAF |
| | | Delete | Request / Response | DCCF, NWDAF |
| Nadrf_MLModelManagement | This service enables the NF service consumers to store, retrieve and delete ML model(s) in an ADRF. | StorageRequest | Request / Response | NWDAF |
| | | RetrievalRequest | Request / Response | NWDAF, LMF |
| | | Delete | Request / Response | NWDAF |


> **NOTE 1:** The services correspond to the Nadrf_DataManagement service as defined in 3GPP TS 23.288 [14]. 
> **NOTE 2:** The services correspond to the Nadrf_MLModelManagement service as defined in 3GPP TS 23.288 [14]. 

Table 4.1-2 summarizes the corresponding APIs defined for this specification. 

**Table 4.1-2: API Descriptions** 

| Service Name | Clause | Description | OpenAPI Specification File | apiName | Annex |
| :--- | :--- | :--- | :--- | :--- | :--- |
| Nadrf_DataManagement | 5.1 | API for Nadrf_DataManagement | TS29575_Nadrf_DataManagement.yaml | `nadrf-datamanagement` | A.2 |
| Nadrf_MLModelManagement | 5.2 | API for Nadrf_MLModelManagement | TS29575_Nadrf_MLModelManagement.yaml | `nadrf-mlmodelmanagement` | A.3 |


## 4.2 Nadrf_DataManagement Service 

### 4.2.1 Service Description 

#### 4.2.1.1 Overview 

The Nadrf_DataManagement service as defined in 3GPP TS 23.288 [14], is provided by the Analytics Data Repository Function (ADRF). 
This service: 
* allows NF service consumers to store data or analytics in the ADRF, and request/receive notifications about data or analytics that are about to be deleted; 
* allows NF service consumers to retrieve data or analytics from an ADRF; and 
* allows NF service consumers to delete data or analytics from an ADRF. 

#### 4.2.1.2 Service Architecture 

The 5G System Architecture is defined in 3GPP TS 23.501 [2]. The Network Data Analytics Exposure architecture is defined in 3GPP TS 23.288 [14]. 
The Nadrf_DataManagement service is part of the Nadrf service-based interface exhibited by the Analytics Data Repository Function (ADRF). 
Known consumers of the Nadrf_DataManagement service are: 
* Data Collection Coordination Function (DCCF) 
* Network Data Analytics Function (NWDAF) 
* Messaging Framework Adaptor Function (MFAF) 

The Nadrf_DataManagement service is provided by the ADRF and consumed by the NF service consumers as shown in figure 4.2.1.2-1 for the SBI representation model and in figure 4.2.1.2-2 for the reference point representation model. 

**Figure 4.2.1.2-1: Reference Architecture for the Nadrf_DataManagement Service; SBI representation** 

**Figure 4.2.1.2-2: Nadrf_DataManagement service architecture, reference point representation** 

#### 4.2.1.3 Network Functions 

##### 4.2.1.3.1 Analytics Data Repository Function (ADRF) 

The Analytics Data Repository Function (ADRF) provides the functionality to allow NF service consumers to store, retrieve, and remove data or analytics from the ADRF, and request/receive notifications about data or analytics that are about to be deleted. 

##### 4.2.1.3.2 NF Service Consumers 

The NWDAF and DCCF: 
* supports storing data or analytics in the ADRF, and requesting/receiving notifications about data or analytics that are about to be deleted; 
* supports retrieving data or analytics from an ADRF; and 
* supports deletion data or analytics from an ADRF. 

The MFAF: 
* supports storing data or analytics in the ADRF. 

### 4.2.2 Service Operations 

#### 4.2.2.1 Introduction 

**Table 4.2.2.1-1: Operations of the Nadrf_DataManagement Service** 

| Service operation name | Description | Initiated by |
| :--- | :--- | :--- |
| Nadrf_DataManagement_StorageRequest | This service operation is used by an NF to request the ADRF to store data or analytics. Data or analytics are provided to the ADRF in the request message. | NF service consumer (DCCF, NWDAF, MFAF) |
| Nadrf_DataManagement_StorageSubscriptionRequest | This service operation is used by an NF to request the ADRF to initiate a subscription for data or analytics. Data or analytics provided in notifications as a result of the subsequent subscription by the ADRF are stored in the ADRF. | NF service consumer (DCCF, NWDAF) |
| Nadrf_DataManagement_StorageSubscriptionRemoval | This service operation is used by an NF to request that the ADRF no longer subscribes to data or analytics it is collecting and storing. | NF service consumer (DCCF, NWDAF) |
| Nadrf_DataManagement_RetrievalRequest | This service operation is used by an NF to retrieve stored data or analytics from the ADRF. | NF service consumer (DCCF, NWDAF) |
| Nadrf_DataManagement_RetrievalSubscribe | This service operation is used by an NF to retrieve stored data or analytics from the ADRF and to receive future notifications containing the corresponding data or analytics received by ADRF. | NF service consumer (DCCF, NWDAF) |
| Nadrf_DataManagement_RetrievalUnsubscribe | This service operation is used by an NF to request that the ADRF no longer sends data or analytics to a notification endpoint. | NF service consumer (DCCF, NWDAF) |
| Nadrf_DataManagement_RetrievalNotify | This service operation is used by the ADRF to notify an NF with either data or analytics, or instructions to fetch the data or analytics from the ADRF. It is also used by the ADRF to notify NF service consumers about data or analytics that are about to be deleted. | ADRF |
| Nadrf_DataManagement_Delete | This service operation is used by an NF to delete stored data or analytics in ADRF. | NF service consumer (DCCF, NWDAF) |


#### 4.2.2.2 Nadrf_DataManagement_StorageRequest service operation 

##### 4.2.2.2.1 General 

The Nadrf_DataManagement_StorageRequest service operation is used by an NF service consumer to store data or analytics. 

##### 4.2.2.2.2 Request Storage of data or analytics 

Figure 4.2.2.2.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to store data or analytics. 

**Figure 4.2.2.2.2-1: NF service consumer requesting to store data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_StorageRequest service operation to store data or analytics. 
The NF service consumer shall send an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records` as Resource URI representing the "ADRF Data Store Records" resource, as shown in figure 4.2.2.2.2-1, step 1, to create an "Individual ADRF Data Store Record" according to the information in the message body. 
The `NadrfDataStoreRecord` data structure provided in the request body shall include: 
* one of the following: 
  * analytics subscription notification(s) within the `` `anaNotifications` `` attribute together with the corresponding subscription information within the `` `anaSub` `` attribute; 
  * data subscription notification within the `` `dataNotif` `` attribute together with the corresponding subscription information within the `` `dataSub` `` attribute. 

and may include: 
* storage handling information within the `` `storeHandl` `` attribute, if the "EnhDataMgmt" feature is supported. 
* a data set tag within the `` `dataSetTag` `` attribute, if the "EnhDataMgmt" feature is supported; 
* data synthesis and compression information within the `` `dsc` `` attribute, if the "EnhDataMgmt" feature is supported. 

> **NOTE:** The data synthesis and compression information can include an indication that the data have been generated using a data synthesis tool, an indication that the data have been generated using a data compression tool, and information about the data synthesis and/or compression technique. 

Upon the reception of an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records` as Resource URI and `NadrfDataStoreRecord` data structure as request body, the ADRF shall: 
* create a new data store record; 
* assign a `storeTransId`; 
* store the data or analytics. 

> **NOTE 1:** If the data and/or analytics is already stored or being stored in the ADRF, the ADRF will still create a new "Individual ADRF Data Store Record" resource and assign a new `storeTransId` if the ADRF intends to not really store the data again in the memory again based on the implementation. 

If the ADRF created an "Individual ADRF Data Store Record" resource, the ADRF shall respond with "201 Created" with the message body containing a representation of the created record, as shown in figure 4.2.2.2.2-1, step 2. The ADRF shall include a Location HTTP header field. 
The Location header field shall contain the URI of the created record i.e. `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId}`. 
If the ADRF receives storage handling information in the request but determines (e.g. based on local policy) that a different storage approach shall be followed, it indicates the determined storage approach to the consumer by setting accordingly the `` `storeHandl` `` attribute (e.g. providing a different lifetime, or omitting the deletion notification URI to indicate that no deletion alerts will be sent) in the message body of the response. 
When more than one consumer has requested storage lifetime for the same data or analytics, the storage approach should be based on the longest requested storage lifetime. 

> **NOTE 2:** The default operator policy for how long data or analytics are to be stored can be longer or shorter than the lifetime requested by the consumer.  A default operator policy can for example accept only consumer requested lifetimes that are shorter or longer than the default policy. 

If an error occurs when processing the HTTP POST request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 

#### 4.2.2.3 Nadrf_DataManagement_StorageSubscriptionRequest service operation 

##### 4.2.2.3.1 General 

The Nadrf_DataManagement_StorageSubscriptionRequest service operation is used by an NF service consumer to request that the ADRF creates a subscription to data or analytics and subsequently stores notified data or analytics in the ADRF. 

##### 4.2.2.3.2 Requesting subscription and storage of data or analytics 

Figure 4.2.2.3.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to subscribe for data or analytics to be stored in the ADRF. 

**Figure 4.2.2.3.2-1: NF service consumer requesting that the ADRF subscribes to and subsequently stores data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_StorageSubscriptionRequest service operation to request the ADRF to subscribe to data or analytics. 
The NF service consumer shall send an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub` as URI, as shown in figure 4.2.2.3.2-1, step 1. The `NadrfDataStoreSubscription` data structure provided in the request body shall include: 
* one of the following subscription attributes: 
  * analytics subscription information within the `` `anaSub` `` attribute; 
  * data subscription information within the `` `dataSub` `` attribute; 
* one of the following target identifiers: 
  * DCCF or NWDAF instance identifier within the `` `targetNfId` `` attribute; 
  * DCCF or NWDAF NF set identifier within the `` `targetNfSetId` `` attribute; 

and may include: 
* formatting instructions within the `` `formatInstruct` `` attribute; 
* processing instructions within the `` `procInstruct` `` attribute or the `` `multiProcInstructs` `` attribute if the "MultiProcessingInstruction" feature is supported. 

> **NOTE 1:** The parameters provided by the NF service consumer, including Formatting and Processing Instructions (if provided) are used by the ADRF when subscribing to a DCCF or NWDAF for Data or Analytics to be stored. 

* storage handling information within the `` `storeHandl` `` attribute, if the "EnhDataMgmt" feature is supported. 
* a data set tag to be associated with this subscription and with the data or analytics collected based on this subscription within the `` `dataSetTag` `` attribute, if the "EnhDataMgmt" feature is supported. 

Upon the reception of an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub` as URI and `NadrfDataStoreSubscription` data structure as request body, the ADRF shall assign a transaction reference identifier to this request and, if the request is successfully processed and accepted, the ADRF shall respond with "200 OK" as shown in figure 4.2.2.3.2-1 step 2, with the message body containing an `NadrfDataStoreSubscriptionRef` data structure, which shall include the assigned transaction reference identifier as `` `transRefId` `` attribute. 

> **NOTE 2:** If the data and/or analytics is already stored or being stored in the ADRF, the ADRF will still assign a new transaction reference identifier if the ADRF intends to not really store the data again in the memory again based on the implementation. 

If an error occurs when processing the HTTP POST request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 
If the ADRF receives storage handling information in the request but determines (e.g. based on local policy) that a different storage approach shall be followed, it indicates the determined storage approach to the consumer by setting accordingly the `` `storeHandl` `` attribute (e.g. providing a different lifetime, or omitting the deletion notification URI to indicate that no deletion alerts will be sent) in the message body of the response. 
When more than one consumer has requested storage lifetime for the same data or analytics, the storage approach should be based on the longest requested storage lifetime. 

> **NOTE 3:** The default operator policy for how long data or analytics are to be stored can be longer or shorter than the lifetime requested by the consumer.  A default operator policy can for example accept only consumer requested lifetimes that are shorter or longer than the default policy. 

In the case of a successful response, the ADRF may subsequently create a data or analytics subscription (according to inputs that had been received in the `NadrfDataStoreSubscription` data structure; this is not performed if the ADRF determines that the data is already being stored based on an existing subscription) with a DCCF as described in 3GPP TS 29.574 [23] or with an NWDAF as described in 3GPP TS 29.520 [15], and create a mapping between the previously assigned and returned transaction reference identifier and the subscription that is used to serve the transaction. 

#### 4.2.2.4 Nadrf_DataManagement_StorageSubscriptionRemoval service operation 

##### 4.2.2.4.1 General 

The Nadrf_DataManagement_StorageSubscriptionRemoval service operation is used by an NF service consumer to request the ADRF to remove a subscription for data or analytics. 

##### 4.2.2.4.2 Requesting removal of subscription of data or analytics 

Figure 4.2.2.4.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to unsubscribe for storage of data or analytics. 

**Figure 4.2.2.4.2-1: NF service consumer requesting the removal of subscription(s) to storage of data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_StorageSubscriptionRemoval service operation to request the ADRF to remove subscription(s) to data or analytics that are stored in the ADRF. 
The NF service consumer shall send an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub-removal` as URI, as shown in figure 4.2.2.4.2-1, step 1. The POST request body shall contain an `NadrfDataStoreSubscriptionRef` data structure, which shall include a transaction reference identifier as `` `transRefId` `` attribute or, if the "EnhDataMgmt" feature is supported, a data set identifier as `` `dataSetId` `` attribute. 
Upon the reception of an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/request-storage-sub-removal` as URI, if the ADRF successfully processed and accepted the received HTTP POST request, the ADRF shall respond with HTTP "204 No Content" status. 
Subsequently, the ADRF shall remove the (DCCF or NWDAF) subscription that had been created and mapped to the received transaction reference identifier or the (DCCF or NWDAF) subscription(s) associated to the received data set identifier as described in clause 4.2.2.3, unless this subscription is mapped to further transaction reference identifier(s) (of transactions that are still active) or associated with further data set identifier(s). 
If errors occur when processing the HTTP POST request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 

#### 4.2.2.5 Nadrf_DataManagement_RetrievalRequest service operation 

##### 4.2.2.5.1 General 

The Nadrf_DataManagement_RetrievalRequest service operation is used by an NF service consumer to retrieve stored data or analytics. 

> **NOTE:** If the data is to be collected for a user, i.e. SUPI or GPSI, the consumer needs to check the user consent by retrieving the user consent information from the UDM as described in clause 5.5 of 3GPP TS 29.552 [28] before invoking this service operation. 

##### 4.2.2.5.2 Request and get stored data or analytics from ADRF Data Store 

Figure 4.2.2.5.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to retrieve stored data or analytics. 

**Figure 4.2.2.5.2-1: NF service consumer requesting to retrieve stored data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_RetrievalRequest service operation to retrieve stored data or analytics. 
The NF service consumer shall send an HTTP GET request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records` as Resource URI representing the "ADRF Data Store Records" resource, as shown in figure 4.2.2.5.2-1, step 1, to request ADRF data store records according to the query parameter value of the store transaction identifier within the `` `store-trans-id` `` attribute, the query parameter value of the fetch correlation identifier(s) within the `` `fetch-correlation-ids` `` attribute, or, if the "EnhDataMgmt" feature is supported, the query parameter value of the data set identifier within the `` `data-set-id` `` attribute. 
Upon the reception of the HTTP GET request, the ADRF shall: 
* find the data or analytics according to the requested parameters. 

If the requested data or analytics is found, the ADRF shall respond with "200 OK" status code with the message body containing the `NadrfDataStoreRecord` data structure. 
The `NadrfDataStoreRecord` data structure in the response body shall include: 
* one of the following: 
  * information about network analytics function events that occurred in the `` `anaNotifications` `` attribute together with the corresponding subscription information within the `` `anaSub` `` attribute; 
  * information about data event within the `` `dataNotif` `` attribute together with the corresponding subscription information within the `` `dataSub` `` attribute. 

and may include: 
* a data set tag within the `` `dataSetTag` `` attribute, if the "EnhDataMgmt" feature is supported. 
* data synthesis and compression information within the `` `dsc` `` attribute, if the "EnhDataMgmt" feature is supported. 

> **NOTE:** The data synthesis and compression information can include an indication that the data have been generated using a data synthesis tool, an indication that the data have been generated using a data compression tool, and information about the data synthesis and/or compression technique. 

If the requested analytics or data does not exist, the ADRF shall respond with "204 No Content". 
If an error occurs when processing the HTTP GET request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 

#### 4.2.2.6 Nadrf_DataManagement_RetrievalSubscribe service operation 

##### 4.2.2.6.1 General 

The Nadrf_DataManagement_RetrievalSubscribe service operation is used by an NF service consumer to subscribe to the ADRF to retrieve via notifications data or analytics that is stored in the ADRF and to receive future notifications with data or analytics when they are received by the ADRF. 

> **NOTE:** If the data is to be collected for a user, i.e. SUPI or GPSI, the consumer needs to check the user consent by retrieving the user consent information from the UDM as described in clause 5.5 of 3GPP TS 29.552 [28] before invoking this service operation. 

##### 4.2.2.6.2 Requesting retrieval and subscription of data or analytics 

Figure 4.2.2.6.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to retrieve and subscribe to data or analytics. 

**Figure 4.2.2.6.2-1: NF service consumer requesting to retrieve and subscribe to data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_RetrievalSubscribe service operation to retrieve and subscribe to data or analytics. 
The NF service consumer shall send an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions` as Resource URI representing the "ADRF Data Retrieval Subscriptions" resource, as shown in figure 4.2.2.6.2-1, step 1, to create an "Individual ADRF Data Retrieval Subscription" according to the information in the message body. 
The `NadrfDataRetrievalSubscription` data structure provided in the request body shall include: 
* notification correlation identfier within the `` `notifCorrId` `` attribute; 
* one of the following: 
  * analytics subscription information within the `` `anaSub` `` attribute; 
  * data subscription information within the `` `dataSub` `` attribute; 
* data set identifier within the `` `dataSetId` `` attribute, if the "EnhDataMgmt" feature is supported; 
* a notification target address within the `` `notificationURI` `` attribute; 
* a time window for the data retrieval and subscription within the `` `timePeriod` `` attribute; 

and may include: 
* a Consumer triggered Notification indication within the `` `consTrigNotif` `` attribute. 

Upon the reception of an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions` as Resource URI and `NadrfDataRetrievalSubscription` data structure as request body, the ADRF shall: 
* create a new subscription; 
* assign a `subscriptionId`; 
* store the subscription. 

If the ADRF created an "Individual ADRF Data Retrieval Subscription" resource, the ADRF shall respond with "201 Created" with the message body containing a representation of the created subscription, as shown in figure 4.2.2.6.2-1, step 2. The ADRF shall include a Location HTTP header field. 
The Location header field shall contain the URI of the created record i.e. `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId}`. 
If an error occurs when processing the HTTP POST request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 

#### 4.2.2.7 Nadrf_DataManagement_RetrievalUnsubscribe service operation 

##### 4.2.2.7.1 General 

The Nadrf_DataManagement_RetrievalUnsubscribe service operation is used by an NF service consumer to remove a retrieval subscription to data or analytics. 

##### 4.2.2.7.2 Requesting removal of retrieval subscription for data or analytics 

Figure 4.2.2.7.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to remove a retrieval subscription for data or analytics. 

**Figure 4.2.2.7.2-1: NF service consumer requesting to remove retrieval subscription for data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_RetrievalUnsubscribe service operation to remove a retrieval subscription for data or analytics. 
The NF service consumer shall send an HTTP DELETE request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId}` as Resource URI representing an "Individual ADRF Data Retrieval Subscription" resource, as shown in figure 4.2.2.7.2-1, step 1, where `{subscriptionId}` is the identifier of the existing data retrieval subscription that is to be deleted. 
Upon the reception of an HTTP DELETE request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-retrieval-subscriptions/{subscriptionId}` as Resource URI, if the ADRF successfully processed and accepted the received HTTP DELETE request, the ADRF shall: 
* remove the corresponding subscription; 
* respond with HTTP "204 No Content" status. 

If errors occur when processing the HTTP DELETE request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 
If the ADRF determines the received HTTP DELETE request needs to be redirected, the ADRF shall send an HTTP redirect response as specified in clause 6.10.9 of 3GPP TS 29.500 [4]. 

#### 4.2.2.8 Nadrf_DataManagement_RetrievalNotify service operation 

##### 4.2.2.8.1 General 

The Nadrf_DataManagement_RetrievalNotify service operation is used by ADRF to notify NF service consumers about subscribed events related to data or analytics and about data or analytics that are about to be deleted. 

##### 4.2.2.8.2 Notification about subscribed data or analytics 

Figure 4.2.2.8.2-1 shows a scenario where the ADRF sends a request to the NF service consumer to notify it about data or analytics events. 

**Figure 4.2.2.8.2-1: ADRF notifies the NF service consumer about subscribed data or analytics** 

The ADRF shall invoke the Nadrf_DataManagement_RetrievalNotify service operation to notify about subscribed data or analytics events. 
The ADRF shall send an HTTP POST request to the `{notificationURI}` received in the subscription (see clause 5.1.5 for the definition of this notificationURI), as shown in figure 4.2.2.8.2-1, step 1. The `NadrfDataRetrievalNotification` data structure provided in the request body shall include: 
* notification correlation Id within the `` `notifCorrId` `` attribute; 
* the time stamp which represents the time when ADRF completes preparation of the requested data or analytics within the `` `timeStamp` `` attribute; 
* one of the following: 
  * information about network data analytics function events that occurred in the `` `anaNotifications` `` attribute; 
  * data collected from data sources (e.g. SMF, NEF) in the `` `dataNotif` `` attribute; 
  * information for fetching the data or analytics in the `` `fetchInstruct` `` attribute. 

> **NOTE:** The fetch correlation identifiers included in the fetch instructions of the `` `fetchInstruct` `` attribute can be used to fetch data or analytics using the Nadrf_DataManagement_RetrievalRequest service operation as described in clause 4.2.2.5.2.  The (mandatory) fetch URI included in the fetch instructions of the `` `fetchInstruct` `` attribute is expected to be in line with the standard resource URI defined for the Nadrf_DataManagement_RetrievalRequest service operation, i.e. `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records`, but it can be anything because it is actually not needed by the NF service consumer in this case. 

The `NadrfDataRetrievalNotification` data structure provided in the request body may include: 
* a termination request provided by the ADRF within the `` `terminationReq` `` attribute. 
* data synthesis and compression information within the `` `dsc` `` attribute, if the "EnhDataMgmt" feature is supported. 

> **NOTE:** The data synthesis and compression information can include an indication that the data have been generated using a data synthesis tool, an indication that the data have been generated using a data compression tool, and information about the data synthesis and/or compression technique. 

Upon the reception of an HTTP POST request with `{notificationURI}` as Resource URI and `NadrfDataRetrievalNotification` data structure as request body, if the NF service consumer successfully processed and accepted the received HTTP POST request, the NF Service Consumer shall: 
* store the notification; 
* respond with HTTP "204 No Content" status code. 

If errors occur when processing the HTTP POST request, the NF service consumer shall send an HTTP error response as specified in clause 5.1.7. 
If the NF service consumer determines the received HTTP POST request needs to be redirected, the NF service consumer shall send an HTTP redirect response as specified in clause 6.10.9 of 3GPP TS 29.500 [4]. 
After the successful processing of the HTTP POST request, if the ADRF requests the NF service consumer to retrieve the data or analytics with the `` `fetchInstruct` `` attribute, the NF service consumer may invoke the Nadrf_DataManagement_RetrievalRequest service operation to retrieve the notified data or analytics as defined in clause 4.2.2.5. 

##### 4.2.2.8.3 Notification about data or analytics that are about to be deleted 

Figure 4.2.2.8.3-1 shows a scenario where the ADRF sends a request to the NF service consumer to notify it about data or analytics that are about to be deleted. 

**Figure 4.2.2.8.3-1: ADRF notifies the NF service consumer about data or analytics that are about to be deleted.** 

In order to notify about data or analytics that are about to be deleted, the ADRF shall invoke the Nadrf_DataManagement_RetrievalNotify service operation. 
The ADRF shall send an HTTP POST request to the `{delNotifUri}` received in a storage request as defined in clause 4.2.2.2.2 or in a storage subscription request as defined in clause 4.2.2.3.2, as shown in figure 4.2.2.8.3-1, step 1. The `NadrfAlertNotification` data structure provided in the request body shall include: 
* a notification correlation identifier within the `` `delNotifCorrId` `` attribute; 
* a storage transaction identifier, which may be used by the NF service consumer to retrieve the data, within the `` `alertStorTransId` `` attribute; 

> **NOTE:** The `` `alertStorTransId` `` attribute, which is used for retrieving data prior to deletion, does not have to be the same with or related to the `storeTransId` attribute that is assigned and returned during the storage of the data. 

Upon the reception of an HTTP POST request with `{delNotifUri}` as Resource URI and `NadrfAlertNotification` data structure as request body, if the NF service consumer successfully processed and accepted the received HTTP POST request, the NF Service Consumer shall either respond with HTTP "204 No Content" status code or with HTTP "200 OK" status code and the `NadrfAlertNotificationResponse` data structure in the message body. 
If errors occur when processing the HTTP POST request, the NF service consumer shall send an HTTP error response as specified in clause 5.1.7. 
If the NF service consumer determines the received HTTP POST request needs to be redirected, the NF service consumer shall send an HTTP redirect response as specified in clause 6.10.9 of 3GPP TS 29.500 [4]. 
After the successful processing of the HTTP POST request, the NF service consumer may invoke the Nadrf_DataManagement_RetrievalRequest service operation as defined in clause 4.2.2.5, using the storage transaction identifier received within the `` `alertStorTransId` `` attribute of the `NadrfAlertNotification`, in order to retrieve the data or analytics that are about to be deleted. 

#### 4.2.2.9 Nadrf_DataManagement_Delete service operation 

##### 4.2.2.9.1 General 

The Nadrf_DataManagement_Delete service operation is used by an NF service consumer to delete stored data or analytics. 

##### 4.2.2.9.2 Requesting removal of stored data or analytics 

Figure 4.2.2.9.2-1 shows a scenario where the NF service consumer sends a request to the ADRF to delete stored data or analytics. 

**Figure 4.2.2.9.2-1: NF service consumer requesting to remove stored data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_Delete service operation to remove stored data or analytics. 
The NF service consumer shall send an HTTP DELETE request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId}` as Resource URI representing an "Individual ADRF Data Store Record" resource, as shown in figure 4.2.2.9.2-1, step 1, where `{storeTransId}` is the transaction identifier of the stored record that is to be deleted. 
Upon the reception of an HTTP DELETE request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/data-store-records/{storeTransId}` as Resource URI, if the ADRF successfully processed and accepted the received HTTP DELETE request, the ADRF shall: 
* remove the corresponding stored record; 
* respond with HTTP "204 No Content" status code. 

If errors occur when processing the HTTP DELETE request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 
If the ADRF determines the received HTTP DELETE request needs to be redirected, the ADRF shall send an HTTP redirect response as specified in clause 6.10.9 of 3GPP TS 29.500 [4]. 

##### 4.2.2.9.3 Requesting removal of stored data or analytics using data or analytics specification 

Figure 4.2.2.9.3-1 shows a scenario where the NF service consumer sends a request to the ADRF to delete stored data or analytics based on a data or analytics specification. 

**Figure 4.2.2.9.3-1: NF service consumer requesting to remove stored data or analytics** 

The NF service consumer shall invoke the Nadrf_DataManagement_Delete service operation to remove stored data or analytics based on a data or analytics specification. 
The NF service consumer shall send an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/remove-stored-data-analytics` as URI, as shown in figure 4.2.2.9.3-1, step 1. The POST request body shall contain an `NadrfStoredDataSpec` data structure. 
The `NadrfStoredDataSpec` data structure provided in the request body shall include: 
* a time window in which the data to be deleted was collected in the `` `timePeriod` `` attribute; and 
* one of the following: 
  * a data specification in the `` `dataSpec` `` attribute; 
  * an analytics specification in the `` `anaSpec` `` attribute; 
  * a data set identifier within the `` `dataSetId` `` attribute, if the "EnhDataMgmt" feature is supported. 

Upon the reception of an HTTP POST request with `{apiRoot}/nadrf-datamanagement/<apiVersion>/remove-stored-data-analytics` as URI, if the ADRF successfully processed and accepted the received HTTP POST request, the ADRF shall respond with HTTP "204 No Content" status. 
The ADRF shall remove any stored analytics or data that match the analytics specification, the data specification, or the data set identifier received in the request. 
If errors occur when processing the HTTP POST request, the ADRF shall send an HTTP error response as specified in clause 5.1.7. 