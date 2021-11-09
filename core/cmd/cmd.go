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

	applicationctl "g.hz.netease.com/horizon/core/controller/application"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	memberctl "g.hz.netease.com/horizon/core/controller/member"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	roltctl "g.hz.netease.com/horizon/core/controller/role"
	templatectl "g.hz.netease.com/horizon/core/controller/template"
	"g.hz.netease.com/horizon/core/http/api/v1/application"
	"g.hz.netease.com/horizon/core/http/api/v1/cluster"
	"g.hz.netease.com/horizon/core/http/api/v1/environment"
	"g.hz.netease.com/horizon/core/http/api/v1/group"
	"g.hz.netease.com/horizon/core/http/api/v1/member"
	"g.hz.netease.com/horizon/core/http/api/v1/pipelinerun"
	roleapi "g.hz.netease.com/horizon/core/http/api/v1/role"
	"g.hz.netease.com/horizon/core/http/api/v1/template"
	"g.hz.netease.com/horizon/core/http/api/v1/user"
	"g.hz.netease.com/horizon/core/http/health"
	"g.hz.netease.com/horizon/core/http/metrics"
	"g.hz.netease.com/horizon/core/middleware/authenticate"
	metricsmiddle "g.hz.netease.com/horizon/core/middleware/metrics"
	regionmiddle "g.hz.netease.com/horizon/core/middleware/region"
	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/config/region"
	roleconfig "g.hz.netease.com/horizon/pkg/config/role"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
	logmiddle "g.hz.netease.com/horizon/pkg/server/middleware/log"
	ormmiddle "g.hz.netease.com/horizon/pkg/server/middleware/orm"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile       string
	RoleConfigFile   string
	RegionConfigFile string
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
		logrus.SetReportCaller(true)
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{})
		logrus.SetReportCaller(true)
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
	commitGetter, err := code.NewCommitGetter(ctx, gitlabFactory)
	if err != nil {
		panic(err)
	}
	tektonFty, err := factory.NewFactory(config.TektonMapper)
	if err != nil {
		panic(err)
	}

	var (
		// init controller
		memberCtl      = memberctl.NewController(mservice)
		applicationCtl = applicationctl.NewController(applicationGitRepo, templateSchemaGetter, nil)
		clusterCtl     = clusterctl.NewController(clusterGitRepo, applicationGitRepo, commitGetter,
			cd.NewCD(config.ArgoCDMapper), tektonFty, templateSchemaGetter, nil)
		prCtl       = prctl.NewController(tektonFty)

		templateCtl = templatectl.NewController(templateSchemaGetter)
		roleCtl     = roltctl.NewController(roleService)
	)

	var (
		// init API
		groupAPI       = group.NewAPI()
		templateAPI    = template.NewAPI(templateCtl)
		userAPI        = user.NewAPI()
		applicationAPI = application.NewAPI(applicationCtl)
		memberAPI      = member.NewAPI(memberCtl, roleService)
		clusterAPI     = cluster.NewAPI(clusterCtl)
		prAPI          = pipelinerun.NewAPI(prCtl)
		environmentAPI = environment.NewAPI()
		roleAPI        = roleapi.NewAPI(roleCtl)
	)

	// init server
	r := gin.New()
	// use middleware
	middlewares := []gin.HandlerFunc{
		gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/metrics"),
		gin.Recovery(),
		requestid.Middleware(),        // requestID middleware, attach a requestID to context
		logmiddle.Middleware(),        // log middleware, attach a logger to context
		ormmiddle.Middleware(mysqlDB), // orm db middleware, attach a db to context
		metricsmiddle.Middleware( // metrics middleware
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		regionmiddle.Middleware(regionConfig),
	}
	// enable usermiddle and auth when current env is not dev
	if !flags.Dev {
		// TODO(gjq): remove this authentication, add OIDC provider
		middlewares = append(middlewares, authenticate.Middleware(
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))))
		middlewares = append(middlewares,
			usermiddle.Middleware(config.OIDCConfig, //  user middleware, check user and attach current user to context.
				middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
				middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		)
		middlewares = append(middlewares,
			auth.Middleware(rbacAuthorizer, middleware.MethodAndPathSkipper("*",
				regexp.MustCompile("(^/apis/[^c][^o][^r][^e].*)|(^/health)|(^/metrics)|(^/apis/login)|(^/apis/core/v1/roles)"))))
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
	cluster.RegisterRoutes(r, clusterAPI)
	pipelinerun.RegisterRoutes(r, prAPI)
	environment.RegisterRoutes(r, environmentAPI)
	member.RegisterRoutes(r, memberAPI)
	roleapi.RegisterRoutes(r, roleAPI)

	log.Printf("Server started")
	log.Fatal(r.Run(fmt.Sprintf(":%d", config.ServerConfig.Port)))
}
