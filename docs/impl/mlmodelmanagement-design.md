# ADRF Nadrf_MLModelManagement 設計文件 (Design Document)

## 1. 服務定位與 3GPP TS 29.575 架構 (Service Architecture)

Analytics Data Repository Function (ADRF) 增加了 3GPP TS 29.575 標準規定的 **`Nadrf_MLModelManagement`** 服務，與原有的 `Nadrf_DataManagement` 並列，實現 5GC 中 ML 模型訓練產物的正式持久化、版控與分發。

```
+---------------------------------------------------------------------------------+
|                               5GC NWDAF / MTLF                                  |
+---------------------------------------------------------------------------------+
                                         |
     POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records (儲存/註冊模型)
                                         v
+---------------------------------------------------------------------------------+
|                                 5GC ADRF                                        |
|  - Nadrf_DataManagement      (UE 流量與歷史數據儲存)                             |
|  - Nadrf_MLModelManagement   (ML 模型檔案儲存與 TTL 背景清理)                     |
+---------------------------------------------------------------------------------+
```

---

## 2. 核心 REST 介面與資料模型設計 (API & Schema Design)

### 2.1 儲存記錄介面 (Model Store Records API)
- **Endpoint**: `POST /nadrf-mlmodelmanagement/v1/mlmodel-store-records`
- **Request Body (NadrfMLModelStoreRecord)**:
  - `nfInstanceId`: 註冊來源 NF 的識別碼 (如 `nwdaf-sbi-gateway`)
  - `mlModelInfo`: 陣列結構，含 `modelUniqueId` (模型 UUID) 與 `mlFileAddr` (原始 Staging URL)
- **Response**:
  - `HTTP 201 Created`
  - `Location Header`: `/nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}`
  - **正式下載位址格式**：`Location + "/model"` (如 `http://192.168.107.5:9888/nadrf-mlmodelmanagement/v1/mlmodel-store-records/{storeTransId}/model`)

### 2.2 背景生命週期管理 (TTL Worker Design)
- **設計目標**：為防止未續期或陳舊的模型檔案佔滿 ADRF 磁碟與 MongoDB，設計獨立的背景 TTL Worker。
- **機制**：
  - 定期掃描 `mlmodel_store_records` 資料庫集合。
  - 對過期（超出保留時間）的模型記錄與暫存檔案執行自動清理 (Garbage Collection)。
