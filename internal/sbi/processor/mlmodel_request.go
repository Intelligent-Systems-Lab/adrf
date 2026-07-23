package processor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

type NadrfMLModelStoreRecord struct {
	NfInstanceId string        `json:"nfInstanceId,omitempty"`
	NfSetId      string        `json:"nfSetId,omitempty"`
	MlModelInfo  []MLModelInfo `json:"mlModelInfo,omitempty"`
	MlModels     []MLModel     `json:"mlModels,omitempty"`
	StoreResult  string        `json:"storeResult,omitempty"`
}

type MLModelInfo struct {
	ModelUniqueId string `json:"modelUniqueId" binding:"required"`
	MlFileAddr    string `json:"mlFileAddr" binding:"required"`
	MlStorageSize int64  `json:"mlStorageSize,omitempty"`
}

type MLModel struct {
	ModelUniqueId string `json:"modelUniqueId" binding:"required"`
	MlFileAddr    string `json:"mlFileAddr,omitempty"`
}

func (p *Processor) HandleCreateMLModelStoreRecord(c *gin.Context) {
	p.handleCreateMLModelStoreRecord(c)
}

func (p *Processor) HandleGetMLModelStoreRecords(c *gin.Context) {
	p.handleGetMLModelStoreRecords(c)
}

func (p *Processor) HandleGetIndividualMLModelStoreRecord(c *gin.Context) {
	p.handleGetIndividualMLModelStoreRecord(c)
}

func (p *Processor) HandleDownloadMLModelFile(c *gin.Context) {
	p.handleDownloadMLModelFile(c)
}

func (p *Processor) HandleUpdateIndividualMLModelStoreRecord(c *gin.Context) {
	p.handleUpdateIndividualMLModelStoreRecord(c)
}

func (p *Processor) HandleDeleteIndividualMLModelStoreRecord(c *gin.Context) {
	p.handleDeleteIndividualMLModelStoreRecord(c)
}

func validateMLModelStoreRecordPayload(req *NadrfMLModelStoreRecord) *models.ProblemDetails {
	if req == nil {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Request body is required",
			[]models.InvalidParam{
				invalidParam("/", "request body is required"),
			},
		)
	}

	invalidParams := make([]models.InvalidParam, 0, 4)

	hasNfInstanceId := strings.TrimSpace(req.NfInstanceId) != ""
	hasNfSetId := strings.TrimSpace(req.NfSetId) != ""

	if !hasNfInstanceId && !hasNfSetId {
		invalidParams = append(
			invalidParams,
			invalidParam("/nfInstanceId", "exactly one of nfInstanceId or nfSetId is required"),
			invalidParam("/nfSetId", "exactly one of nfInstanceId or nfSetId is required"),
		)
	} else if hasNfInstanceId && hasNfSetId {
		invalidParams = append(
			invalidParams,
			invalidParam("/nfInstanceId", "only one of nfInstanceId or nfSetId shall be provided"),
			invalidParam("/nfSetId", "only one of nfInstanceId or nfSetId shall be provided"),
		)
	}

	hasMlModelInfo := len(req.MlModelInfo) > 0
	hasMlModels := len(req.MlModels) > 0
	if !hasMlModelInfo && !hasMlModels {
		invalidParams = append(
			invalidParams,
			invalidParam("/mlModelInfo", "at least one of mlModelInfo or mlModels is required"),
		)
	}

	if len(invalidParams) > 0 {
		return newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"Mandatory IE is missing or invalid in NadrfMLModelStoreRecord payload",
			invalidParams,
		)
	}

	return nil
}

func (p *Processor) handleCreateMLModelStoreRecord(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"MLModelStore repository is unavailable",
			nil,
		))
		return
	}

	var req NadrfMLModelStoreRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		problem := mapJSONBindingErrorToProblemDetails(err)
		logger.ProcLog.Warnf("Failed to parse MLModelStore request: %v", err)
		p.writeProblem(c, problem)
		return
	}

	if problem := validateMLModelStoreRecordPayload(&req); problem != nil {
		logger.ProcLog.Warnf(
			"Invalid MLModelStore request payload: status=%d cause=%s detail=%s",
			problem.Status,
			problem.Cause,
			problem.Detail,
		)
		p.writeProblem(c, problem)
		return
	}

	// Currently supporting single model registration per request as per 3GPP TS 29.575 mapping.
	var modelUniqueId string
	var sourceMlFileAddr string

	if len(req.MlModelInfo) > 0 {
		modelUniqueId = req.MlModelInfo[0].ModelUniqueId
		sourceMlFileAddr = req.MlModelInfo[0].MlFileAddr
	} else if len(req.MlModels) > 0 {
		modelUniqueId = req.MlModels[0].ModelUniqueId
		sourceMlFileAddr = req.MlModels[0].MlFileAddr
	}

	storeTransId := store.NewStoreTransID()

	// Resolve local storage directory
	localDir := "./storage/models"
	if factory.AdrfConfig != nil && factory.AdrfConfig.Configuration != nil && factory.AdrfConfig.Configuration.MLModelStorage != nil {
		if dir := factory.AdrfConfig.Configuration.MLModelStorage.LocalDirectory; dir != "" {
			localDir = dir
		}
	}

	if err := os.MkdirAll(localDir, 0755); err != nil {
		logger.ProcLog.Errorf("Failed to create local model storage directory %q: %v", localDir, err)
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to initialize storage directory",
			nil,
		))
		return
	}

	destPath := filepath.Join(localDir, fmt.Sprintf("%s.tar.gz", storeTransId))
	logger.ProcLog.Infof("Downloading model binary from %s to %s", sourceMlFileAddr, destPath)

	var written int64
	var storeResult string
	var err error

	if sourceMlFileAddr != "" {
		written, err = downloadFile(sourceMlFileAddr, destPath)
		if err != nil {
			logger.ProcLog.Warnf("Failed to download model file from %s: %v", sourceMlFileAddr, err)
			storeResult = "ML_MODEL_FILE_DOWNLOAD_FAILED"
			req.StoreResult = storeResult
		} else {
			logger.ProcLog.Infof("Model binary stored successfully: size=%d bytes", written)
			storeResult = "ML_MODEL_FILE_STORED_IN_ADRF"
			req.StoreResult = storeResult
		}
	} else {
		storeResult = "ML_MODEL_FILE_STORED_IN_ADRF"
		req.StoreResult = storeResult
	}

	// Construct local ADRF download URL
	host := c.Request.Host
	if host == "" {
		host = "localhost:9888"
	}
	scheme := requestScheme(c.Request)
	adrfMlFileAddr := fmt.Sprintf("%s://%s%s%s/%s/model",
		scheme,
		host,
		factory.AdrfMLModelManagementResUriPrefix,
		factory.AdrfMLModelStoreRecordsPath,
		storeTransId,
	)

	var mlModelInfoDocs []store.MLModelInfoDoc
	if len(req.MlModelInfo) > 0 {
		mlModelInfoDocs = []store.MLModelInfoDoc{
			{
				ModelUniqueID: req.MlModelInfo[0].ModelUniqueId,
				MlFileAddr:    adrfMlFileAddr,
				MlStorageSize: written,
			},
		}
	}

	var mlModelDocs []store.MLModelDoc
	if len(req.MlModels) > 0 {
		mlModelDocs = []store.MLModelDoc{
			{
				ModelUniqueID: req.MlModels[0].ModelUniqueId,
				MlFileAddr:    adrfMlFileAddr,
			},
		}
	}

	doc := &store.MLModelStoreRecordDocument{
		StoreTransID:  storeTransId,
		NfInstanceID:  req.NfInstanceId,
		NfSetID:       req.NfSetId,
		MlModelInfo:   mlModelInfoDocs,
		MlModels:      mlModelDocs,
		ModelUniqueID: modelUniqueId,
		MlFileAddr:    adrfMlFileAddr,
		SourceAddr:    sourceMlFileAddr,
		StorageSize:   written,
		StoreResult:   storeResult,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	if err := p.mlModelRepo.InsertMLModelStoreRecord(dbCtx, doc); err != nil {
		logger.ProcLog.Errorf("Failed to insert MLModel record to DB: %v", err)
		if storeResult == "ML_MODEL_FILE_STORED_IN_ADRF" {
			_ = os.Remove(destPath)
		}
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to persist model record",
			nil,
		))
		return
	}

	if len(req.MlModelInfo) > 0 {
		req.MlModelInfo[0].MlFileAddr = adrfMlFileAddr
		req.MlModelInfo[0].MlStorageSize = doc.StorageSize
	} else if len(req.MlModels) > 0 {
		req.MlModels[0].MlFileAddr = adrfMlFileAddr
	}

	location := fmt.Sprintf("%s://%s%s%s/%s",
		scheme,
		host,
		factory.AdrfMLModelManagementResUriPrefix,
		factory.AdrfMLModelStoreRecordsPath,
		storeTransId,
	)
	c.Header("Location", location)
	c.JSON(http.StatusCreated, req)
}

func (p *Processor) handleGetMLModelStoreRecords(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"MLModelStore repository is unavailable",
			nil,
		))
		return
	}

	modelUniqueIds := c.QueryArray("model-unique-ids")

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	docs, err := p.mlModelRepo.GetMLModelStoreRecords(dbCtx, modelUniqueIds)
	if err != nil {
		logger.ProcLog.Errorf("Failed to query MLModel records: %v", err)
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to query model records",
			nil,
		))
		return
	}

	if len(docs) == 0 {
		c.Status(http.StatusNoContent)
		return
	}

	records := make([]NadrfMLModelStoreRecord, 0, len(docs))
	for _, doc := range docs {
		rec := NadrfMLModelStoreRecord{
			NfInstanceId: doc.NfInstanceID,
			NfSetId:      doc.NfSetID,
			StoreResult:  doc.StoreResult,
		}

		if len(doc.MlModelInfo) > 0 {
			rec.MlModelInfo = make([]MLModelInfo, 0, len(doc.MlModelInfo))
			for _, infoDoc := range doc.MlModelInfo {
				rec.MlModelInfo = append(rec.MlModelInfo, MLModelInfo{
					ModelUniqueId: infoDoc.ModelUniqueID,
					MlFileAddr:    infoDoc.MlFileAddr,
					MlStorageSize: infoDoc.MlStorageSize,
				})
			}
		} else {
			rec.MlModelInfo = []MLModelInfo{
				{
					ModelUniqueId: doc.ModelUniqueID,
					MlFileAddr:    doc.MlFileAddr,
					MlStorageSize: doc.StorageSize,
				},
			}
		}

		if len(doc.MlModels) > 0 {
			rec.MlModels = make([]MLModel, 0, len(doc.MlModels))
			for _, mDoc := range doc.MlModels {
				rec.MlModels = append(rec.MlModels, MLModel{
					ModelUniqueId: mDoc.ModelUniqueID,
					MlFileAddr:    mDoc.MlFileAddr,
				})
			}
		}

		records = append(records, rec)
	}

	c.JSON(http.StatusOK, records)
}

func (p *Processor) handleGetIndividualMLModelStoreRecord(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"MLModelStore repository is unavailable",
			nil,
		))
		return
	}

	storeTransId := c.Param("storeTransId")

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	doc, err := p.mlModelRepo.GetMLModelStoreRecord(dbCtx, storeTransId)
	if err != nil {
		logger.ProcLog.Errorf("Failed to retrieve MLModel record: %v", err)
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to query model record",
			nil,
		))
		return
	}

	if doc == nil {
		c.Status(http.StatusNotFound)
		return
	}

	record := NadrfMLModelStoreRecord{
		NfInstanceId: doc.NfInstanceID,
		NfSetId:      doc.NfSetID,
		StoreResult:  doc.StoreResult,
	}

	if len(doc.MlModelInfo) > 0 {
		record.MlModelInfo = make([]MLModelInfo, 0, len(doc.MlModelInfo))
		for _, infoDoc := range doc.MlModelInfo {
			record.MlModelInfo = append(record.MlModelInfo, MLModelInfo{
				ModelUniqueId: infoDoc.ModelUniqueID,
				MlFileAddr:    infoDoc.MlFileAddr,
				MlStorageSize: infoDoc.MlStorageSize,
			})
		}
	} else {
		record.MlModelInfo = []MLModelInfo{
			{
				ModelUniqueId: doc.ModelUniqueID,
				MlFileAddr:    doc.MlFileAddr,
				MlStorageSize: doc.StorageSize,
			},
		}
	}

	if len(doc.MlModels) > 0 {
		record.MlModels = make([]MLModel, 0, len(doc.MlModels))
		for _, mDoc := range doc.MlModels {
			record.MlModels = append(record.MlModels, MLModel{
				ModelUniqueId: mDoc.ModelUniqueID,
				MlFileAddr:    mDoc.MlFileAddr,
			})
		}
	}

	c.JSON(http.StatusOK, record)
}

func (p *Processor) handleDownloadMLModelFile(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		c.Status(http.StatusInternalServerError)
		return
	}

	storeTransId := c.Param("storeTransId")

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	doc, err := p.mlModelRepo.GetMLModelStoreRecord(dbCtx, storeTransId)
	if err != nil || doc == nil {
		c.Status(http.StatusNotFound)
		return
	}

	localDir := "./storage/models"
	if factory.AdrfConfig != nil && factory.AdrfConfig.Configuration != nil && factory.AdrfConfig.Configuration.MLModelStorage != nil {
		if dir := factory.AdrfConfig.Configuration.MLModelStorage.LocalDirectory; dir != "" {
			localDir = dir
		}
	}

	filePath := filepath.Join(localDir, fmt.Sprintf("%s.tar.gz", storeTransId))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.ProcLog.Warnf("Model file not found on disk at: %s", filePath)
		c.Status(http.StatusNotFound)
		return
	}

	c.Header("Content-Type", "application/gzip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.tar.gz\"", storeTransId))
	c.File(filePath)
}

func (p *Processor) handleUpdateIndividualMLModelStoreRecord(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"MLModelStore repository is unavailable",
			nil,
		))
		return
	}

	storeTransId := c.Param("storeTransId")

	var req NadrfMLModelStoreRecord
	if err := c.ShouldBindJSON(&req); err != nil {
		p.writeProblem(c, mapJSONBindingErrorToProblemDetails(err))
		return
	}

	if problem := validateMLModelStoreRecordPayload(&req); problem != nil {
		p.writeProblem(c, problem)
		return
	}

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	existing, err := p.mlModelRepo.GetMLModelStoreRecord(dbCtx, storeTransId)
	if err != nil {
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to query model record",
			nil,
		))
		return
	}

	if existing == nil {
		c.Status(http.StatusNotFound)
		return
	}

	existing.NfInstanceID = req.NfInstanceId
	existing.NfSetID = req.NfSetId

	if len(req.MlModelInfo) > 0 {
		info := req.MlModelInfo[0]
		existing.ModelUniqueID = info.ModelUniqueId
		existing.MlFileAddr = info.MlFileAddr
		existing.StorageSize = info.MlStorageSize
	} else if len(req.MlModels) > 0 {
		m := req.MlModels[0]
		existing.ModelUniqueID = m.ModelUniqueId
		if m.MlFileAddr != "" {
			existing.MlFileAddr = m.MlFileAddr
		}
	}

	if req.StoreResult != "" {
		existing.StoreResult = req.StoreResult
	}

	if err := p.mlModelRepo.UpdateMLModelStoreRecord(dbCtx, storeTransId, existing); err != nil {
		logger.ProcLog.Errorf("Failed to update MLModel record: %v", err)
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"SYSTEM_FAILURE",
			"Failed to update model record",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, req)
}

func (p *Processor) handleDeleteIndividualMLModelStoreRecord(c *gin.Context) {
	if p.mlModelRepo == nil {
		logger.ProcLog.Error("MLModelStore repository is not initialized")
		c.Status(http.StatusInternalServerError)
		return
	}

	storeTransId := c.Param("storeTransId")

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	// Verify existence and clean up disk file
	doc, err := p.mlModelRepo.GetMLModelStoreRecord(dbCtx, storeTransId)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
	if doc == nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Delete from Database
	if err := p.mlModelRepo.DeleteMLModelStoreRecord(dbCtx, storeTransId); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	// Delete from local disk
	localDir := "./storage/models"
	if factory.AdrfConfig != nil && factory.AdrfConfig.Configuration != nil && factory.AdrfConfig.Configuration.MLModelStorage != nil {
		if dir := factory.AdrfConfig.Configuration.MLModelStorage.LocalDirectory; dir != "" {
			localDir = dir
		}
	}
	filePath := filepath.Join(localDir, fmt.Sprintf("%s.tar.gz", storeTransId))
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.ProcLog.Warnf("Failed to delete model file %s from disk: %v", filePath, err)
	}

	c.Status(http.StatusNoContent)
}

func downloadFile(url string, destPath string) (int64, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("bad status code: %s", resp.Status)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return 0, err
	}

	return written, nil
}
