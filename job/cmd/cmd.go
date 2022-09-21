package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/config"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	userctl "g.hz.netease.com/horizon/core/controller/user"
	"g.hz.netease.com/horizon/core/http/health"
	ginlogmiddle "g.hz.netease.com/horizon/core/middleware/ginlog"
	"g.hz.netease.com/horizon/job/autofree"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clusterservice "g.hz.netease.com/horizon/pkg/cluster/service"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cmdb"
	oauthconfig "g.hz.netease.com/horizon/pkg/config/oauth"
	roleconfig "g.hz.netease.com/horizon/pkg/config/role"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/grafana"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook"
	"g.hz.netease.com/horizon/pkg/hook/handler"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/scope"
	oauthstore "g.hz.netease.com/horizon/pkg/oauth/store"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschemarepo "g.hz.netease.com/horizon/pkg/templaterelease/schema/repo"
	templaterepoharbor "g.hz.netease.com/horizon/pkg/templaterepo/harbor"
	userservice "g.hz.netease.com/horizon/pkg/user/service"
	"g.hz.netease.com/horizon/pkg/util/kube"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile         string
	RoleConfigFile     string
	ScopeRoleFile      string
	Environment        string
	LogLevel           string
	AutoReleaseAccount string
}

// ParseFlags parses agent CLI flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(
		&flags.ConfigFile, "config", "", "configuration file path")

	flag.StringVar(
		&flags.RoleConfigFile, "roles", "", "roles file path")

	flag.StringVar(
		&flags.ScopeRoleFile, "scopes", "", "configuration file path")

	flag.StringVar(
		&flags.Environment, "environment", "production", "environment string tag")

	flag.StringVar(
		&flags.LogLevel, "loglevel", "info", "the loglevel(panic/fatal/error/warn/info/debug/trace))")

	flag.StringVar(
		&flags.AutoReleaseAccount, "autoreleaseaccount", "", "auto release cluster account")

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

	// init roles
	file, err := os.OpenFile(flags.RoleConfigFile, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	content, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var roleConfig roleconfig.Config
	if err := yaml.Unmarshal(content, &roleConfig); err != nil {
		panic(err)
	} else {
		log.Printf("the roleConfig = %+v\n", roleConfig)
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

	// init manager parameter
	manager := managerparam.InitManager(mysqlDB)

	// init context
	ctx := context.Background()

	// init service
	gitlabFactory := gitlabfty.NewFactory(coreConfig.GitlabMapper)
	gitlabCompute, err := gitlabFactory.GetByName(ctx, common.GitlabCompute)
	if err != nil {
		panic(err)
	}
	gitlabControl, err := gitlabFactory.GetByName(ctx, common.GitlabControl)
	if err != nil {
		panic(err)
	}
	applicationGitRepo, err := gitrepo.NewApplicationGitlabRepo(ctx, coreConfig.GitlabRepoConfig, gitlabCompute)
	if err != nil {
		panic(err)
	}
	templateRepo, err := templaterepoharbor.NewTemplateRepo(coreConfig.TemplateRepo)
	if err != nil {
		panic(err)
	}
	clusterGitRepo, err := clustergitrepo.NewClusterGitlabRepo(ctx, coreConfig.GitlabRepoConfig,
		templateRepo, gitlabCompute)
	if err != nil {
		panic(err)
	}
	templateSchemaGetter := templateschemarepo.NewSchemaGetter(ctx, templateRepo, manager)
	outputGetter, err := output.NewOutPutGetter(ctx, templateRepo, manager)
	if err != nil {
		panic(err)
	}
	tektonFty, err := factory.NewFactory(coreConfig.TektonMapper)
	if err != nil {
		panic(err)
	}
	gitGetter, err := code.NewGitGetter(ctx, gitlabControl)
	if err != nil {
		panic(err)
	}

	handlers := make([]hook.EventHandler, 0)
	if coreConfig.CmdbConfig.Enabled {
		cmdbController := cmdb.NewController(coreConfig.CmdbConfig)
		cmdbHandler := handler.NewCMDBEventHandler(cmdbController)
		handlers = append(handlers, cmdbHandler)
	}
	memHook := hook.NewInMemHook(2000, handlers...)
	oauthAppStore := oauthstore.NewOauthAppStore(mysqlDB)
	oauthTokenStore := oauthstore.NewTokenStore(mysqlDB)
	oauthManager := oauthmanager.NewManager(oauthAppStore, oauthTokenStore,
		generate.NewAuthorizeGenerate(), coreConfig.Oauth.AuthorizeCodeExpireIn, coreConfig.Oauth.AccessTokenExpireIn)
	roleService, err := role.NewFileRoleFrom2(context.TODO(), roleConfig)
	if err != nil {
		panic(err)
	}
	mservice := memberservice.NewService(roleService, oauthManager, manager)

	// init scope service
	scopeFile, err := os.OpenFile(flags.ScopeRoleFile, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	content, err = ioutil.ReadAll(scopeFile)
	if err != nil {
		panic(err)
	}
	var oauthConfig oauthconfig.Scopes
	if err = yaml.Unmarshal(content, &oauthConfig); err != nil {
		panic(err)
	} else {
		log.Printf("the oauthScopeConfig = %+v\n", oauthConfig)
	}
	scopeService, err := scope.NewFileScopeService(oauthConfig)
	if err != nil {
		panic(err)
	}

	groupSvc := groupservice.NewService(manager)
	applicationSvc := applicationservice.NewService(groupSvc, manager)
	clusterSvc := clusterservice.NewService(applicationSvc, manager)
	userSvc := userservice.NewService(manager)

	parameter := &param.Param{
		Manager:              manager,
		OauthManager:         oauthManager,
		MemberService:        mservice,
		ApplicationSvc:       applicationSvc,
		ClusterSvc:           clusterSvc,
		GroupSvc:             groupSvc,
		UserSvc:              userSvc,
		RoleService:          roleService,
		ScopeService:         scopeService,
		Hook:                 memHook,
		ApplicationGitRepo:   applicationGitRepo,
		TemplateSchemaGetter: templateSchemaGetter,
		Cd:                   cd.NewCD(coreConfig.ArgoCDMapper),
		OutputGetter:         outputGetter,
		TektonFty:            tektonFty,
		ClusterGitRepo:       clusterGitRepo,
		GitGetter:            gitGetter,
	}

	// init controller
	var (
		userCtl        = userctl.NewController(parameter)
		clusterCtl     = clusterctl.NewController(coreConfig, parameter)
		prCtl          = prctl.NewController(parameter)
		environmentCtl = environmentctl.NewController(parameter)
	)

	// init kube client
	_, client, err := kube.BuildClient("")
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

	// automatically release expired clusters
	go func() {
		autofree.AutoReleaseExpiredClusterJob(cancellableCtx, flags.AutoReleaseAccount,
			userCtl, clusterCtl, prCtl, environmentCtl)
	}()

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
