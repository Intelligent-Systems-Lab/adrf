package sbi

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/sbi/processor"
	"github.com/free5gc/adrf/pkg/factory"
)

type adrfApp interface {
	Config() *factory.Config
	Processor() *processor.Processor
}

type Server struct {
	adrfApp

	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(adrf adrfApp) (*Server, error) {
	gin.SetMode(gin.ReleaseMode)

	s := &Server{
		adrfApp: adrf,
		router:  gin.New(),
	}

	s.router.Use(gin.Recovery())
	s.router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Output: logger.GinLog.WriterLevel(logrus.DebugLevel),
	}))

	dataMgmtRoutes := s.getDataManagementRoutes()
	dataMgmtGroup := s.router.Group(factory.AdrfDataManagementResUriPrefix)
	applyRoutes(dataMgmtGroup, dataMgmtRoutes)

	mlModelMgmtRoutes := s.getMLModelManagementRoutes()
	mlModelMgmtGroup := s.router.Group(factory.AdrfMLModelManagementResUriPrefix)
	applyRoutes(mlModelMgmtGroup, mlModelMgmtRoutes)


	cfg := adrf.Config()
	bindAddr := fmt.Sprintf("%s:%d", cfg.Configuration.Sbi.BindingIPv4, cfg.Configuration.Sbi.Port)
	logger.SBILog.Infof("binding addr: [%s]", bindAddr)

	s.httpServer = &http.Server{
		Addr:    bindAddr,
		Handler: s.router,
	}
	s.httpServer.ErrorLog = log.New(logger.SBILog.WriterLevel(logrus.ErrorLevel), "HTTP: ", 0)

	return s, nil
}

func (s *Server) getDataManagementRoutes() []Route {
	return []Route{
		{
			Name:    "CreateDataStoreRecord",
			Method:  "POST",
			Pattern: factory.AdrfDataStoreRecordsPath,
			APIFunc: s.HandleCreateDataStoreRecord,
		},
		{
			Name:    "GetDataStoreRecords",
			Method:  "GET",
			Pattern: factory.AdrfDataStoreRecordsPath,
			APIFunc: s.HandleGetDataStoreRecords,
		},
		{
			Name:    "SearchDataStoreRecords",
			Method:  "POST",
			Pattern: factory.AdrfDataStoreRecordsPath + "/search",
			APIFunc: s.HandleSearchDataStoreRecords,
		},
		{
			Name:    "CreateDataRetrievalSubscription",
			Method:  "POST",
			Pattern: factory.AdrfDataRetrievalSubscriptionsPath,
			APIFunc: s.HandleCreateDataRetrievalSubscription,
		},
		{
			Name:    "DeleteDataRetrievalSubscription",
			Method:  "DELETE",
			Pattern: factory.AdrfDataRetrievalSubscriptionsPath + "/:subscriptionId",
			APIFunc: s.HandleDeleteDataRetrievalSubscription,
		},
	}
}

func (s *Server) getMLModelManagementRoutes() []Route {
	return []Route{
		{
			Name:    "CreateMLModelStoreRecord",
			Method:  "POST",
			Pattern: factory.AdrfMLModelStoreRecordsPath,
			APIFunc: s.HandleCreateMLModelStoreRecord,
		},
		{
			Name:    "GetMLModelStoreRecords",
			Method:  "GET",
			Pattern: factory.AdrfMLModelStoreRecordsPath,
			APIFunc: s.HandleGetMLModelStoreRecords,
		},
		{
			Name:    "GetIndividualMLModelStoreRecord",
			Method:  "GET",
			Pattern: factory.AdrfMLModelStoreRecordsPath + "/:storeTransId",
			APIFunc: s.HandleGetIndividualMLModelStoreRecord,
		},
		{
			Name:    "DownloadMLModelFile",
			Method:  "GET",
			Pattern: factory.AdrfMLModelStoreRecordsPath + "/:storeTransId/model",
			APIFunc: s.HandleDownloadMLModelFile,
		},
		{
			Name:    "UpdateIndividualMLModelStoreRecord",
			Method:  "PUT",
			Pattern: factory.AdrfMLModelStoreRecordsPath + "/:storeTransId",
			APIFunc: s.HandleUpdateIndividualMLModelStoreRecord,
		},
		{
			Name:    "DeleteIndividualMLModelStoreRecord",
			Method:  "DELETE",
			Pattern: factory.AdrfMLModelStoreRecordsPath + "/:storeTransId",
			APIFunc: s.HandleDeleteIndividualMLModelStoreRecord,
		},
	}
}

func (s *Server) Run(traceCtx context.Context, wg *sync.WaitGroup) error {
	_ = traceCtx
	wg.Add(1)
	go s.startServer(wg)
	return nil
}

func (s *Server) Shutdown(traceCtx context.Context) {
	const defaultShutdownTimeout = 2 * time.Second

	if s.httpServer != nil {
		logger.SBILog.Infof("stop SBI server (listen on %s)", s.httpServer.Addr)
		toCtx, cancel := context.WithTimeout(traceCtx, defaultShutdownTimeout)
		defer cancel()
		if err := s.httpServer.Shutdown(toCtx); err != nil {
			logger.SBILog.Errorf("could not close SBI server: %#v", err)
		}
	}
}

func (s *Server) startServer(wg *sync.WaitGroup) {
	defer func() {
		if p := recover(); p != nil {
			logger.SBILog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
		wg.Done()
	}()

	logger.SBILog.Infof("start SBI server (listen on %s)", s.httpServer.Addr)

	cfg := s.Config()
	scheme := cfg.GetSbiScheme()

	var err error
	switch scheme {
	case "http":
		err = s.httpServer.ListenAndServe()
	case "https":
		err = fmt.Errorf("HTTPS not yet supported")
	default:
		err = fmt.Errorf("unsupported scheme: %s", scheme)
	}

	if err != nil && err != http.ErrServerClosed {
		logger.SBILog.Errorf("SBI server error: %v", err)
	}
	logger.SBILog.Infof("SBI server (listen on %s) stopped", s.httpServer.Addr)
}
