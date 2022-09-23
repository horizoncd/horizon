package cmd

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/config"
	accessctl "g.hz.netease.com/horizon/core/controller/access"
	applicationctl "g.hz.netease.com/horizon/core/controller/application"
	applicationregionctl "g.hz.netease.com/horizon/core/controller/applicationregion"
	"g.hz.netease.com/horizon/core/controller/build"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	codectl "g.hz.netease.com/horizon/core/controller/code"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	environmentregionctl "g.hz.netease.com/horizon/core/controller/environmentregion"
	envtemplatectl "g.hz.netease.com/horizon/core/controller/envtemplate"
	groupctl "g.hz.netease.com/horizon/core/controller/group"
	idpctl "g.hz.netease.com/horizon/core/controller/idp"
	memberctl "g.hz.netease.com/horizon/core/controller/member"
	oauthservicectl "g.hz.netease.com/horizon/core/controller/oauth"
	oauthappctl "g.hz.netease.com/horizon/core/controller/oauthapp"
	oauthcheckctl "g.hz.netease.com/horizon/core/controller/oauthcheck"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	regionctl "g.hz.netease.com/horizon/core/controller/region"
	registryctl "g.hz.netease.com/horizon/core/controller/registry"
	roltctl "g.hz.netease.com/horizon/core/controller/role"
	sloctl "g.hz.netease.com/horizon/core/controller/slo"
	tagctl "g.hz.netease.com/horizon/core/controller/tag"
	templatectl "g.hz.netease.com/horizon/core/controller/template"
	templateschematagctl "g.hz.netease.com/horizon/core/controller/templateschematag"
	terminalctl "g.hz.netease.com/horizon/core/controller/terminal"
	userctl "g.hz.netease.com/horizon/core/controller/user"
	accessapi "g.hz.netease.com/horizon/core/http/api/v1/access"
	"g.hz.netease.com/horizon/core/http/api/v1/application"
	"g.hz.netease.com/horizon/core/http/api/v1/applicationregion"
	"g.hz.netease.com/horizon/core/http/api/v1/cluster"
	codeapi "g.hz.netease.com/horizon/core/http/api/v1/code"
	"g.hz.netease.com/horizon/core/http/api/v1/environment"
	"g.hz.netease.com/horizon/core/http/api/v1/environmentregion"
	"g.hz.netease.com/horizon/core/http/api/v1/envtemplate"
	"g.hz.netease.com/horizon/core/http/api/v1/group"
	"g.hz.netease.com/horizon/core/http/api/v1/idp"
	"g.hz.netease.com/horizon/core/http/api/v1/member"
	"g.hz.netease.com/horizon/core/http/api/v1/oauthapp"
	"g.hz.netease.com/horizon/core/http/api/v1/oauthserver"
	"g.hz.netease.com/horizon/core/http/api/v1/pipelinerun"
	"g.hz.netease.com/horizon/core/http/api/v1/region"
	"g.hz.netease.com/horizon/core/http/api/v1/registry"
	roleapi "g.hz.netease.com/horizon/core/http/api/v1/role"
	sloapi "g.hz.netease.com/horizon/core/http/api/v1/slo"
	"g.hz.netease.com/horizon/core/http/api/v1/tag"
	"g.hz.netease.com/horizon/core/http/api/v1/template"
	templatev2 "g.hz.netease.com/horizon/core/http/api/v2/template"

	templateschematagapi "g.hz.netease.com/horizon/core/http/api/v1/templateschematag"
	terminalapi "g.hz.netease.com/horizon/core/http/api/v1/terminal"
	"g.hz.netease.com/horizon/core/http/api/v1/user"
	appv2 "g.hz.netease.com/horizon/core/http/api/v2/application"
	buildAPI "g.hz.netease.com/horizon/core/http/api/v2/build"
	envtemplatev2 "g.hz.netease.com/horizon/core/http/api/v2/envtemplate"
	"g.hz.netease.com/horizon/core/http/health"
	"g.hz.netease.com/horizon/core/http/metrics"
	"g.hz.netease.com/horizon/core/middleware/authenticate"
	ginlogmiddle "g.hz.netease.com/horizon/core/middleware/ginlog"
	metricsmiddle "g.hz.netease.com/horizon/core/middleware/metrics"
	oauthmiddle "g.hz.netease.com/horizon/core/middleware/oauth"
	prehandlemiddle "g.hz.netease.com/horizon/core/middleware/prehandle"
	regionmiddle "g.hz.netease.com/horizon/core/middleware/region"
	tagmiddle "g.hz.netease.com/horizon/core/middleware/tag"
	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clusterservice "g.hz.netease.com/horizon/pkg/cluster/service"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cmdb"
	oauthconfig "g.hz.netease.com/horizon/pkg/config/oauth"
	"g.hz.netease.com/horizon/pkg/config/pprof"
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
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
	logmiddle "g.hz.netease.com/horizon/pkg/server/middleware/log"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschemarepo "g.hz.netease.com/horizon/pkg/templaterelease/schema/repo"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	userservice "g.hz.netease.com/horizon/pkg/user/service"
	"g.hz.netease.com/horizon/pkg/util/kube"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"

	clusterv2 "g.hz.netease.com/horizon/core/http/api/v2/cluster"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile          string
	RoleConfigFile      string
	ScopeRoleFile       string
	BuildJSONSchemaFile string
	BuildUISchemaFile   string
	Dev                 bool
	Environment         string
	LogLevel            string
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

	flag.StringVar(&flags.BuildJSONSchemaFile, "buildjsonschema", "",
		"build json schema file path")

	flag.StringVar(&flags.BuildUISchemaFile, "builduischema", "",
		"build ui schema file path")

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

func runPProfServer(config *pprof.Config) {
	if config.Enabled {
		go func() {
			if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil); err != nil {
				log.Printf("[pprof] failed to start, error: %s", err.Error())
			}
		}()
		log.Printf("[pprof] Listening and serving HTTP on :%d", config.Port)
	}
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
	callbacks.RegisterCustomCallbacks(mysqlDB)

	redisClient := redis.NewClient(&redis.Options{
		Network:  coreConfig.RedisConfig.Protocol,
		Addr:     coreConfig.RedisConfig.Address,
		Password: coreConfig.RedisConfig.Password,
		DB:       int(coreConfig.RedisConfig.DB),
	})

	// session store
	store, err := redisstore.NewRedisStore(context.Background(), redisClient)
	if err != nil {
		panic(err)
	}

	store.Options(sessions.Options{
		Path:   "/",
		MaxAge: int(coreConfig.SessionConfig.MaxAge),
	})
	// https://pkg.go.dev/github.com/gorilla/sessions#section-readme
	gob.Register(&userauth.DefaultInfo{})

	// init manager parameter
	manager := managerparam.InitManager(mysqlDB)
	// init service
	ctx := context.Background()
	gitlabFactory := gitlabfty.NewFactory(coreConfig.GitlabMapper)

	gitlabGitops, err := gitlabFactory.GetByName(ctx, common.GitlabGitops)
	if err != nil {
		panic(err)
	}
	// check existence of gitops root group
	rootGroupPath := coreConfig.GitopsRepoConfig.RootGroupPath
	rootGroup, err := gitlabGitops.GetGroup(ctx, rootGroupPath)
	if err != nil {
		log.Printf("failed to get gitops root group, error: %s, start to create it", err.Error())
		rootGroup, err = gitlabGitops.CreateGroup(ctx, rootGroupPath, rootGroupPath, nil)
		if err != nil {
			panic(err)
		}
	}

	applicationGitRepo, err := gitrepo.NewApplicationGitlabRepo(ctx, rootGroup, gitlabGitops)

	if err != nil {
		panic(err)
	}

	templateRepo, err := templaterepo.NewRepo(coreConfig.TemplateRepo)
	if err != nil {
		panic(err)
	}

	clusterGitRepo, err := clustergitrepo.NewClusterGitlabRepo(ctx, rootGroup, templateRepo, gitlabGitops,
		coreConfig.GitopsRepoConfig.URLSchema)
	if err != nil {
		panic(err)
	}

	gitlabCode, err := gitlabFactory.GetByName(ctx, common.GitlabCode)
	if err != nil {
		panic(err)
	}

	gitlabTemplate, err := gitlabFactory.GetByName(ctx, common.GitlabTemplate)
	if err != nil {
		panic(err)
	}

	templateSchemaGetter := templateschemarepo.NewSchemaGetter(ctx, templateRepo, manager)

	outputGetter, err := output.NewOutPutGetter(ctx, templateRepo, manager)
	if err != nil {
		panic(err)
	}

	gitGetter, err := code.NewGitGetter(ctx, gitlabCode)
	if err != nil {
		panic(err)
	}
	tektonFty, err := factory.NewFactory(coreConfig.TektonMapper)
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
	go memHook.Process()
	common.ElegantExit(memHook)

	oauthAppStore := oauthstore.NewOauthAppStore(mysqlDB)
	oauthTokenStore := oauthstore.NewTokenStore(mysqlDB)
	oauthManager := oauthmanager.NewManager(oauthAppStore, oauthTokenStore,
		generate.NewAuthorizeGenerate(), coreConfig.Oauth.AuthorizeCodeExpireIn, coreConfig.Oauth.AccessTokenExpireIn)

	roleService, err := role.NewFileRoleFrom2(context.TODO(), roleConfig)
	if err != nil {
		panic(err)
	}
	mservice := memberservice.NewService(roleService, oauthManager, manager)
	rbacAuthorizer := rbac.NewAuthorizer(roleService, mservice)

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

	// init build schema controller
	readJSONFileFunc := func(filePath string) map[string]interface{} {
		fileFd, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
		if err != nil {
			panic(err)
		}
		fileContent, err := ioutil.ReadAll(fileFd)
		if err != nil {
			panic(err)
		}
		var schemaFile map[string]interface{}
		err = json.Unmarshal(fileContent, &schemaFile)
		if err != nil {
			panic(err)
		}
		return schemaFile
	}

	buildSchema := &build.Schema{
		JSONSchema: readJSONFileFunc(flags.BuildJSONSchemaFile),
		UISchema:   readJSONFileFunc(flags.BuildUISchemaFile),
	}

	groupSvc := groupservice.NewService(manager)
	applicationSvc := applicationservice.NewService(groupSvc, manager)
	clusterSvc := clusterservice.NewService(applicationSvc, manager)
	userSvc := userservice.NewService(manager)

	// init kube client
	_, client, err := kube.BuildClient("")
	if err != nil {
		panic(err)
	}

	grafanaService := grafana.NewService(coreConfig.GrafanaConfig, manager, client)

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
		GrafanaService:       grafanaService,
		BuildSchema:          buildSchema,
	}

	var (
		rbacSkippers = []middleware.Skipper{
			middleware.MethodAndPathSkipper("*",
				regexp.MustCompile("(^/apis/front/.*)|(^/health)|(^/metrics)|(^/apis/login)|"+
					"(^/apis/core/v1/roles)|(^/apis/internal/.*)|(^/login/oauth/authorize)|(^/login/oauth/access_token)|"+
					"(^/apis/core/v1/templates$)")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v1/idps/endpoints")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v1/login/callback")),
			middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/apis/core/v1/logout")),
			middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/apis/core/v1/users/self")),
		}

		// init controller
		memberCtl            = memberctl.NewController(parameter)
		applicationCtl       = applicationctl.NewController(parameter)
		envTemplateCtl       = envtemplatectl.NewController(parameter)
		clusterCtl           = clusterctl.NewController(coreConfig, parameter)
		prCtl                = prctl.NewController(parameter)
		templateCtl          = templatectl.NewController(parameter, gitlabTemplate, templateRepo)
		roleCtl              = roltctl.NewController(parameter)
		terminalCtl          = terminalctl.NewController(parameter)
		sloCtl               = sloctl.NewController(coreConfig.GrafanaSLO)
		codeGitCtl           = codectl.NewController(gitGetter)
		tagCtl               = tagctl.NewController(parameter)
		templateSchemaTagCtl = templateschematagctl.NewController(parameter)
		accessCtl            = accessctl.NewController(rbacAuthorizer, rbacSkippers...)
		applicationRegionCtl = applicationregionctl.NewController(parameter)
		groupCtl             = groupctl.NewController(parameter)
		oauthCheckerCtl      = oauthcheckctl.NewOauthChecker(parameter)
		oauthAppCtl          = oauthappctl.NewController(parameter)
		oauthServerCtl       = oauthservicectl.NewController(parameter)
		regionCtl            = regionctl.NewController(parameter)
		userCtl              = userctl.NewController(parameter)
		environmentCtl       = environmentctl.NewController(parameter)
		environmentregionCtl = environmentregionctl.NewController(parameter)
		registryCtl          = registryctl.NewController(parameter)
		idpCtrl              = idpctl.NewController(parameter)
		buildSchemaCtrl      = build.NewController(buildSchema)
	)

	var (
		// init API
		groupAPI             = group.NewAPI(groupCtl)
		userAPI              = user.NewAPI(userCtl)
		applicationAPI       = application.NewAPI(applicationCtl)
		applicationAPIV2     = appv2.NewAPI(applicationCtl)
		envTemplateAPI       = envtemplate.NewAPI(envTemplateCtl)
		memberAPI            = member.NewAPI(memberCtl, roleService)
		clusterAPI           = cluster.NewAPI(clusterCtl)
		clusterAPIV2         = clusterv2.NewAPI(clusterCtl)
		prAPI                = pipelinerun.NewAPI(prCtl)
		environmentAPI       = environment.NewAPI(environmentCtl)
		regionAPI            = region.NewAPI(regionCtl, tagCtl)
		environmentRegionAPI = environmentregion.NewAPI(environmentregionCtl)
		registryAPI          = registry.NewAPI(registryCtl)
		roleAPI              = roleapi.NewAPI(roleCtl)
		terminalAPI          = terminalapi.NewAPI(terminalCtl)
		sloAPI               = sloapi.NewAPI(sloCtl)
		codeGitAPI           = codeapi.NewAPI(codeGitCtl)
		tagAPI               = tag.NewAPI(tagCtl)
		templateSchemaTagAPI = templateschematagapi.NewAPI(templateSchemaTagCtl)
		templateAPI          = template.NewAPI(templateCtl, templateSchemaTagCtl)
		templateAPIV2        = templatev2.NewAPI(templateCtl)
		accessAPI            = accessapi.NewAPI(accessCtl)
		applicationRegionAPI = applicationregion.NewAPI(applicationRegionCtl)
		oauthAppAPI          = oauthapp.NewAPI(oauthAppCtl)
		oauthServerAPI       = oauthserver.NewAPI(oauthServerCtl, oauthAppCtl,
			coreConfig.Oauth.OauthHTMLLocation, scopeService)
		idpAPI           = idp.NewAPI(idpCtrl, store)
		buildSchemaAPI   = buildAPI.NewAPI(buildSchemaCtrl)
		envtemplatev2API = envtemplatev2.NewAPI(envTemplateCtl)
	)

	// init server
	r := gin.New()
	// use middleware
	middlewares := []gin.HandlerFunc{
		ginlogmiddle.Middleware(gin.DefaultWriter, "/health", "/metrics"),
		gin.Recovery(),
		requestid.Middleware(), // requestID middleware, attach a requestID to context
		logmiddle.Middleware(), // log middleware, attach a logger to context

		metricsmiddle.Middleware( // metrics middleware
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		regionmiddle.Middleware(parameter, applicationRegionCtl),
		// TODO(gjq): remove this authentication, add OIDC provider
		authenticate.Middleware(coreConfig.AccessSecretKeys, // authenticate middleware, check authentication
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		oauthmiddle.MiddleWare(oauthCheckerCtl, rbacSkippers...),
		//  user middleware, check user and attach current user to context.
		usermiddle.Middleware(parameter, store,
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/front/v1/terminal")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/front/v2/buildschema")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/login/oauth/access_token")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v1/idps/endpoints")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v1/login/callback"))),
		prehandlemiddle.Middleware(r, manager),
		auth.Middleware(rbacAuthorizer, rbacSkippers...),
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
	appv2.RegisterRoutes(r, applicationAPIV2)
	envtemplate.RegisterRoutes(r, envTemplateAPI)
	cluster.RegisterRoutes(r, clusterAPI)
	clusterv2.RegisterRoutes(r, clusterAPIV2)
	pipelinerun.RegisterRoutes(r, prAPI)
	environment.RegisterRoutes(r, environmentAPI)
	region.RegisterRoutes(r, regionAPI)
	environmentregion.RegisterRoutes(r, environmentRegionAPI)
	registry.RegisterRoutes(r, registryAPI)
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
	idp.RegisterRoutes(r, idpAPI)
	buildAPI.RegisterRoutes(r, buildSchemaAPI)
	envtemplatev2.RegisterRoutes(r, envtemplatev2API)
	templatev2.RegisterRoutes(r, templateAPIV2)

	// start cloud event server
	go runCloudEventServer(
		tektonFty,
		coreConfig.CloudEventServerConfig,
		parameter,
		ginlogmiddle.Middleware(gin.DefaultWriter, "/health", "/metrics"),
		requestid.Middleware(),
	)

	// enable pprof
	runPProfServer(&coreConfig.PProf)

	// start api server
	log.Printf("Server started")
	log.Print(r.Run(fmt.Sprintf(":%d", coreConfig.ServerConfig.Port)))
}
