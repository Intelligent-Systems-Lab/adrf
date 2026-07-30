package processor

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/adrf/internal/logger"
)

// handleDownloadDataSnapshot handles GET /nadrf-datamanagement/v1/data-snapshots/:snapshotId/download.
// Serves the exported snapshot JSON file to authorized consumers (e.g. MTLF-subp).
func (p *Processor) handleDownloadDataSnapshot(c *gin.Context) {
	snapshotID := strings.TrimSpace(c.Param("snapshotId"))
	if snapshotID == "" {
		p.writeProblem(c, newProblemDetails(
			http.StatusBadRequest,
			"MANDATORY_IE_MISSING",
			"snapshotId is required in path",
			nil,
		))
		return
	}

	filePath := filepath.Join("./storage/snapshots", fmt.Sprintf("%s.json", snapshotID))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logger.ProcLog.Warnf("Data snapshot download requested for non-existent file: %s", filePath)
		p.writeProblem(c, newProblemDetails(
			http.StatusNotFound,
			"RESOURCE_NOT_FOUND",
			"Requested snapshot dataset was not found",
			nil,
		))
		return
	}

	logger.ProcLog.Infof("Serving data snapshot dataset file for download: snapshotId=%s", snapshotID)
	c.Header("Content-Type", "application/json")
	c.File(filePath)
}
