# ADRF Free5GC 對齊實作指南（V0）

## 1. 目的

本文件整理你現有的 free5gc-based `smf-nwdaf-ext`、`go-upf-ess/go-upf`、`NWDAF/NWDAF` 的共同寫法，做為 ADRF 實作基準。
目標是讓 ADRF 在程式骨架、風格、錯誤處理、配置、可測性上都盡量貼近 free5gc。

## 2. 參考來源（已閱讀）

1. `5G_Infrastructure/5GC/smf-nwdaf-ext`
2. `5G_Infrastructure/go-upf-ess/go-upf`
3. `5G_Infrastructure/NWDAF/NWDAF`

## 3. 必須對齊的骨架

### 3.1 入口層（`cmd/main.go`）

建議對齊模式：

1. 使用 `urfave/cli/v2`。
2. 保留 `--config/-c`、`--log/-l`。
3. `defer recover()` + `debug.Stack()`。
4. `action()` 內完成：
   - 讀 config
   - 初始化 app
   - 啟動 app
5. 全域 app 變數（與 SMF/NWDAF 一致）。

### 3.2 服務層（`pkg/service/init.go`）

建議 ADRF `App` 結構至少包含：

1. `cfg`
2. `ctx/cancel`
3. `wg`
4. `sbiServer`
5. `consumer`（即使 V0 可能不需，也保留擴充點）
6. `processor`
7. `context`（runtime state）

生命週期方法建議：

1. `NewApp(ctx, cfg)`
2. `Start()`
3. `Terminate()`
4. `listenShutdownEvent()`
5. `WaitRoutineStopped()`

### 3.3 SBI 層（`internal/sbi/server.go`）

建議對齊模式：

1. `Route` struct + `applyRoutes()`。
2. Router 用 gin，分群掛載 URI prefix。
3. `Run()` 啟 server goroutine；`Shutdown()/Stop()` 做 graceful shutdown。
4. route group 由常數驅動（不要硬編字串散落）。

## 4. ADRF 建議目錄（free5gc-aligned）

```text
adrf/
  cmd/main.go
  go.mod
  config/adrfcfg.yaml
  pkg/
    app/app.go
    service/init.go
    factory/config.go
    factory/factory.go
  internal/
    logger/logger.go
    context/context.go
    context/store_index.go
    sbi/
      server.go
      routes.go
      api_datamanagement.go
      processor/processor.go
      consumer/consumer.go
```

## 5. API 實作風格對齊點

### 5.1 回應與 header

1. `POST /data-store-records` 成功：`201 Created`。
2. `POST /data-retrieval-subscriptions` 成功：`201 Created`。
3. 成功建立資源時，回 `Location` header + response body（你們已定稿）。
4. callback 類通知 ACK：`204 No Content`。

### 5.2 錯誤模型

建議優先使用 `models.ProblemDetails`（free5gc 常見作法）：

1. JSON parse 錯：`400` + `Cause=INVALID_JSON`。
2. schema/語意錯：`400`。
3. 內部失敗：`500`。
4. 資源不存在：`404`（但 `RetrievalRequest` 無資料走 `204`，不是 `404`）。

### 5.3 路徑常數

建議集中於 `pkg/factory/config.go`（或常數檔）：

1. `AdrfDataManagementResUriPrefix = "/nadrf-datamanagement/v1"`
2. `AdrfDataStoreRecordsPath = "/data-store-records"`
3. `AdrfDataRetrievalSubscriptionsPath = "/data-retrieval-subscriptions"`

## 6. 狀態管理與一致性（從 SMF/UPF/NWDAF 抽出的共通原則）

### 6.1 先成功外部動作，再落地本地狀態

沿用 SMF event exposure 的關鍵原則：

1. 外部依賴失敗時，不先寫入半套本地 state。
2. 成功後再持久化 local store/index。

### 6.2 清理流程必須 deterministic

1. `DELETE` 時，即使下游通知清理失敗，本地清理仍要可收斂。
2. idempotency 要明確：重複刪除不應破壞狀態。

### 6.3 併發安全

建議參考 `subscription_store.go` 寫法：

1. `sync.RWMutex` 保護 maps。
2. ID 生成與資料 map 分離鎖。
3. 提供 `Create/Delete/Get/List` 清楚語意。

## 7. ADRF V0 功能行為（依你們定稿）

### 7.1 Store

1. 接收 `NadrfDataStoreRecord(dataSub + dataNotif)`。
2. 儲存時記錄 `ingestedAt`（ADRF internal metadata）。
3. 產生唯一 `storeTransId`。

### 7.2 Retrieval Subscribe + Fetch

1. `POST /data-retrieval-subscriptions` 建立訂閱時凍結 snapshot：
   - `notificationItems[*].startTime` 落在 `timePeriod`
   - 且 `ingestedAt <= T_sub`
2. callback 下發 `fetchInstruct.fetchCorrIds[]` 可多筆（由 ADRF 的 `corrIdBatchSize` 控制）。
3. NWDAF 端 fetch 固定逐 ID（每次 GET 僅 1 個 `fetch-correlation-id`）。

### 7.3 Retrieval Request 狀態碼

1. `200`：有資料。
2. `204`：該條件無資料（規格允許，且為成功分支）。
3. `4xx`（除 204）：請求/語意錯誤。
4. `5xx`：伺服端錯誤。

### 7.4 Unsubscribe

1. `DELETE /data-retrieval-subscriptions/{subscriptionId}`。
2. 成功 `204`。
3. 若已不存在，依你們 V0 策略可視為已清理。

## 8. Config 設計建議（`adrfcfg.yaml`）

建議最小結構：

```yaml
info:
  version: 1.0.0
  description: ADRF initial configuration

configuration:
  adrfName: ADRF
  sbi:
    scheme: http
    registerIPv4: 127.0.0.1
    bindingIPv4: 127.0.0.1
    port: 9888
  mongodb:
    name: adrf
    url: mongodb://127.0.0.1:27017
  retrieval:
    corrIdBatchSize: 100
    oneIdPerFetch: true
    snapshot:
      enabled: true

logger:
  enable: true
  level: info
  reportCaller: false
```

備註：

1. 你們已決定 Mongo 先本機小型實例，這份 config 與該決策一致。
2. `corrIdBatchSize` 屬 ADRF 端策略參數。

## 9. 測試與貢獻建議

### 9.1 測試層級

1. 單元測試：store/index/snapshot/filter/status-code。
2. API 測試：create/get/delete 基本流程。
3. E2E 測試：
   - NWDAF `POST /data-store-records`
   - retrain 前 `subscribe -> notify(fetch IDs) -> fetch -> unsubscribe`

### 9.2 測試工具風格

建議沿用 NWDAF 測試目錄風格：

1. `test/fake_*` 放 mock server。
2. `test/scripts/*.sh` 放 API smoke tests。

### 9.3 合併前檢查

1. `go test ./...`
2. `go fmt` / `gofmt` 統一格式。
3. API status code 與 `Location` header 行為和 spec 一致。
4. 文件同步更新（`docs/contract`、`docs/impl`）。

## 10. 實作策略建議（落地順序）

1. 先做 `StorageRequest + local Mongo persist + storeTransId index`。
2. 再做 `RetrievalSubscribe + snapshot freeze + callback(fetch IDs)`。
3. 再做 `RetrievalRequest`（先支援單 ID fetch）。
4. 最後做 `RetrievalUnsubscribe` 與清理。

這個順序最貼近你們目前 NWDAF E2 已完成、E3 待實作的整合節奏，也最容易先交付可驗證里程碑。
