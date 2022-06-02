package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"time"

	"g.hz.netease.com/horizon/core/common"
	accessctl "g.hz.netease.com/horizon/core/controller/access"
	applicationctl "g.hz.netease.com/horizon/core/controller/application"
	applicationregionctl "g.hz.netease.com/horizon/core/controller/applicationregion"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	codectl "g.hz.netease.com/horizon/core/controller/code"
	envtemplatectl "g.hz.netease.com/horizon/core/controller/envtemplate"
	groupctl "g.hz.netease.com/horizon/core/controller/group"
	memberctl "g.hz.netease.com/horizon/core/controller/member"
	oauthservicectl "g.hz.netease.com/horizon/core/controller/oauth"
	oauthappctl "g.hz.netease.com/horizon/core/controller/oauthapp"
	oauthcheckctl "g.hz.netease.com/horizon/core/controller/oauthcheck"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"

	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	roltctl "g.hz.netease.com/horizon/core/controller/role"
	sloctl "g.hz.netease.com/horizon/core/controller/slo"
	tagctl "g.hz.netease.com/horizon/core/controller/tag"
	templatectl "g.hz.netease.com/horizon/core/controller/template"
	templateschematagctl "g.hz.netease.com/horizon/core/controller/templateschematag"
	terminalctl "g.hz.netease.com/horizon/core/controller/terminal"
	accessapi "g.hz.netease.com/horizon/core/http/api/v1/access"
	"g.hz.netease.com/horizon/core/http/api/v1/application"
	"g.hz.netease.com/horizon/core/http/api/v1/applicationregion"
	"g.hz.netease.com/horizon/core/http/api/v1/cluster"
	codeapi "g.hz.netease.com/horizon/core/http/api/v1/code"
	"g.hz.netease.com/horizon/core/http/api/v1/environment"
	"g.hz.netease.com/horizon/core/http/api/v1/envtemplate"
	"g.hz.netease.com/horizon/core/http/api/v1/group"
	"g.hz.netease.com/horizon/core/http/api/v1/member"
	"g.hz.netease.com/horizon/core/http/api/v1/oauthapp"
	"g.hz.netease.com/horizon/core/http/api/v1/oauthserver"
	"g.hz.netease.com/horizon/core/http/api/v1/pipelinerun"

	roleapi "g.hz.netease.com/horizon/core/http/api/v1/role"
	sloapi "g.hz.netease.com/horizon/core/http/api/v1/slo"
	"g.hz.netease.com/horizon/core/http/api/v1/tag"
	"g.hz.netease.com/horizon/core/http/api/v1/template"
	templateschematagapi "g.hz.netease.com/horizon/core/http/api/v1/templateschematag"
	terminalapi "g.hz.netease.com/horizon/core/http/api/v1/terminal"
	"g.hz.netease.com/horizon/core/http/api/v1/user"
	"g.hz.netease.com/horizon/core/http/health"
	"g.hz.netease.com/horizon/core/http/metrics"
	"g.hz.netease.com/horizon/core/middleware/authenticate"
	ginlogmiddle "g.hz.netease.com/horizon/core/middleware/ginlog"
	metricsmiddle "g.hz.netease.com/horizon/core/middleware/metrics"
	oauthmiddle "g.hz.netease.com/horizon/core/middleware/oauth"
	regionmiddle "g.hz.netease.com/horizon/core/middleware/region"

	tagmiddle "g.hz.netease.com/horizon/core/middleware/tag"
	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cmdb"
	oauthconfig "g.hz.netease.com/horizon/pkg/config/oauth"
	"g.hz.netease.com/horizon/pkg/config/region"
	roleconfig "g.hz.netease.com/horizon/pkg/config/role"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/hook"
	"g.hz.netease.com/horizon/pkg/hook/handler"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/oauth/generate"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/scope"
	oauthstore "g.hz.netease.com/horizon/pkg/oauth/store"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
	logmiddle "g.hz.netease.com/horizon/pkg/server/middleware/log"
	ormmiddle "g.hz.netease.com/horizon/pkg/server/middleware/orm"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	tagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile       string
	RoleConfigFile   string
	RegionConfigFile string
	ScopeRoleFile    string
	Dev              bool
	Environment      string
	LogLevel         string
}

// ParseFlags parses agent CLI flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(
		&flags.ConfigFile, "config", "", "configuration file path")

	flag.StringVar(
		&flags.RoleConfigFile, "roles", "", "roles file path")

	flag.StringVar(
		&flags.RegionConfigFile, "regions", "", "regions file path")

	flag.StringVar(
		&flags.ScopeRoleFile, "scopes", "", "configuration file path")

	flag.BoolVar(
		&flags.Dev, "dev", false, "if true, turn off the usermiddleware to skip login")

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

	// load config
	config, err := loadConfig(flags.ConfigFile)
	if err != nil {
		panic(err)
	}
	body, err := json.MarshalIndent(config, "", " ")
	if err != nil {
		panic(err)
	}
	log.Printf("config = %s\n", string(body))

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

	roleService, err := role.NewFileRoleFrom2(context.TODO(), roleConfig)
	if err != nil {
		panic(err)
	}
	mservice := memberservice.NewService(roleService)
	rbacAuthorizer := rbac.NewAuthorizer(roleService, mservice)

	// load region config
	regionFile, err := os.OpenFile(flags.RegionConfigFile, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}

	regionConfig, err := region.LoadRegionConfig(regionFile)
	if err != nil {
		panic(err)
	}
	regionConfigBytes, _ := json.Marshal(regionConfig)
	log.Printf("regions: %v\n", string(regionConfigBytes))

	// init db
	mysqlDB, err := orm.NewMySQLDB(&orm.MySQL{
		Host:              config.DBConfig.Host,
		Port:              config.DBConfig.Port,
		Username:          config.DBConfig.Username,
		Password:          config.DBConfig.Password,
		Database:          config.DBConfig.Database,
		PrometheusEnabled: config.DBConfig.PrometheusEnabled,
	})
	if err != nil {
		panic(err)
	}
	callbacks.RegisterCustomCallbacks(mysqlDB)

	// init service
	ctx := orm.NewContext(context.Background(), mysqlDB)
	gitlabFactory := gitlabfty.NewFactory(config.GitlabMapper)
	applicationGitRepo, err := gitrepo.NewApplicationGitlabRepo(ctx, config.GitlabRepoConfig, gitlabFactory)
	if err != nil {
		panic(err)
	}
	clusterGitRepo, err := clustergitrepo.NewClusterGitlabRepo(ctx, config.GitlabRepoConfig,
		config.HelmRepoMapper, gitlabFactory)
	if err != nil {
		panic(err)
	}
	templateSchemaGetter, err := templateschema.NewSchemaGetter(ctx, gitlabFactory)
	if err != nil {
		panic(err)
	}

	outputGetter, err := output.NewOutPutGetter(ctx, gitlabFactory)
	if err != nil {
		panic(err)
	}

	gitGetter, err := code.NewGitGetter(ctx, gitlabFactory)
	if err != nil {
		panic(err)
	}
	tektonFty, err := factory.NewFactory(config.TektonMapper)
	if err != nil {
		panic(err)
	}

	handlers := make([]hook.EventHandler, 0)
	if config.CmdbConfig.Enabled {
		cmdbController := cmdb.NewController(config.CmdbConfig)
		cmdbHandler := handler.NewCMDBEventHandler(cmdbController)
		handlers = append(handlers, cmdbHandler)
	}
	memHook := hook.NewInMemHook(2000, handlers...)
	go memHook.Process()
	common.ElegantExit(memHook)

	oauthAppStore := oauthstore.NewOauthAppStore(mysqlDB)
	oauthTokenStore := oauthstore.NewTokenStore(mysqlDB)

	authorizeCodeExpireIn := time.Minute * 30
	accessTokenExpireIn := time.Hour * 24
	oauthManager := oauthmanager.NewManager(oauthAppStore, oauthTokenStore,
		generate.NewAuthorizeGenerate(), authorizeCodeExpireIn, accessTokenExpireIn)

	// init scope service
	scopeFile, err := os.OpenFile(flags.ScopeRoleFile, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	content, err = ioutil.ReadAll(scopeFile)
	if err != nil {
		panic(err)
	}
	var oauthConfig oauthconfig.Config
	if err = yaml.Unmarshal(content, &oauthConfig); err != nil {
		panic(err)
	} else {
		log.Printf("the oauthScopeConfig = %+v\n", roleConfig)
	}
	scopeService, err := scope.NewFileScopeService(oauthConfig)
	if err != nil {
		panic(err)
	}

	var (
		rbacSkippers = middleware.MethodAndPathSkipper("*",
			regexp.MustCompile("(^/apis/front/.*)|(^/health)|(^/metrics)|(^/apis/login)|"+
				"(^/apis/core/v1/roles)|(^/apis/internal/.*)"))

		// init controller
		memberCtl      = memberctl.NewController(mservice)
		applicationCtl = applicationctl.NewController(applicationGitRepo, templateSchemaGetter, memHook)
		envTemplateCtl = envtemplatectl.NewController(applicationGitRepo, templateSchemaGetter)
		clusterCtl     = clusterctl.NewController(clusterGitRepo, applicationGitRepo, gitGetter,
			cd.NewCD(config.ArgoCDMapper), tektonFty, templateSchemaGetter, outputGetter, memHook,
			config.GrafanaMapper, tagmanager.Mgr)
		prCtl = prctl.NewController(tektonFty, gitGetter, clusterGitRepo)

		templateCtl          = templatectl.NewController(templateSchemaGetter)
		roleCtl              = roltctl.NewController(roleService)
		terminalCtl          = terminalctl.NewController(clusterGitRepo)
		sloCtl               = sloctl.NewController(config.GrafanaSLO)
		codeGitCtl           = codectl.NewController(gitGetter)
		tagCtl               = tagctl.NewController(clusterGitRepo)
		templateSchemaTagCtl = templateschematagctl.NewController()
		accessCtl            = accessctl.NewController(rbacAuthorizer, rbacSkippers)
		applicationRegionCtl = applicationregionctl.NewController(regionConfig)
		groupCtl             = groupctl.NewController(mservice)
		oauthCheckerCtl      = oauthcheckctl.NewOauthChecker(oauthManager, usermanager.Mgr, scopeService)
		oauthAppCtl          = oauthappctl.NewController(oauthManager, usermanager.Mgr)
		oauthServerCtl       = oauthservicectl.NewController(oauthManager)
	)

	var (
		// init API
		groupAPI             = group.NewAPI(groupCtl)
		userAPI              = user.NewAPI()
		applicationAPI       = application.NewAPI(applicationCtl)
		envTemplateAPI       = envtemplate.NewAPI(envTemplateCtl)
		memberAPI            = member.NewAPI(memberCtl, roleService)
		clusterAPI           = cluster.NewAPI(clusterCtl)
		prAPI                = pipelinerun.NewAPI(prCtl)
		environmentAPI       = environment.NewAPI()
		roleAPI              = roleapi.NewAPI(roleCtl)
		terminalAPI          = terminalapi.NewAPI(terminalCtl)
		sloAPI               = sloapi.NewAPI(sloCtl)
		codeGitAPI           = codeapi.NewAPI(codeGitCtl)
		tagAPI               = tag.NewAPI(tagCtl)
		templateSchemaTagAPI = templateschematagapi.NewAPI(templateSchemaTagCtl)
		templateAPI          = template.NewAPI(templateCtl, templateSchemaTagCtl)
		accessAPI            = accessapi.NewAPI(accessCtl)
		applicationRegionAPI = applicationregion.NewAPI(applicationRegionCtl)
		oauthAppAPI          = oauthapp.NewAPI(oauthAppCtl)
		oauthServerAPI       = oauthserver.NewAPI(oauthServerCtl, oauthAppCtl, config.OauthHTMLLocation, scopeService)
		_                    = oauthapp.NewAPI(oauthAppCtl)
		_                    = oauthserver.NewAPI(oauthServerCtl, oauthAppCtl, config.OauthHTMLLocation, scopeService)
	)

	// init server
	r := gin.New()
	// use middleware
	ormMiddleware := ormmiddle.Middleware(mysqlDB)
	middlewares := []gin.HandlerFunc{
		ginlogmiddle.Middleware(gin.DefaultWriter, "/health", "/metrics"),
		gin.Recovery(),
		requestid.Middleware(), // requestID middleware, attach a requestID to context
		logmiddle.Middleware(), // log middleware, attach a logger to context
		ormMiddleware,          // orm db middleware, attach a db to context
		metricsmiddle.Middleware( // metrics middleware
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		regionmiddle.Middleware(regionConfig),
		// TODO(gjq): remove this authentication, add OIDC provider
		authenticate.Middleware(config.AccessSecretKeys, // authenticate middleware, check authentication
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		usermiddle.Middleware(config.OIDCConfig, //  user middleware, check user and attach current user to context.
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/front/v1/terminal"))),
		auth.Middleware(rbacAuthorizer, rbacSkippers),
		oauthmiddle.MiddleWare(oauthCheckerCtl, rbacSkippers),
		ormmiddle.MiddlewareSetUserContext(mysqlDB),
		tagmiddle.Middleware(), // tag middleware, parse and attach tagSelector to context
	}
	r.Use(middlewares...)

	gin.ForceConsoleColor()

	// register routes
	health.RegisterRoutes(r)
	metrics.RegisterRoutes(r)
	group.RegisterRoutes(r, groupAPI)
	template.RegisterRoutes(r, templateAPI)
	user.RegisterRoutes(r, userAPI)
	application.RegisterRoutes(r, applicationAPI)
	envtemplate.RegisterRoutes(r, envTemplateAPI)
	cluster.RegisterRoutes(r, clusterAPI)
	pipelinerun.RegisterRoutes(r, prAPI)
	environment.RegisterRoutes(r, environmentAPI)
	member.RegisterRoutes(r, memberAPI)
	roleapi.RegisterRoutes(r, roleAPI)
	terminalapi.RegisterRoutes(r, terminalAPI)
	sloapi.RegisterRoutes(r, sloAPI)
	codeapi.RegisterRoutes(r, codeGitAPI)
	tag.RegisterRoutes(r, tagAPI)
	templateschematagapi.RegisterRoutes(r, templateSchemaTagAPI)
	accessapi.RegisterRoutes(r, accessAPI)
	applicationregion.RegisterRoutes(r, applicationRegionAPI)
	oauthapp.RegisterRoutes(r, oauthAppAPI)
	oauthserver.RegisterRoutes(r, oauthServerAPI)

	// start cloud event server
	go runCloudEventServer(ormMiddleware, tektonFty, config.CloudEventServerConfig)
	// start api server
	log.Printf("Server started")
	log.Print(r.Run(fmt.Sprintf(":%d", config.ServerConfig.Port)))
}
