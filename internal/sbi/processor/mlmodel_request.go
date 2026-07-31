package processor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/openapi/models"
)

type NadrfMLModelStoreRecord struct {
	NfInstanceId     string            `json:"nfInstanceId,omitempty"`
	NfSetId          string            `json:"nfSetId,omitempty"`
	MlModelInfo      []MLModelInfo     `json:"mlModelInfo,omitempty"`
	MlModels         []MLModel         `json:"mlModels,omitempty"`
	ModelStoreResult *ModelStoreResult `json:"modelStoreResult,omitempty"`
	SuppFeat         string            `json:"suppFeat,omitempty"`
}

type MLModelInfo struct {
	ModelUniqueId     *int64            `json:"modelUniqueId"`
	MlFileAddr        *MLModelAddress   `json:"mlFileAddr"`
	MlStorageSize     *int64            `json:"mlStorageSize"`
	AllowConsumerList []AllowedConsumer `json:"allowConsumerList,omitempty"`
}

type MLModelAddress struct {
	MLModelURL string `json:"mLModelUrl,omitempty"`
	MLFileFQDN string `json:"mlFileFqdn,omitempty"`
}

type AllowedConsumer struct {
	NfInstanceId string `json:"nfInstanceId,omitempty"`
	NfSetId      string `json:"nfSetId,omitempty"`
}

type MLModel struct {
	ModelUniqueId *int64 `json:"modelUniqueId"`
	MlModel       []byte `json:"mlModel"`
}

type ModelStoreResult struct {
	ModelUniqueId *int64 `json:"modelUniqueId"`
	StoreResult   string `json:"storeResult"`
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
	if hasMlModelInfo == hasMlModels {
		invalidParams = append(
			invalidParams,
			invalidParam("/mlModelInfo", "exactly one of mlModelInfo or mlModels is required"),
		)
	}
	if hasMlModels {
		invalidParams = append(
			invalidParams,
			invalidParam("/mlModels", "inline ML model storage is not supported by this deployment"),
		)
	}
	if hasMlModelInfo && len(req.MlModelInfo) != 1 {
		invalidParams = append(
			invalidParams,
			invalidParam("/mlModelInfo", "exactly one URL-backed ML model is supported per request"),
		)
	}
	for index, info := range req.MlModelInfo {
		path := fmt.Sprintf("/mlModelInfo/%d", index)
		if info.ModelUniqueId == nil || *info.ModelUniqueId < 0 {
			invalidParams = append(invalidParams, invalidParam(path+"/modelUniqueId", "a non-negative integer is required"))
		}
		if info.MlStorageSize == nil || *info.MlStorageSize < 0 {
			invalidParams = append(invalidParams, invalidParam(path+"/mlStorageSize", "a non-negative integer is required"))
		}
		if info.MlFileAddr == nil ||
			(strings.TrimSpace(info.MlFileAddr.MLModelURL) == "") ==
				(strings.TrimSpace(info.MlFileAddr.MLFileFQDN) == "") {
			invalidParams = append(invalidParams, invalidParam(path+"/mlFileAddr", "exactly one of mLModelUrl or mlFileFqdn is required"))
		}
		if info.MlFileAddr != nil && strings.TrimSpace(info.MlFileAddr.MLFileFQDN) != "" {
			invalidParams = append(invalidParams, invalidParam(path+"/mlFileAddr/mlFileFqdn", "FQDN-backed storage is not supported by this deployment"))
		}
		for consumerIndex, consumer := range info.AllowConsumerList {
			hasInstance := strings.TrimSpace(consumer.NfInstanceId) != ""
			hasSet := strings.TrimSpace(consumer.NfSetId) != ""
			if hasInstance == hasSet {
				invalidParams = append(invalidParams, invalidParam(
					fmt.Sprintf("%s/allowConsumerList/%d", path, consumerIndex),
					"exactly one of nfInstanceId or nfSetId is required",
				))
			}
		}
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
	modelUniqueId := *req.MlModelInfo[0].ModelUniqueId
	sourceMlFileAddr := req.MlModelInfo[0].MlFileAddr.MLModelURL

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
	var err error

	if sourceMlFileAddr != "" {
		written, err = downloadFile(sourceMlFileAddr, destPath)
		if err != nil {
			logger.ProcLog.Warnf("Failed to download model file from %s: %v", sourceMlFileAddr, err)
			_ = os.Remove(destPath)
			status := http.StatusInternalServerError
			cause := "ML_MODEL_FILE_DOWNLOAD_FAILED"
			var sourceError *modelSourceHTTPError
			if errors.As(err, &sourceError) && sourceError.statusCode == http.StatusNotFound {
				status = http.StatusNotFound
				cause = "ML_MODEL_FILE_ADDRESS_NOT_FOUND"
			}
			p.writeProblem(c, newProblemDetails(
				status,
				cause,
				"ADRF could not retrieve the supplied ML model file",
				nil,
			))
			return
		}
		logger.ProcLog.Infof("Model binary stored successfully: size=%d bytes", written)
	}
	if expected := *req.MlModelInfo[0].MlStorageSize; written != expected {
		_ = os.Remove(destPath)
		logger.ProcLog.Warnf(
			"Downloaded model size mismatch: expected=%d actual=%d source=%s",
			expected,
			written,
			sourceMlFileAddr,
		)
		p.writeProblem(c, newProblemDetails(
			http.StatusInternalServerError,
			"ML_MODEL_FILE_DOWNLOAD_FAILED",
			"Downloaded ML model size does not match mlStorageSize",
			nil,
		))
		return
	}
	const storeResult = "ML_MODEL_FILE_STORED_IN_ADRF"

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

	allowedConsumers := make([]store.AllowedConsumerDoc, 0, len(req.MlModelInfo[0].AllowConsumerList))
	for _, consumer := range req.MlModelInfo[0].AllowConsumerList {
		allowedConsumers = append(allowedConsumers, store.AllowedConsumerDoc{
			NfInstanceID: consumer.NfInstanceId,
			NfSetID:      consumer.NfSetId,
		})
	}
	mlModelInfoDocs := []store.MLModelInfoDoc{{
		ModelUniqueID:     modelUniqueId,
		MlFileAddr:        adrfMlFileAddr,
		MlStorageSize:     written,
		AllowConsumerList: allowedConsumers,
	}}

	doc := &store.MLModelStoreRecordDocument{
		StoreTransID:  storeTransId,
		NfInstanceID:  req.NfInstanceId,
		NfSetID:       req.NfSetId,
		MlModelInfo:   mlModelInfoDocs,
		MlModels:      nil,
		ModelUniqueID: modelUniqueId,
		MlFileAddr:    adrfMlFileAddr,
		SourceAddr:    sourceMlFileAddr,
		StorageSize:   written,
		ModelStoreResult: store.ModelStoreResultDoc{
			ModelUniqueID: modelUniqueId,
			StoreResult:   storeResult,
		},
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
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

	req.MlModelInfo[0].MlFileAddr = &MLModelAddress{MLModelURL: adrfMlFileAddr}
	req.MlModelInfo[0].MlStorageSize = &doc.StorageSize
	req.ModelStoreResult = &ModelStoreResult{
		ModelUniqueId: &doc.ModelUniqueID,
		StoreResult:   storeResult,
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

	storeTransId := strings.TrimSpace(c.Query("store-trans-id"))
	rawModelUniqueIds, hasModelUniqueIds := c.Request.URL.Query()["model-unique-ids"]
	if (storeTransId != "") == hasModelUniqueIds {
		p.writeProblem(c, newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_QUERY_PARAM_INCORRECT",
			"exactly one of store-trans-id or model-unique-ids is required",
			nil,
		))
		return
	}

	dbCtx, dbCancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer dbCancel()

	var (
		doc *store.MLModelStoreRecordDocument
		err error
	)
	if storeTransId != "" {
		doc, err = p.mlModelRepo.GetMLModelStoreRecord(dbCtx, storeTransId)
	} else {
		modelUniqueIds, parseErr := parseModelUniqueIDs(rawModelUniqueIds)
		if parseErr != nil {
			p.writeProblem(c, newProblemDetails(
				http.StatusBadRequest,
				"MANDATORY_QUERY_PARAM_INCORRECT",
				parseErr.Error(),
				nil,
			))
			return
		}
		var docs []*store.MLModelStoreRecordDocument
		docs, err = p.mlModelRepo.GetMLModelStoreRecords(dbCtx, modelUniqueIds)
		if len(docs) == 1 {
			doc = docs[0]
		} else if len(docs) > 1 {
			p.writeProblem(c, newProblemDetails(
				http.StatusInternalServerError,
				"SYSTEM_FAILURE",
				"model-unique-ids did not resolve to one unambiguous store record",
				nil,
			))
			return
		}
	}
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

	if doc == nil {
		c.Status(http.StatusNoContent)
		return
	}

	c.JSON(http.StatusOK, mlModelStoreRecordFromDocument(doc))
}

func parseModelUniqueIDs(rawValues []string) ([]int64, error) {
	if len(rawValues) == 0 {
		return nil, fmt.Errorf("model-unique-ids must contain at least one identifier")
	}
	ids := make([]int64, 0, len(rawValues))
	for _, raw := range rawValues {
		for _, part := range strings.Split(raw, ",") {
			value, err := strconv.ParseInt(strings.TrimSpace(part), 10, 64)
			if err != nil || value < 0 {
				return nil, fmt.Errorf("model-unique-ids must contain non-negative integers")
			}
			ids = append(ids, value)
		}
	}
	return ids, nil
}

func mlModelStoreRecordFromDocument(doc *store.MLModelStoreRecordDocument) NadrfMLModelStoreRecord {
	record := NadrfMLModelStoreRecord{
		NfInstanceId: doc.NfInstanceID,
		NfSetId:      doc.NfSetID,
		ModelStoreResult: &ModelStoreResult{
			ModelUniqueId: &doc.ModelStoreResult.ModelUniqueID,
			StoreResult:   doc.ModelStoreResult.StoreResult,
		},
	}
	for _, infoDoc := range doc.MlModelInfo {
		allowedConsumers := make([]AllowedConsumer, 0, len(infoDoc.AllowConsumerList))
		for _, consumer := range infoDoc.AllowConsumerList {
			allowedConsumers = append(allowedConsumers, AllowedConsumer{
				NfInstanceId: consumer.NfInstanceID,
				NfSetId:      consumer.NfSetID,
			})
		}
		info := infoDoc
		record.MlModelInfo = append(record.MlModelInfo, MLModelInfo{
			ModelUniqueId:     &info.ModelUniqueID,
			MlFileAddr:        &MLModelAddress{MLModelURL: info.MlFileAddr},
			MlStorageSize:     &info.MlStorageSize,
			AllowConsumerList: allowedConsumers,
		})
	}
	return record
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

	c.JSON(http.StatusOK, mlModelStoreRecordFromDocument(doc))
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
		existing.ModelUniqueID = *info.ModelUniqueId
		existing.SourceAddr = info.MlFileAddr.MLModelURL
		existing.StorageSize = *info.MlStorageSize
		allowedConsumers := make([]store.AllowedConsumerDoc, 0, len(info.AllowConsumerList))
		for _, consumer := range info.AllowConsumerList {
			allowedConsumers = append(allowedConsumers, store.AllowedConsumerDoc{
				NfInstanceID: consumer.NfInstanceId,
				NfSetID:      consumer.NfSetId,
			})
		}
		existing.MlModelInfo = []store.MLModelInfoDoc{{
			ModelUniqueID:     existing.ModelUniqueID,
			MlFileAddr:        existing.MlFileAddr,
			MlStorageSize:     existing.StorageSize,
			AllowConsumerList: allowedConsumers,
		}}
	}

	if req.ModelStoreResult != nil {
		existing.ModelStoreResult = store.ModelStoreResultDoc{
			ModelUniqueID: *req.ModelStoreResult.ModelUniqueId,
			StoreResult:   req.ModelStoreResult.StoreResult,
		}
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

	c.JSON(http.StatusOK, mlModelStoreRecordFromDocument(existing))
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

type modelSourceHTTPError struct {
	statusCode int
	status     string
}

func (e *modelSourceHTTPError) Error() string {
	return fmt.Sprintf("source returned %s", e.status)
}

func downloadFile(url string, destPath string) (int64, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, &modelSourceHTTPError{
			statusCode: resp.StatusCode,
			status:     resp.Status,
		}
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
