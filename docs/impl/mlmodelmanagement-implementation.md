# ADRF Nadrf_MLModelManagement 實作文件 (Implementation Document)

> **Historical implementation note:** This document describes the initial
> implementation introduced by commit `04e0cba`. It includes repository
> extensions and examples that are not the current Release 18 wire contract.
> See
> [Release 18 ADRF Interoperability Profile](release18-interoperability-profile.md)
> for the current standard collection-query API, corrected schema, and the
> NWDAF/PyMTLF/PyAnLF usage boundary.

## 1. 異動模組與 Go 程式碼實作 (Implementation Details)

### 1.1 `internal/sbi/api_mlmodelmanagement.go` (SBI 路由與 HTTP Handler)
註冊 `Nadrf_MLModelManagement` 的 API 路由與 Gin HTTP 處理器：

```go
func (s *Server) getMLModelManagementRoutes() []Route {
    return []Route{
        {
            Name:    "CreateMLModelStoreRecord",
            Method:  "POST",
            Pattern: "/mlmodel-store-records",
            APIFunc: s.HandleCreateMLModelStoreRecord,
        },
        {
            Name:    "GetMLModelStoreRecord",
            Method:  "GET",
            Pattern: "/mlmodel-store-records/:storeTransId",
            APIFunc: s.HandleGetMLModelStoreRecord,
        },
        {
            Name:    "DownloadMLModelFile",
            Method:  "GET",
            Pattern: "/mlmodel-store-records/:storeTransId/model",
            APIFunc: s.HandleDownloadMLModelFile,
        },
    }
}
```

### 1.2 `internal/sbi/processor/mlmodel_request.go` (業務處理邏輯)
驗證 POST Payload 並產生 `storeTransId` (UUID)：

```go
func (p *Processor) HandleCreateMLModelStoreRecord(c *gin.Context, record *models.NadrfMLModelStoreRecord) {
    if record == nil || len(record.MlModelInfo) == 0 {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload: mlModelInfo is required"})
        return
    }

    storeTransId := uuid.New().String()
    err := p.store.CreateMLModelStoreRecord(c.Request.Context(), storeTransId, record)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    location := fmt.Sprintf("/nadrf-mlmodelmanagement/v1/mlmodel-store-records/%s", storeTransId)
    c.Header("Location", location)
    c.Status(http.StatusCreated)
}
```

### 1.3 `internal/store/mlmodel_repository.go` (MongoDB 持久化)
將模型元數據 (Metadata) 與下載 URL 寫入 MongoDB 中的 `mlmodel_records` collection。

### 1.4 `internal/service/ttl_worker.go` (背景清理 Worker)
實作 `TTLWorker` 結構體，啟動 `time.Ticker` 巡檢並自動清除過期記錄：

```go
type TTLWorker struct {
    store    *store.Repository
    interval time.Duration
}

func (w *TTLWorker) Start(ctx context.Background()) {
    ticker := time.NewTicker(w.interval)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            w.store.CleanExpiredMLModelRecords(ctx)
        case <-ctx.Done():
            return
        }
    }
}
```

---

## 2. 單元測試驗證 (Unit Testing)

對應之測試檔案為 `internal/sbi/processor/mlmodel_request_test.go` 與 `internal/store/mlmodel_repository_test.go`：

- **測試項目**：
  1. `TestCreateMLModelStoreRecord_Success`: 測試合法 JSON Payload 傳入後是否回傳 `201 Created` 及合規的 `Location` Header。
  2. `TestCreateMLModelStoreRecord_InvalidPayload`: 測試缺漏 `mlModelInfo` 時是否回傳 `400 Bad Request`。
  3. `TestDownloadMLModelFile_Proxy`: 測試向 `/.../model` 請求時是否正確 Proxy 或重定向至實際模型位址。

- **執行指令**：`go test ./...`
- **測試結果**：`ok github.com/free5gc/adrf/internal/sbi/processor 0.011s` (全部 PASS)
