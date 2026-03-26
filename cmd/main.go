package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/free5gc/adrf/internal/logger"
	"github.com/free5gc/adrf/pkg/factory"
	"github.com/free5gc/adrf/pkg/service"
	logger_util "github.com/free5gc/util/logger"
)

var ADRF *service.AdrfApp

func main() {
	defer func() {
		if p := recover(); p != nil {
			logger.MainLog.Fatalf("panic: %v\n%s", p, string(debug.Stack()))
		}
	}()

	app := cli.NewApp()
	app.Name = "adrf"
	app.Usage = "5G Analytics Data Repository Function (ADRF)"
	app.Action = action
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "config",
			Aliases: []string{"c"},
			Usage:   "Load configuration from `FILE`",
		},
		&cli.StringSliceFlag{
			Name:    "log",
			Aliases: []string{"l"},
			Usage:   "Output NF log to `FILE`",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Printf("ADRF run error: %v\n", err)
	}
}

func action(cliCtx *cli.Context) error {
	if err := initLogFile(cliCtx.StringSlice("log")); err != nil {
		return err
	}

	logger.MainLog.Infoln("ADRF")
	logger.MainLog.Infoln("ADRF version: v0.1.0")

	ctx, cancel := context.WithCancel(context.Background())
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	cfg, err := factory.ReadConfig(cliCtx.String("config"))
	if err != nil {
		return err
	}
	factory.AdrfConfig = cfg

	adrf, err := service.NewApp(ctx, cfg)
	if err != nil {
		return err
	}
	ADRF = adrf

	adrf.Start()
	return nil
}

func initLogFile(logNfPath []string) error {
	for _, path := range logNfPath {
		if err := logger_util.LogFileHook(logger.Log, path); err != nil {
			return err
		}
	}
	return nil
}
