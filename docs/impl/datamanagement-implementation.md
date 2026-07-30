# ADRF Nadrf_DataManagement 實作規範與細節文件

## 1. 服務介紹 (Overview)

本文件記載 `adrf` 模組中關於 3GPP TS 29.575 **`Nadrf_DataManagement`** 服務項目的實作細節，包含資料儲存紀錄派發、快照匯出、`fetchUri` 下載端點及訂閱清理等。

---

## 2. 核心 API 流程與程式碼對照

### 2.1 數據紀錄儲存 (`POST /nadrf-datamanagement/v1/data-store-records`)

- **對應檔案**：`internal/sbi/processor/data_store.go`
- **邏輯**：接收 NWDAF 或 UPF 收集之歷史流量數據，將紀錄寫入 MongoDB (`adrf.data_store_records` collection) 或記憶體快取。
- **標頭與回應**：傳回 `HTTP 201 Created` 並附加 `Location` 標頭 (例如 `/nadrf-datamanagement/v1/data-store-records/{recordId}`)。

---

### 2.2 歷史快照匯出與通知 (`POST /nadrf-datamanagement/v1/data-retrieval-subscriptions`)

- **對應檔案**：
  - `internal/sbi/processor/retrieval_subscription_create.go`
  - `internal/sbi/processor/snapshot_download_request.go`
- **邏輯**：
  1. 接收 NWDAF / MTLF 之歷史資料檢索訂閱。
  2. 根據查詢條件檢索紀錄，並匯出為本地 JSON 快照檔案：`./storage/snapshots/{subscriptionId}.json`。
  3. 構造 3GPP TS 29.575 OpenAPI 3.0 標準之 `NadrfDataRetrievalNotification` 物件：
     ```go
     notif := models.NadrfDataRetrievalNotification{
         NotifCorrId: subscriptionId,
         TimeStamp:   &now,
         FetchInstruct: &models.FetchInstruction{
             FetchUri: fmt.Sprintf("http://192.168.107.5:9888/nadrf-datamanagement/v1/data-snapshots/%s/download", subscriptionId),
         },
     }
     ```
  4. 非同步 POST 通知訂閱端 (如 `MTLF-subp` 的 `/api/v1/mtlf/adrf-callback`)。
  5. 回傳 `HTTP 201 Created` 與 `Location` 標頭。

---

### 2.3 快照下載端點 (`GET /nadrf-datamanagement/v1/data-snapshots/{subscriptionId}/download`)

- **對應檔案**：`internal/sbi/processor/snapshot_download_request.go`
- **邏輯**：
  提供 `fetchInstruct.fetchUri` 所指向的 RESTful HTTP GET 下載端點，將指定的快照 JSON 檔案輸出傳送給消費端，徹底解決跨 VM 或容器間共享磁碟掛載之依賴問題。

```go
func HandleSnapshotDownloadRequest(c *gin.Context, subId string) {
    filePath := filepath.Join("./storage/snapshots", subId+".json")
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        c.JSON(http.StatusNotFound, gin.H{"error": "snapshot not found"})
        return
    }
    c.FileAttachment(filePath, subId+".json")
}
```

---

### 2.4 訂閱釋放與快照清理 (`DELETE /nadrf-datamanagement/v1/data-retrieval-subscriptions/{subscriptionId}`)

- **對應檔案**：`internal/sbi/processor/retrieval_subscription_delete.go`
- **邏輯**：
  1. 消費端 (如 `MTLF-subp`) 在完成快照下載與重訓啟動後，呼叫此刪除端點 (`RetrievalUnsubscribe`)。
  2. ADRF 刪除記憶體與數據庫中的訂閱紀錄。
  3. **自動清理實體檔案**：主動刪除對應的快照 JSON 檔案 (`./storage/snapshots/{subscriptionId}.json`)，防止虛擬機磁碟空間暴增。
  4. 回傳 **`HTTP 204 No Content`**。

---

## 3. 單元測試與驗證 (Testing & Verification)

- **單元測試路徑**：`internal/sbi/processor/`
- **執行指令**：`go test -v -short ./internal/sbi/processor/...`
- **驗證結果**：**100% PASSED** (包含 `HandleSnapshotDownloadRequest` 與 `HandleRetrievalSubscriptionDelete` 之功能與邊界測試)。
