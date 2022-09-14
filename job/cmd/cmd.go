package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"g.hz.netease.com/horizon/core/config"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	"g.hz.netease.com/horizon/core/http/health"
	ginlogmiddle "g.hz.netease.com/horizon/core/middleware/ginlog"
	"g.hz.netease.com/horizon/job/autofree"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	eventhandlersvc "g.hz.netease.com/horizon/pkg/eventhandler"
	"g.hz.netease.com/horizon/pkg/eventhandler/wlgenerator"
	"g.hz.netease.com/horizon/pkg/grafana"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/util/kube"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	webhooksvc "g.hz.netease.com/horizon/pkg/webhook/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile  string
	Environment string
	LogLevel    string
}

// ParseFlags parses agent CLI flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(
		&flags.ConfigFile, "config", "", "configuration file path")

	flag.StringVar(
		&flags.Environment, "environment", "production", "environment string tag")

	flag.StringVar(
		&flags.LogLevel, "loglevel", "info", "the loglevel(panic/fatal/error/warn/info/debug/trace))")

	flag.Parse()
	return &flags
}

func InitLog(flags *Flags) {
	if flags.Environment == "production" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
	}
	logrus.SetOutput(os.Stdout)
	level, err := logrus.ParseLevel(flags.LogLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
}

// Run runs the agent.
func Run(flags *Flags) {
	// init log
	InitLog(flags)

	// load coreConfig
	coreConfig, err := config.LoadConfig(flags.ConfigFile)
	if err != nil {
		panic(err)
	}
	_, err = json.MarshalIndent(coreConfig, "", " ")
	if err != nil {
		panic(err)
	}

	// init db
	mysqlDB, err := orm.NewMySQLDB(&orm.MySQL{
		Host:              coreConfig.DBConfig.Host,
		Port:              coreConfig.DBConfig.Port,
		Username:          coreConfig.DBConfig.Username,
		Password:          coreConfig.DBConfig.Password,
		Database:          coreConfig.DBConfig.Database,
		PrometheusEnabled: coreConfig.DBConfig.PrometheusEnabled,
	})
	if err != nil {
		panic(err)
	}
	callbacks.RegisterCustomCallbacks(mysqlDB)

	// init manager parameter
	manager := managerparam.InitManager(mysqlDB)
	// init context
	ctx := context.Background()

	parameter := &param.Param{
		Manager: manager,
		Cd:      cd.NewCD(coreConfig.ArgoCDMapper),
	}

	// init controller
	var (
		clusterCtl     = clusterctl.NewController(&config.Config{}, parameter)
		prCtl          = prctl.NewController(parameter)
		environmentCtl = environmentctl.NewController(parameter)
	)

	// init kube client
	_, client, err := kube.BuildClient("/tmp/config_dev2")
	if err != nil {
		panic(err)
	}

	// sync grafana datasource periodically
	grafanaService := grafana.NewService(coreConfig.GrafanaConfig, manager, client)
	cancellableCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		grafanaService.SyncDatasource(cancellableCtx)
	}()

	log.Printf("auto-free job Config: %+v", coreConfig.AutoFreeConfig)
	// automatically release expired clusters
	go func() {
		autofree.AutoReleaseExpiredClusterJob(cancellableCtx, &coreConfig.AutoFreeConfig,
			parameter.UserManager, clusterCtl, prCtl, environmentCtl)
	}()

	// start event handler service to generate webhook log by events
	eventHandlerService := eventhandlersvc.NewService(ctx, manager)
	if err := eventHandlerService.RegisterEventHandler("webhook",
		wlgenerator.NewWebhookLogGenerator(manager)); err != nil {
		log.Printf("failed to register event handler, error: %s", err.Error())
	}
	eventHandlerService.Start()

	// start webhook service with multi workers to consume webhook logs and send webhook events
	webhookService := webhooksvc.NewService(ctx, manager)
	webhookService.Start()

	// graceful exit
	setTasksBeforeExit(webhookService.StopAndWait, eventHandlerService.StopAndWait)

	r := gin.New()
	// use middleware
	middlewares := []gin.HandlerFunc{
		ginlogmiddle.Middleware(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	}
	r.Use(middlewares...)

	gin.ForceConsoleColor()

	health.RegisterRoutes(r)

	log.Printf("Server started")
	log.Print(r.Run(fmt.Sprintf(":%d", coreConfig.ServerConfig.Port)))
}

// setTasksBeforeExit set stop funcs which will be executed after sigterm and sigint catched
func setTasksBeforeExit(stopFuncs ...func()) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sig
		log.Printf("got %s signal, stop tasks...\n", s)
		if len(stopFuncs) == 0 {
			return
		}
		wg := sync.WaitGroup{}
		wg.Add(len(stopFuncs))
		for _, stopFunc := range stopFuncs {
			go func(stop func()) {
				stop()
			}(stopFunc)
		}
		wg.Wait()
		log.Printf("all tasks stopped, exit now.")
	}()
}
