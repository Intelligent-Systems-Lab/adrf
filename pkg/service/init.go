package service

import (
	"context"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	adrf_context "github.com/free5gc/adrf/internal/context"
	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/sbi"
	"github.com/free5gc/adrf/internal/sbi/consumer"
	"github.com/free5gc/adrf/internal/sbi/processor"
	"github.com/free5gc/adrf/internal/service"
	"github.com/free5gc/adrf/internal/store"
	"github.com/free5gc/adrf/pkg/app"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/util/mongoapi"
)

var _ app.App = &AdrfApp{}

type AdrfApp struct {
	cfg          *factory.Config
	adrfCtx      *adrf_context.ADRFContext
	ctx          context.Context
	cancel       context.CancelFunc
	consumer     *consumer.Consumer
	processor    *processor.Processor
	dataStore    *store.DataStoreRepository
	mlModelStore *store.MLModelRepository
	sbiServer    *sbi.Server
	wg           sync.WaitGroup
}


func NewApp(ctx context.Context, cfg *factory.Config) (*AdrfApp, error) {
	adrf := &AdrfApp{
		cfg: cfg,
		wg:  sync.WaitGroup{},
	}

	if cfg.Logger != nil {
		adrf.SetLogEnable(cfg.Logger.Enable)
		adrf.SetLogLevel(cfg.Logger.Level)
		adrf.SetReportCaller(cfg.Logger.ReportCaller)
	}

	adrf.ctx, adrf.cancel = context.WithCancel(ctx)

	adrf_context.Init()
	adrf.adrfCtx = adrf_context.GetSelf()
	adrf.adrfCtx.InitFromConfig(cfg)

	if cfg.Configuration != nil && cfg.Configuration.Mongodb != nil {
		adrf.dataStore = store.NewDataStoreRepository(cfg.Configuration.Mongodb.Name)
		adrf.mlModelStore = store.NewMLModelRepository(cfg.Configuration.Mongodb.Name)
	}

	var err error
	adrf.consumer, err = consumer.NewConsumer()
	if err != nil {
		return nil, err
	}

	adrf.processor = processor.NewProcessor(adrf.dataStore)
	adrf.processor.SetMLModelRepo(adrf.mlModelStore)

	adrf.sbiServer, err = sbi.NewServer(adrf)
	if err != nil {
		return nil, err
	}

	return adrf, nil
}

func (a *AdrfApp) Config() *factory.Config {
	return a.cfg
}

func (a *AdrfApp) Context() *adrf_context.ADRFContext {
	return a.adrfCtx
}

func (a *AdrfApp) Processor() *processor.Processor {
	return a.processor
}

func (a *AdrfApp) Consumer() *consumer.Consumer {
	return a.consumer
}

func (a *AdrfApp) SetLogEnable(enable bool) {
	logger.MainLog.Infof("log enable is set to [%v]", enable)
	if enable {
		logger.Log.SetOutput(os.Stderr)
	} else {
		logger.Log.SetOutput(io.Discard)
	}
}

func (a *AdrfApp) SetLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		logger.MainLog.Warnf("log level [%s] is invalid", level)
		return
	}
	logger.MainLog.Infof("log level is set to [%s]", level)
	logger.Log.SetLevel(lvl)
}

func (a *AdrfApp) SetReportCaller(reportCaller bool) {
	logger.MainLog.Infof("report caller is set to [%v]", reportCaller)
	logger.Log.SetReportCaller(reportCaller)
}

func (a *AdrfApp) Start() {
	logger.InitLog.Infoln("ADRF server started")
	a.initMongoDataStore()

	// Start TTL Background Worker
	ttlWorker := service.NewTTLWorker(a.dataStore, a.mlModelStore, 60*time.Second)
	ttlWorker.Start(a.ctx)

	// Start NRF Registration Procedure & Heartbeat Loop
	consumer.RegisterADRFProcedure(a.ctx, a.consumer, a.adrfCtx)

	a.wg.Add(1)
	go a.listenShutdownEvent()

	if err := a.sbiServer.Run(context.Background(), &a.wg); err != nil {
		logger.InitLog.Fatalf("run SBI server failed: %+v", err)
	}

	a.WaitRoutineStopped()
}

func (a *AdrfApp) listenShutdownEvent() {
	defer func() {
		if p := recover(); p != nil {
			logger.InitLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
		a.wg.Done()
	}()

	<-a.ctx.Done()
	a.terminateProcedure()
}

func (a *AdrfApp) Terminate() {
	a.cancel()
}

func (a *AdrfApp) terminateProcedure() {
	logger.MainLog.Infof("terminating ADRF...")

	if a.sbiServer != nil {
		a.sbiServer.Shutdown(context.Background())
	}

	logger.InitLog.Infof("ADRF terminated")
}

func (a *AdrfApp) WaitRoutineStopped() {
	a.wg.Wait()
	logger.MainLog.Infof("ADRF app is terminated")
}

// initMongoDataStore initializes Mongo client connectivity and index bootstrap.
func (a *AdrfApp) initMongoDataStore() {
	logger.InitLog.Info("Initializing MongoDB data store bootstrap")

	if a.cfg == nil || a.cfg.Configuration == nil || a.cfg.Configuration.Mongodb == nil {
		logger.InitLog.Warn("MongoDB configuration is missing; skip data store bootstrap")
		return
	}

	mongodb := a.cfg.Configuration.Mongodb
	if mongodb.Name == "" || mongodb.Url == "" {
		logger.InitLog.Warn("MongoDB name/url is empty; skip data store bootstrap")
		return
	}

	logger.InitLog.Infof("Configuring MongoDB client: db=%s url=%s", mongodb.Name, mongodb.Url)
	if err := mongoapi.SetMongoDB(mongodb.Name, mongodb.Url); err != nil {
		logger.InitLog.Errorf("Failed to initialize MongoDB client: %v", err)
		return
	}

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := mongoapi.Client.Ping(pingCtx, nil); err != nil {
		logger.InitLog.Errorf("MongoDB ping failed (%s): %v", mongodb.Url, err)
		return
	}
	logger.InitLog.Infof("MongoDB connected: %s", mongodb.Url)

	if a.dataStore == nil {
		a.dataStore = store.NewDataStoreRepository(mongodb.Name)
		logger.InitLog.Infof("DataStore repository initialized: db=%s", mongodb.Name)
	}

	if a.mlModelStore == nil {
		a.mlModelStore = store.NewMLModelRepository(mongodb.Name)
		logger.InitLog.Infof("MLModelStore repository initialized: db=%s", mongodb.Name)
	}

	indexCtx, indexCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer indexCancel()
	if err := a.dataStore.EnsureIndexes(indexCtx); err != nil {
		logger.InitLog.Errorf("Failed to ensure MongoDB indexes: %v", err)
		return
	}
	if err := a.mlModelStore.EnsureIndexes(indexCtx); err != nil {
		logger.InitLog.Errorf("Failed to ensure MongoDB MLModel indexes: %v", err)
		return
	}

	logger.InitLog.Info("MongoDB data store index bootstrap completed")
}
