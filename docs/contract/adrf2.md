---

## 1. 概述

記錄 NWDAF 整合 ADRF 的設計決策與 E2E 流程。
整合目的：**在 retrain 時，由 ADRF 提供歷史 UPF 流量資料作為 Daisy 的訓練集。**

規格參照：3GPP TS 29.575（Nadrf_DataManagement Service）。

---

## 2. 規格重點摘要（TS 29.575）

### 2.1 Nadrf_DataManagement 服務操作

| 操作 | 方向 | 說明 |
| --- | --- | --- |
| `StorageRequest` | NWDAF → ADRF | 將資料存入 ADRF，回 201 + storeTransId |
| `RetrievalSubscribe` | NWDAF → ADRF | 訂閱歷史資料，ADRF 主動 push 通知（含 fetch correlation IDs） |
| `RetrievalNotify` | ADRF → NWDAF | 通知內容可以是直接帶資料，或帶 fetchInstruct（ID list） |
| `RetrievalRequest` | NWDAF → ADRF | 用 fetch-correlation-ids GET 實際資料 |
| `RetrievalUnsubscribe` | NWDAF → ADRF | 取消 retrieval 訂閱 |

### 2.2 關鍵 Schema

**`NadrfDataStoreRecord`**（存入與取出的資料格式）：

```
oneOf:
  - anaSub + anaNotifications  （analytics 路徑）
  - dataSub + dataNotif        （原始資料路徑）
```

**`DataSubscription`**（dataSub 的型別，每個物件 oneOf）：

```
oneOf: amfDataSub | smfDataSub | upfDataSub | udmDataSub | ...
```

`dataSub` 是 **array**，允許一筆 dataNotif 對應多個訂閱描述。

**`DataNotification`**（dataNotif 的型別，oneOf）：

```
oneOf: amfEventNotifs | smfEventNotifs | upfEventNotifs | ...
```

---

## 3. smfDataSub + upfEventNotifs 設計決策

### 3.1 問題背景

NWDAF 取 UPF 資料的路徑：

1. NWDAF 向 **SMF** 訂閱，指定 event = `UPF_EVENT`
2. SMF 在內部向 UPF 建立訂閱（`UpfEventSubscription`）—— NWDAF 不可見、無法取得
3. UPF **直接**將資料（`NotificationData`）送至 NWDAF，不經過 SMF

因此 NWDAF 手上有：

- `smfDataSub`（`NsmfEventExposure`）：自己向 SMF 建立的訂閱物件
- `upfEventNotifs`（`NotificationData` array）：UPF 直接送來的資料
- **沒有** `upfDataSub`（`UpfEventSubscription`）：SMF 內部的，NWDAF 無從取得

### 3.2 合規分析

**線索 1 — Schema 無跨欄位約束**：`dataSub` 的 NF 選擇與 `dataNotif` 的 NF 選擇是兩個獨立的 `oneOf`，schema 層沒有要求兩者選同一個 NF 來源。

**線索 2 — "corresponding" 的語義**：spec 說 dataSub 是 "the subscription information of the corresponding data notification"。「corresponding」可解讀為「導致此筆資料被收到的那條訂閱」——即 NWDAF 向 SMF 建立的 `NsmfEventExposure`，而非 SMF 內部對 UPF 的訂閱。

**線索 3 — UPF event 只能透過 SMF 訂閱**：TS 29.508 明確說明 `UPF_EVENT: UPF event subscribed via SMF`。若要求 `upfDataSub` 與 `upfEventNotifs` 同源配對，在此情境下 `upfDataSub` 永遠無法填，形成無解死角，不符合規格設計意圖。

**結論**：`smfDataSub + upfEventNotifs` **不違反 TS 29.575 規格**，是此情境下唯一合理可行的做法。

---

## 4. E2E 整合流程

### 4.1 Phase 1：NWDAF 即時存入 ADRF

```
UPF → (NotificationData) → NWDAF
                              ↓
                    [buffer 累積至閾值]
                              ↓
              NWDAF → POST /data-store-records → ADRF
              body: NadrfDataStoreRecord {
                dataSub: [{ smfDataSub: NsmfEventExposure }],
                dataNotif: { upfEventNotifs: [...] }
              }
              → 回 201 + storeTransId
```

**存入觸發條件**：buffer 累積筆數達到設定閾值（config：`adrf.storageThreshold`）後 flush。預設值為 **1**（逐筆存入），日後視需求調高以批次存入。

### 4.2 Phase 2：MTLF 觸發 Retrain 前的資料準備

**步驟 1 — 按 SUPI 向 ADRF 訂閱歷史資料**

NWDAF 維護 SUPI → Group ID 的 map（因為向 SMF 訂閱時無法指定 group，只能逐 SUPI 訂閱）。
MTLF 對每個 SUPI 個別發起 `RetrievalSubscribe`：

```
MTLF → POST /data-retrieval-subscriptions → ADRF
body: NadrfDataRetrievalSubscription {
  notifCorrId: "<TID>",
  notificationURI: "<http://nwdaf/adrf/retrieval-notify>",
  timePeriod: { startTime: ..., stopTime: ... },
  dataSub: { smfDataSub: { supi: "imsi-001", ... } },
  consTrigNotif: true
}
→ 回 201 + subscriptionId
```

`consTrigNotif: true` 讓 ADRF 進入 fetch 模式：callback 只送 `fetchInstruct`（ID list），不直接帶資料本體。`notifCorrId` 與本次 retrain 的 TID 一對一綁定。

**步驟 2 — 接收 fetch correlation ID list**

ADRF 主動 POST 到 `notificationURI`，body 為 `NadrfDataRetrievalNotification`，帶 `fetchInstruct`（fetch correlation IDs），不直接帶資料本體。

MTLF 收到後記錄所有 IDs，等待所有 SUPI 的通知收齊。

**步驟 3 — 逐 ID GET 實際資料**

```
MTLF → GET /data-store-records?fetch-correlation-ids=id1 → ADRF
→ 回 200 + NadrfDataStoreRecord
```

即使 callback 一次帶多個 fetch correlation IDs，NWDAF 仍採一個 ID 一個 ID 逐次 fetch。

`RetrievalRequest` 回應分支（依 TS 29.575）：

1. `200 OK`：有資料，body 為 `NadrfDataStoreRecord`，照既有流程推入 Daisy。
2. `204 No Content`：此查詢條件下沒有可回傳資料（spec 明確允許）；MTLF 將該 fetchCorrId 視為「已完成但無資料」，不重試、繼續下一個 ID。
3. `4xx`（除 204）: 視為請求錯誤或語意錯誤，記錄失敗原因並進入異常收斂（後續仍要做 unsubscribe 清理）。
4. `5xx/timeout`：有限次重試（backoff）；超過上限後進入異常收斂（後續仍要做 unsubscribe 清理）。

規格依據：

- TS 29.575，`Nadrf_DataManagement_RetrievalRequest`（`GET /data-store-records`）描述：requested data 不存在時 ADRF 回 `204 No Content`。
- TS 29.575，`/data-store-records` API response table：`204 No Content` 對應 requested ADRF data store record 不存在。

**callback 批次上限（ADRF）**：
ADRF 可按 `corrIdBatchSize` 將 `fetchCorrIds` 切成多包通知（每包最多 `corrIdBatchSize` 個）。
`corrIdBatchSize` 為 ADRF 端設定值（由 ADRF config 控制），不是 NWDAF 的 fetch request 大小。

GET response 結構（`NadrfDataStoreRecord`）：

```json
{
  "dataSub": [{ "smfDataSub": { "supi": "imsi-001", ... } }],
  "dataNotif": {
    "upfEventNotifs": [
      { "notificationItems": [...], "correlationId": "corr-session-001" }
    ]
  }
}
```

**步驟 4 — 標記資料並推入 Daisy**

每筆 GET 回來的 `NadrfDataStoreRecord`，MTLF：

1. 從 `dataSub[0].smfDataSub.supi` 取出 SUPI
2. 查 SUPI → Group ID map 得到 `group_id`
3. 附上 `TID` 和 `group_id`，將 `dataNotif.upfEventNotifs` 原封推入 Daisy

```json
POST /upload_data

{
  "TID": "abc123",
  "group_id": "ue-comm-group1",
  "upfEventNotifs": [
    { "notificationItems": [...], "correlationId": "corr-session-001" }
  ]
}
```

**步驟 5 — 清理訂閱（RetrievalUnsubscribe）**

正常收斂時，需同時滿足以下條件再送 delete：

1. 已取完所有 queued fetch IDs
2. 已收到 `terminationReq: true`

異常收斂時（retrain job 取消、超時、流程失敗），也應主動送 delete 清理資源。

```
DELETE /data-retrieval-subscriptions/{subscriptionId}
```

MTLF 對回應處理：

```
204: success
404: 視為已清理（V0 當成功）
5xx/timeout: 有限次重試（backoff）
```

**步驟 6 — 觸發 Retrain**

訂閱清理完畢後，MTLF 發起 Daisy retrain，附上 `TID`。
Daisy 以 `TID` 查出對應的 dataset，執行訓練。

### 4.3 整體序列圖（文字描述）

```
UPF  →  NWDAF(AnLF)  →  [buffer]  →  ADRF (持續存入)
                                         ↑
                                    [歷史資料庫]
                                         ↓
MTLF 偵測到 accuracy 下降
  → 按每個 SUPI 發 RetrievalSubscribe 給 ADRF
  → ADRF push fetchInstruct (ID list)
  → MTLF 逐 ID GET 實際資料（每次 1 個 storeTransId）
  → 每筆資料標記 group_id + TID，推入 Daisy
  → 所有資料推送完畢
  → DELETE /data-retrieval-subscriptions/{subscriptionId}
  → MTLF 發 TriggerTraining（帶 TID）給 Daisy
  → Daisy 以 TID 查 dataset → 執行 FL 訓練
  → Daisy callback NWDAF（帶 model_url）
  → NWDAF 換版流程（現有 Phase A/B 流程）
```

---

## 5. 實作細節

### 5.1 NWDAF 側（Go）

| 項目 | 說明 |
| --- | --- |
| Processor buffer（E2 已實作） | `internal/sbi/processor/adrf_buffer.go`：收到 UPF notification 時累積，達 `storageThreshold` 後 flush；無獨立 `infos` map，`AdrfSmfInfo` 在 `add()` 時直接傳入 `flushOne()` |
| ADRF consumer（E2 已實作） | `internal/sbi/consumer/adrf_service.go`：`AdrfClient` + `StorageRequest`；含 connection-pool Transport |
| `smfDataSub` 型別（E2 已實作） | 重用 `ExtendedNsmfEventExposure`（已符合 TS29508 `NsmfEventExposure`），不另建新 struct |
| ADRF notify endpoint | `POST /adrf/retrieval-notify`（E3 待實作）：接收 ADRF 的 fetchInstruct 通知 |
| MTLF retrain 前置 | E3 待實作：retrain 觸發後先完成資料抓取再呼叫 `TriggerTraining` |
| Config（E2 已實作） | `pkg/factory/config.go`：獨立 `AdrfConfig` struct（`Url`、`StorageThreshold`、`FetchBatchSize`），與 `SmfConfig`/`MtlfConfig` 並列；非掛在 `AnlfConfig` 下 |
| Context lifecycle（E2 已實作） | `AdrfSmfInfo` keyed by `correlationId`；SMF 訂閱建立時 `StoreAdrfSmfInfo`；訂閱釋放時 `DeleteAdrfSmfInfo`（`cleanupDataCollection`） |

### 5.2 Daisy 側（新需求）

| 項目 | 說明 |
| --- | --- |
| 資料接收 endpoint | `POST /upload_data`：接收帶有 `TID`、`group_id`、`upfEventNotifs` 的訓練資料 |
| 資料庫 | 持久化儲存各 TID 的 dataset |
| 訓練查詢 | `publish_task` 帶 `TID`，Daisy 以此查出 dataset |

### 5.3 ADRF 相依

ADRF 需已部署且可連線。NWDAF config 新增 `adrf.url`（`AdrfConfig.Url`）。

---

## 6. 分階段實作計畫

### Phase E1 — Daisy 資料接收端

**負責方**：Daisy
**前置條件**：無
**狀態**：待實作

- [ ]  新增 `POST /upload_data` endpoint：接收 `{ TID, group_id, upfEventNotifs }` 並持久化
- [ ]  建立以 TID 為 key 的資料庫，支援持續累積同一 TID 的多筆資料
- [ ]  `publish_task` 收到 TID 後，能從 DB 查出對應 dataset 執行訓練

---

### Phase E2 — NWDAF ADRF 存入端

**負責方**：NWDAF
**前置條件**：無（可與 E1 並行）
**狀態**：✅ 完成

- [x]  `pkg/factory/config.go`：新增獨立 `AdrfConfig` struct（`Url`、`StorageThreshold` 預設 1、`FetchBatchSize` 預設 30），掛在 `Configuration.Adrf`
- [x]  `config/nwdafcfg.yaml`：補上 `adrf:` block（`configuration:` 內，2-space indent）
- [x]  `internal/sbi/consumer/adrf_service.go`（新增）：`AdrfClient` + `StorageRequest`（POST `/data-store-records`），含 connection-pool Transport
- [x]  `internal/sbi/processor/adrf_buffer.go`（新增）：per-correlationId buffer；達 `storageThreshold` 時 flush（`AdrfSmfInfo` 直接傳入，無額外 map）
- [x]  `internal/sbi/processor/upf_notify.go`：收到 UPF notification 後 marshal → 丟入 `adrfBuffer.add()`
- [x]  `internal/context/traffic_data.go`：`AdrfSmfInfo` struct + `StoreAdrfSmfInfo` / `GetAdrfSmfInfo` / `DeleteAdrfSmfInfo`
- [x]  `internal/sbi/processor/data_collection.go`：SMF 訂閱成功後呼叫 `StoreAdrfSmfInfo`
- [x]  `internal/sbi/processor/eventssubscription.go`：`cleanupDataCollection` 加上 `DeleteAdrfSmfInfo`
- [x]  整合測試：fake_smf_upf + fake_adrf 驗證 `NadrfDataStoreRecord` 結構正確（`dataSub[0].smfDataSub` + `dataNotif.upfEventNotifs`）

---

### Phase E3 — NWDAF ADRF 取資料 + retrain 前置整合

**負責方**：NWDAF
**前置條件**：E1、E2
**狀態**：待實作

- [ ]  `internal/sbi/consumer/adrf_service.go`：補上 `RetrievalSubscribe`（POST `/data-retrieval-subscriptions`，帶 `consTrigNotif: true`，`notifCorrId` 綁定 TID）、`RetrievalRequest`（GET `/data-store-records`，每次僅帶 1 個 `fetch-correlation-id`；`200` 回資料、`204` 視為該 ID 無資料且不重試、`5xx/timeout` 有限重試）、`RetrievalUnsubscribe`（DELETE `/data-retrieval-subscriptions/{subscriptionId}`）
- [ ]  `internal/sbi/api_adrf_notify.go`（新增）：`POST /adrf/retrieval-notify` endpoint，接收 `NadrfDataRetrievalNotification`，取出 `fetchInstruct` 的 correlation IDs
- [ ]  `internal/sbi/server.go`：註冊 `/adrf/retrieval-notify` route
- [ ]  `internal/sbi/processor/processor.go`：新增 `HandleAdrfNotify`，委派至 MTLF
- [ ]  `internal/mtlf/`：retrain 觸發後，先按 SUPI list 逐一發 `RetrievalSubscribe`；收齊所有 IDs 後逐 ID `RetrievalRequest`；每筆從 `dataSub[0].smfDataSub.supi` 取 SUPI 查 group_id，POST `/upload_data` 給 Daisy；完成條件為「queued IDs 全部取完且收到 `terminationReq: true`」，之後呼叫 `RetrievalUnsubscribe`（`204` 成功、`404` 視為已清理、`5xx` 有限重試），再呼叫 `TriggerTraining`（帶 TID）
- [ ]  E2E 驗證：UPF 資料存入 ADRF → accuracy 下降觸發 retrain → MTLF 抓資料推 Daisy → Daisy 以 TID 訓練 → callback → 換版流程

---

### 彙整

| Phase | 負責方 | 前置條件 | 狀態 |
| --- | --- | --- | --- |
| E1 | Daisy | 無 | 待實作 |
| E2 | NWDAF | 無 | ✅ 完成 |
| E3 | NWDAF | E1、E2 | 待實作 |
