package service

import (
	"context"
	"io"
	"os"
	"runtime/debug"
	"sync"

	"github.com/sirupsen/logrus"

	adrf_context "github.com/free5gc/adrf/internal/context"
	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/internal/sbi"
	"github.com/free5gc/adrf/internal/sbi/consumer"
	"github.com/free5gc/adrf/internal/sbi/processor"
	"github.com/free5gc/adrf/pkg/app"
	"github.com/free5gc/adrf/pkg/factory"
)

var _ app.App = &AdrfApp{}

type AdrfApp struct {
	cfg       *factory.Config
	adrfCtx   *adrf_context.ADRFContext
	ctx       context.Context
	cancel    context.CancelFunc
	consumer  *consumer.Consumer
	processor *processor.Processor
	sbiServer *sbi.Server
	wg        sync.WaitGroup
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
	adrf.adrfCtx.AdrfName = cfg.GetAdrfName()

	var err error
	adrf.consumer, err = consumer.NewConsumer()
	if err != nil {
		return nil, err
	}

	adrf.processor = processor.NewProcessor()

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
