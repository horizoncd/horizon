package cmd

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	applicationctl "g.hz.netease.com/horizon/core/controller/application"
	"g.hz.netease.com/horizon/core/http/api/v1/application"
	"g.hz.netease.com/horizon/core/http/api/v1/group"
	"g.hz.netease.com/horizon/core/http/api/v1/member"
	"g.hz.netease.com/horizon/core/http/api/v1/template"
	"g.hz.netease.com/horizon/core/http/api/v1/user"
	"g.hz.netease.com/horizon/core/http/health"
	"g.hz.netease.com/horizon/core/http/metrics"
	metricsmiddle "g.hz.netease.com/horizon/core/middleware/metrics"
	usermiddle "g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	roleconfig "g.hz.netease.com/horizon/pkg/config/role"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/middleware/auth"
	logmiddle "g.hz.netease.com/horizon/pkg/server/middleware/log"
	ormMiddle "g.hz.netease.com/horizon/pkg/server/middleware/orm"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// Flags defines agent CLI flags.
type Flags struct {
	ConfigFile     string
	RoleConfigFile string
	Dev            bool
	Environment    string
	LogLevel       string
}

// ParseFlags parses agent CLI flags.
func ParseFlags() *Flags {
	var flags Flags

	flag.StringVar(
		&flags.ConfigFile, "config", "", "configuration file path")

	flag.StringVar(
		&flags.RoleConfigFile, "roles", "", "roles file path")

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
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		// The TextFormatter is default, you don't actually have to do this.
		log.SetFormatter(&log.TextFormatter{})
	}
	log.SetOutput(os.Stdout)
	level, err := log.ParseLevel(flags.LogLevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(level)
}

// Run runs the agent.
func Run(flags *Flags) {
	var config Config

	// init log
	InitLog(flags)

	// load config
	data, err := ioutil.ReadFile(flags.ConfigFile)
	if err != nil {
		panic(err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
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
	if err := yaml.Unmarshal(content, &config); err != nil {
		panic(err)
	}

	roleService, err := role.NewFileRoleFrom2(context.TODO(), roleConfig)
	if err != nil {
		panic(err)
	}
	rbacAuthorizer := rbac.NewAuthorizer(roleService, memberservice.Svc)

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

	var (
		// init service
		applicationGitRepo = gitrepo.NewApplicationGitlabRepo(config.GitlabConfig)
	)

	var (
		// init controller
		applicationCtl = applicationctl.NewController(applicationGitRepo)
	)

	var (
		// init API
		groupAPI       = group.NewAPI()
		templateAPI    = template.NewAPI()
		userAPI        = user.NewAPI()
		applicationAPI = application.NewAPI(applicationCtl)
		memberAPI      = member.NewAPI()
	)

	// init server
	r := gin.New()
	// use middleware
	middlewares := []gin.HandlerFunc{
		gin.Recovery(),
		gin.LoggerWithWriter(gin.DefaultWriter, "/health", "/metrics"),
		metricsmiddle.Middleware( // metrics middleware
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
			middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		requestid.Middleware(),        // requestID middleware, attach a requestID to context
		logmiddle.Middleware(),        // log middleware, attach a logger to context
		ormMiddle.Middleware(mysqlDB), // orm db middleware, attach a db to context

	}
	// enable usermiddle and auth when current env is not dev
	if !flags.Dev {
		middlewares = append(middlewares,
			usermiddle.Middleware(config.OIDCConfig, //  user middleware, check user and attach current user to context.
				middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/health")),
				middleware.MethodAndPathSkipper("*", regexp.MustCompile("^/metrics"))),
		)
		middlewares = append(middlewares,
			auth.Middleware(rbacAuthorizer, middleware.MethodAndPathSkipper("*",
				regexp.MustCompile("^/apis/[^c][^o][^r][^e].*"))))
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
	member.RegisterRoutes(r, memberAPI)
	log.Printf("Server started")
	log.Fatal(r.Run(fmt.Sprintf(":%d", config.ServerConfig.Port)))
}
