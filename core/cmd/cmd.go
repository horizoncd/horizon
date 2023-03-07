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

	"github.com/horizoncd/horizon/core/config"
	accessctl "github.com/horizoncd/horizon/core/controller/access"
	accesstokenctl "github.com/horizoncd/horizon/core/controller/accesstoken"
	applicationctl "github.com/horizoncd/horizon/core/controller/application"
	applicationregionctl "github.com/horizoncd/horizon/core/controller/applicationregion"
	"github.com/horizoncd/horizon/core/controller/build"
	clusterctl "github.com/horizoncd/horizon/core/controller/cluster"
	codectl "github.com/horizoncd/horizon/core/controller/code"
	environmentctl "github.com/horizoncd/horizon/core/controller/environment"
	environmentregionctl "github.com/horizoncd/horizon/core/controller/environmentregion"
	envtemplatectl "github.com/horizoncd/horizon/core/controller/envtemplate"
	eventctl "github.com/horizoncd/horizon/core/controller/event"
	groupctl "github.com/horizoncd/horizon/core/controller/group"
	idpctl "github.com/horizoncd/horizon/core/controller/idp"
	memberctl "github.com/horizoncd/horizon/core/controller/member"
	oauthservicectl "github.com/horizoncd/horizon/core/controller/oauth"
	oauthappctl "github.com/horizoncd/horizon/core/controller/oauthapp"
	oauthcheckctl "github.com/horizoncd/horizon/core/controller/oauthcheck"
	prctl "github.com/horizoncd/horizon/core/controller/pipelinerun"
	regionctl "github.com/horizoncd/horizon/core/controller/region"
	registryctl "github.com/horizoncd/horizon/core/controller/registry"
	roltctl "github.com/horizoncd/horizon/core/controller/role"
	scopectl "github.com/horizoncd/horizon/core/controller/scope"
	tagctl "github.com/horizoncd/horizon/core/controller/tag"
	templatectl "github.com/horizoncd/horizon/core/controller/template"
	templateschematagctl "github.com/horizoncd/horizon/core/controller/templateschematag"
	terminalctl "github.com/horizoncd/horizon/core/controller/terminal"
	userctl "github.com/horizoncd/horizon/core/controller/user"
	webhookctl "github.com/horizoncd/horizon/core/controller/webhook"
	accessapi "github.com/horizoncd/horizon/core/http/api/v1/access"
	"github.com/horizoncd/horizon/core/http/api/v1/accesstoken"
	"github.com/horizoncd/horizon/core/http/api/v1/application"
	"github.com/horizoncd/horizon/core/http/api/v1/applicationregion"
	"github.com/horizoncd/horizon/core/http/api/v1/cluster"
	codeapi "github.com/horizoncd/horizon/core/http/api/v1/code"
	"github.com/horizoncd/horizon/core/http/api/v1/environment"
	"github.com/horizoncd/horizon/core/http/api/v1/environmentregion"
	"github.com/horizoncd/horizon/core/http/api/v1/envtemplate"
	"github.com/horizoncd/horizon/core/http/api/v1/event"
	"github.com/horizoncd/horizon/core/http/api/v1/group"
	"github.com/horizoncd/horizon/core/http/api/v1/idp"
	"github.com/horizoncd/horizon/core/http/api/v1/member"
	"github.com/horizoncd/horizon/core/http/api/v1/oauthapp"
	"github.com/horizoncd/horizon/core/http/api/v1/oauthserver"
	"github.com/horizoncd/horizon/core/http/api/v1/pipelinerun"
	"github.com/horizoncd/horizon/core/http/api/v1/region"
	"github.com/horizoncd/horizon/core/http/api/v1/registry"
	roleapi "github.com/horizoncd/horizon/core/http/api/v1/role"
	"github.com/horizoncd/horizon/core/http/api/v1/scope"
	"github.com/horizoncd/horizon/core/http/api/v1/tag"
	"github.com/horizoncd/horizon/core/http/api/v1/template"
	accessv2 "github.com/horizoncd/horizon/core/http/api/v2/access"
	accesstokenv2 "github.com/horizoncd/horizon/core/http/api/v2/accesstoken"
	applicationregionv2 "github.com/horizoncd/horizon/core/http/api/v2/applicationregion"
	codev2 "github.com/horizoncd/horizon/core/http/api/v2/code"
	environmentv2 "github.com/horizoncd/horizon/core/http/api/v2/environment"
	environmentregionv2 "github.com/horizoncd/horizon/core/http/api/v2/environmentregion"
	eventv2 "github.com/horizoncd/horizon/core/http/api/v2/event"
	groupv2 "github.com/horizoncd/horizon/core/http/api/v2/group"
	idpv2 "github.com/horizoncd/horizon/core/http/api/v2/idp"
	memberv2 "github.com/horizoncd/horizon/core/http/api/v2/member"
	oauthappv2 "github.com/horizoncd/horizon/core/http/api/v2/oauthapp"
	pipelinerunv2 "github.com/horizoncd/horizon/core/http/api/v2/pipelinerun"
	regionv2 "github.com/horizoncd/horizon/core/http/api/v2/region"
	registryv2 "github.com/horizoncd/horizon/core/http/api/v2/registry"
	rolev2 "github.com/horizoncd/horizon/core/http/api/v2/role"
	scopev2 "github.com/horizoncd/horizon/core/http/api/v2/scope"
	tagv2 "github.com/horizoncd/horizon/core/http/api/v2/tag"
	templateschematagv2 "github.com/horizoncd/horizon/core/http/api/v2/templateschematag"
	terminalv2 "github.com/horizoncd/horizon/core/http/api/v2/terminal"
	userv2 "github.com/horizoncd/horizon/core/http/api/v2/user"
	webhookv2 "github.com/horizoncd/horizon/core/http/api/v2/webhook"
	"github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/core/middleware/auth"
	recovermiddle "github.com/horizoncd/horizon/core/middleware/recover"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	"github.com/horizoncd/horizon/pkg/environment/service"
	"github.com/horizoncd/horizon/pkg/grafana"
	"github.com/horizoncd/horizon/pkg/token/generator"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"
	tokenstorage "github.com/horizoncd/horizon/pkg/token/storage"
	"github.com/horizoncd/horizon/pkg/util/kube"

	templateschematagapi "github.com/horizoncd/horizon/core/http/api/v1/templateschematag"
	terminalapi "github.com/horizoncd/horizon/core/http/api/v1/terminal"
	"github.com/horizoncd/horizon/core/http/api/v1/user"
	"github.com/horizoncd/horizon/core/http/api/v1/webhook"
	appv2 "github.com/horizoncd/horizon/core/http/api/v2/application"
	buildAPI "github.com/horizoncd/horizon/core/http/api/v2/build"
	envtemplatev2 "github.com/horizoncd/horizon/core/http/api/v2/envtemplate"
	templatev2 "github.com/horizoncd/horizon/core/http/api/v2/template"
	"github.com/horizoncd/horizon/core/http/health"
	"github.com/horizoncd/horizon/core/http/metrics"
	ginlogmiddle "github.com/horizoncd/horizon/core/middleware/ginlog"
	logmiddle "github.com/horizoncd/horizon/core/middleware/log"
	metricsmiddle "github.com/horizoncd/horizon/core/middleware/metrics"
	prehandlemiddle "github.com/horizoncd/horizon/core/middleware/prehandle"
	regionmiddle "github.com/horizoncd/horizon/core/middleware/region"
	tagmiddle "github.com/horizoncd/horizon/core/middleware/tag"
	tokenmiddle "github.com/horizoncd/horizon/core/middleware/token"
	usermiddle "github.com/horizoncd/horizon/core/middleware/user"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/application/gitrepo"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	clustergitrepo "github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clusterservice "github.com/horizoncd/horizon/pkg/cluster/service"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	oauthconfig "github.com/horizoncd/horizon/pkg/config/oauth"
	"github.com/horizoncd/horizon/pkg/config/pprof"
	roleconfig "github.com/horizoncd/horizon/pkg/config/role"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	oauthdao "github.com/horizoncd/horizon/pkg/oauth/dao"
	oauthmanager "github.com/horizoncd/horizon/pkg/oauth/manager"
	scopeservice "github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/rbac"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/templaterelease/output"
	templateschemarepo "github.com/horizoncd/horizon/pkg/templaterelease/schema/repo"
	"github.com/horizoncd/horizon/pkg/templaterepo"
	userservice "github.com/horizoncd/horizon/pkg/user/service"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	clusterv2 "github.com/horizoncd/horizon/core/http/api/v2/cluster"
	"github.com/rbcervilla/redisstore/v8"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile              string
	RoleConfigFile          string
	ScopeRoleFile           string
	BuildJSONSchemaFile     string
	BuildUISchemaFile       string
	Dev                     bool
	Environment             string
	LogLevel                string
	GitOpsRepoDefaultBranch string
}

type RegisterI interface {
	RegisterRoutes(engine *gin.Engine)
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

	flag.StringVar(
		&flags.GitOpsRepoDefaultBranch, "gitOpsRepoDefaultBranch", "master",
		"configure gitops git engine default branch")

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

	gitlabGitops, err := gitlablib.New(coreConfig.GitopsRepoConfig.Token, coreConfig.GitopsRepoConfig.URL)
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

	applicationGitRepo, err := gitrepo.NewApplicationGitlabRepo(ctx, rootGroup, gitlabGitops,
		flags.GitOpsRepoDefaultBranch)

	if err != nil {
		panic(err)
	}

	templateRepo, err := templaterepo.NewRepo(coreConfig.TemplateRepo)
	if err != nil {
		panic(err)
	}

	clusterGitRepo, err := clustergitrepo.NewClusterGitlabRepo(ctx, rootGroup, templateRepo, gitlabGitops,
		flags.GitOpsRepoDefaultBranch)
	if err != nil {
		panic(err)
	}

	templateSchemaGetter := templateschemarepo.NewSchemaGetter(ctx, templateRepo, manager)

	outputGetter, err := output.NewOutPutGetter(ctx, templateRepo, manager)
	if err != nil {
		panic(err)
	}

	gitGetter, err := code.NewGitGetter(ctx, coreConfig.CodeGitRepos)
	if err != nil {
		panic(err)
	}
	tektonFty, err := factory.NewFactory(coreConfig.TektonMapper)
	if err != nil {
		panic(err)
	}

	oauthAppDAO := oauthdao.NewDAO(mysqlDB)
	tokenStorage := tokenstorage.NewStorage(mysqlDB)
	oauthManager := oauthmanager.NewManager(oauthAppDAO, tokenStorage, generator.NewAuthorizeGenerator(),
		coreConfig.Oauth.AuthorizeCodeExpireIn, coreConfig.Oauth.AccessTokenExpireIn)

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
	scopeService, err := scopeservice.NewFileScopeService(oauthConfig)
	if err != nil {
		panic(err)
	}

	autoFreeSvc := service.New(coreConfig.AutoFreeConfig.SupportedEnvs)

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
	tokenSvc := tokenservice.NewService(manager, coreConfig.TokenConfig)

	// init kube client
	_, client, err := kube.BuildClient(coreConfig.KubeConfig)
	if err != nil {
		panic(err)
	}

	grafanaService := grafana.NewService(coreConfig.GrafanaConfig, manager, client)

	parameter := &param.Param{
		Manager:              manager,
		OauthManager:         oauthManager,
		AutoFreeSvc:          autoFreeSvc,
		MemberService:        mservice,
		ApplicationSvc:       applicationSvc,
		ClusterSvc:           clusterSvc,
		GroupSvc:             groupSvc,
		UserSvc:              userSvc,
		TokenSvc:             tokenSvc,
		RoleService:          roleService,
		ScopeService:         scopeService,
		ApplicationGitRepo:   applicationGitRepo,
		TemplateSchemaGetter: templateSchemaGetter,
		Cd:                   cd.NewCD(clusterGitRepo, coreConfig.ArgoCDMapper, flags.GitOpsRepoDefaultBranch),
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
					"(^/apis/core/v[12]/roles)|(^/apis/internal/.*)|(^/login/oauth/authorize)|(^/login/oauth/access_token)|"+
					"(^/apis/core/v[12]/templates$)")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v[12]/idps/endpoints")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v[12]/login/callback")),
			middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/apis/core/v[12]/logout")),
			middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/apis/core/v[12]/users/login")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v[12]/users/self")),
		}

		// init controller
		memberCtl            = memberctl.NewController(parameter)
		applicationCtl       = applicationctl.NewController(parameter)
		envTemplateCtl       = envtemplatectl.NewController(parameter)
		clusterCtl           = clusterctl.NewController(coreConfig, parameter)
		prCtl                = prctl.NewController(parameter)
		templateCtl          = templatectl.NewController(parameter, templateRepo)
		roleCtl              = roltctl.NewController(parameter)
		terminalCtl          = terminalctl.NewController(parameter)
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
		accessTokenCtl       = accesstokenctl.NewController(parameter)
		scopeCtl             = scopectl.NewController(parameter)
		webhookCtl           = webhookctl.NewController(parameter)
		eventCtl             = eventctl.NewController(parameter)
	)

	var (
		// init v1 API
		groupAPI             = group.NewAPI(groupCtl)
		userAPI              = user.NewAPI(userCtl, store)
		applicationAPI       = application.NewAPI(applicationCtl)
		envTemplateAPI       = envtemplate.NewAPI(envTemplateCtl)
		memberAPI            = member.NewAPI(memberCtl, roleService)
		clusterAPI           = cluster.NewAPI(clusterCtl)
		prAPI                = pipelinerun.NewAPI(prCtl)
		environmentAPI       = environment.NewAPI(environmentCtl)
		regionAPI            = region.NewAPI(regionCtl, tagCtl)
		environmentRegionAPI = environmentregion.NewAPI(environmentregionCtl)
		registryAPI          = registry.NewAPI(registryCtl)
		roleAPI              = roleapi.NewAPI(roleCtl)
		terminalAPI          = terminalapi.NewAPI(terminalCtl)
		codeGitAPI           = codeapi.NewAPI(codeGitCtl)
		tagAPI               = tag.NewAPI(tagCtl)
		templateSchemaTagAPI = templateschematagapi.NewAPI(templateSchemaTagCtl)
		templateAPI          = template.NewAPI(templateCtl, templateSchemaTagCtl)
		accessAPI            = accessapi.NewAPI(accessCtl)
		applicationRegionAPI = applicationregion.NewAPI(applicationRegionCtl)
		oauthAppAPI          = oauthapp.NewAPI(oauthAppCtl)
		oauthServerAPI       = oauthserver.NewAPI(oauthServerCtl, oauthAppCtl,
			coreConfig.Oauth.OauthHTMLLocation, scopeService)
		idpAPI         = idp.NewAPI(idpCtrl, store)
		accessTokenAPI = accesstoken.NewAPI(accessTokenCtl, roleService, scopeService)
		scopeAPI       = scope.NewAPI(scopeCtl)
		webhookAPI     = webhook.NewAPI(webhookCtl)
		eventAPI       = event.NewAPI(eventCtl)

		// init v2 API
		accessAPIV2            = accessv2.NewAPI(accessCtl)
		accessTokenAPIV2       = accesstokenv2.NewAPI(accessTokenCtl, roleService, scopeService)
		applicationAPIV2       = appv2.NewAPI(applicationCtl)
		applicationRegionAPIV2 = applicationregionv2.NewAPI(applicationRegionCtl)
		buildSchemaAPI         = buildAPI.NewAPI(buildSchemaCtrl)
		clusterAPIV2           = clusterv2.NewAPI(clusterCtl)
		codeGitAPIV2           = codev2.NewAPI(codeGitCtl)
		environmentAPIV2       = environmentv2.NewAPI(environmentCtl)
		environmentRegionAPIV2 = environmentregionv2.NewAPI(environmentregionCtl)
		envtemplateAPIV2       = envtemplatev2.NewAPI(envTemplateCtl)
		eventAPIV2             = eventv2.NewAPI(eventCtl)
		groupAPIV2             = groupv2.NewAPI(groupCtl)
		idpAPIV2               = idpv2.NewAPI(idpCtrl, store)
		memberAPIV2            = memberv2.NewAPI(memberCtl, roleService)
		oauthAppAPIV2          = oauthappv2.NewAPI(oauthAppCtl)
		pipelinerunAPIV2       = pipelinerunv2.NewAPI(prCtl)
		regionAPIV2            = regionv2.NewAPI(regionCtl, tagCtl)
		registryAPIV2          = registryv2.NewAPI(registryCtl)
		roleAPIV2              = rolev2.NewAPI(roleCtl)
		scopeAPIV2             = scopev2.NewAPI(scopeCtl)
		tagAPIV2               = tagv2.NewAPI(tagCtl)
		templateAPIV2          = templatev2.NewAPI(templateCtl, templateSchemaTagCtl)
		templateSchemaTagAPIV2 = templateschematagv2.NewAPI(templateSchemaTagCtl)
		terminalAPIV2          = terminalv2.NewAPI(terminalCtl)
		userAPIV2              = userv2.NewAPI(userCtl, store)
		webhookAPIV2           = webhookv2.NewAPI(webhookCtl)
	)

	// init server
	r := gin.New()
	// use middleware
	middlewares := []gin.HandlerFunc{
		ginlogmiddle.Middleware(gin.DefaultWriter, "/health", "/metrics"),
		recovermiddle.Middleware(),
		requestid.Middleware(), // requestID middleware, attach a requestID to context
		logmiddle.Middleware(), // log middleware, attach a logger to context

		metricsmiddle.Middleware( // metrics middleware
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		regionmiddle.Middleware(parameter, applicationRegionCtl),
		tokenmiddle.MiddleWare(oauthCheckerCtl, rbacSkippers...),
		//  user middleware, check user and attach current user to context.
		usermiddle.Middleware(parameter, store, coreConfig,
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/front/v1/terminal")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/front/v2/buildschema")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/login/oauth/access_token")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/apis/internal/v2/.*")),
			middleware.MethodAndPathSkipper(http.MethodGet, regexp.MustCompile("^/apis/core/v[12]/idps/endpoints")),
			middleware.MethodAndPathSkipper(http.MethodPost, regexp.MustCompile("^/apis/core/v[12]/users/login"))),
		prehandlemiddle.Middleware(r, manager),
		auth.Middleware(rbacAuthorizer, rbacSkippers...),
		tagmiddle.Middleware(), // tag middleware, parse and attach tagSelector to context
	}
	r.Use(middlewares...)

	gin.ForceConsoleColor()

	// register routes
	health.RegisterRoutes(r)
	metrics.RegisterRoutes(r)

	// v1
	registerV1Group := []RegisterI{
		groupAPI,
		templateAPI,
		userAPI,
		applicationAPI,
		envTemplateAPI,
		clusterAPI,
		prAPI,
		environmentAPI,
		regionAPI,
		environmentRegionAPI,
		registryAPI,
		memberAPI,
		roleAPI,
		terminalAPI,
		codeGitAPI,
		tagAPI,
		templateSchemaTagAPI,
		accessAPI,
		applicationRegionAPI,
		oauthAppAPI,
		oauthServerAPI,
		idpAPI,
		accessTokenAPI,
		scopeAPI,
		webhookAPI,
		eventAPI,
	}

	for _, register := range registerV1Group {
		register.RegisterRoutes(r)
	}

	// v2
	registerV2Group := []RegisterI{
		groupAPIV2,
		accessAPIV2,
		accessTokenAPIV2,
		applicationAPIV2,
		applicationRegionAPIV2,
		buildSchemaAPI,
		clusterAPIV2,
		codeGitAPIV2,
		environmentAPIV2,
		environmentRegionAPIV2,
		envtemplateAPIV2,
		eventAPIV2,
		idpAPIV2,
		memberAPIV2,
		oauthAppAPIV2,
		pipelinerunAPIV2,
		regionAPIV2,
		registryAPIV2,
		roleAPIV2,
		scopeAPIV2,
		tagAPIV2,
		templateAPIV2,
		templateSchemaTagAPIV2,
		terminalAPIV2,
		userAPIV2,
		webhookAPIV2,
	}
	for _, register := range registerV2Group {
		register.RegisterRoutes(r)
	}

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
