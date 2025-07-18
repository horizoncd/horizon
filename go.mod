module github.com/horizoncd/horizon

go 1.23.0

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/argoproj/argo-cd v1.8.7
	github.com/argoproj/argo-rollouts v1.0.7
	github.com/argoproj/gitops-engine v0.3.3
	github.com/aws/aws-sdk-go v1.38.49
	github.com/coreos/go-oidc/v3 v3.2.0
	github.com/gin-gonic/gin v1.7.7
	github.com/go-redis/redis/v8 v8.3.3
	github.com/golang-jwt/jwt/v4 v4.4.3
	github.com/golang/mock v1.6.0
	github.com/google/go-containerregistry v0.1.3
	github.com/google/go-github/v41 v41.0.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.0
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/johannesboyne/gofakes3 v0.0.0-20210819161434-5c8dfcfe5310
	github.com/mattbaird/jsonpatch v0.0.0-20230413205102-771768614e91
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/rbcervilla/redisstore/v8 v8.1.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20201026154019-cb027b2293d7
	github.com/tektoncd/pipeline v0.17.1-0.20201027063619-b7badedd0f65
	github.com/tektoncd/triggers v0.8.2-0.20201007153255-cb1879311818
	github.com/xanzy/go-gitlab v0.50.4
	golang.org/x/net v0.0.0-20220107192237-5cfca573fb4d
	golang.org/x/oauth2 v0.27.0
	gopkg.in/evanphx/json-patch.v5 v5.9.0
	gopkg.in/igm/sockjs-go.v3 v3.0.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/mysql v1.1.2
	gorm.io/driver/sqlite v1.1.5
	gorm.io/gorm v1.21.15
	gorm.io/plugin/prometheus v0.0.0-20210820101226-2a49866f83ee
	gorm.io/plugin/soft_delete v1.0.3
	helm.sh/helm/v3 v3.1.1
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/cli-runtime v0.23.5
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/helm v2.17.0+incompatible
	k8s.io/kubectl v0.23.5
	k8s.io/kubernetes v1.20.10
	knative.dev/pkg v0.0.0-20201026165741-2f75016c1368
	knative.dev/serving v0.18.3
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go v0.65.0 // indirect
	contrib.go.opencensus.io/exporter/ocagent v0.7.1-0.20200907061046-05415f1de66d // indirect
	contrib.go.opencensus.io/exporter/prometheus v0.2.1-0.20200609204449-6bcf6f8577f0 // indirect
	contrib.go.opencensus.io/exporter/stackdriver v0.13.2 // indirect
	github.com/AlecAivazis/survey/v2 v2.0.4 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Microsoft/go-winio v0.4.15 // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578 // indirect
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20190924025748-f65c72e2690d // indirect
	github.com/antlr/antlr4 v0.0.0-20200503195918-621b933c7a7f // indirect
	github.com/argoproj/pkg v0.9.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bombsimon/logrusr v1.0.0 // indirect
	github.com/census-instrumentation/opencensus-proto v0.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/cloudevents/sdk-go/v2 v2.1.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/fatih/color v1.10.0 // indirect
	github.com/gdamore/encoding v1.0.0 // indirect
	github.com/gdamore/tcell v1.3.0 // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/go-openapi/jsonpointer v0.19.3 // indirect
	github.com/go-openapi/jsonreference v0.19.3 // indirect
	github.com/go-openapi/spec v0.19.6 // indirect
	github.com/go-openapi/swag v0.19.7 // indirect
	github.com/go-playground/locales v0.13.0 // indirect
	github.com/go-playground/universal-translator v0.17.0 // indirect
	github.com/go-playground/validator/v10 v10.4.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/google/cel-go v0.6.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/googleapis/gnostic v0.4.1 // indirect
	github.com/gorilla/securecookie v1.1.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/hako/durafmt v0.0.0-20191009132224-3f39dc1ed9f4 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/huandu/xstrings v1.3.2 // indirect
	github.com/igm/sockjs-go v3.0.2+incompatible // indirect
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.2 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/jonboulle/clockwork v0.1.1-0.20190114141812-62fb9bc030d1 // indirect
	github.com/json-iterator/go v1.1.11 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20201106050909-4977a11b4351 // indirect
	github.com/ktr0731/go-fuzzyfinder v0.2.0 // indirect
	github.com/leodido/go-urn v1.2.0 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lightstep/tracecontext.go v0.0.0-20181129014701-1757c391b1ac // indirect
	github.com/lucasb-eyer/go-colorful v1.0.3 // indirect
	github.com/mailru/easyjson v0.7.1 // indirect
	github.com/mattn/go-colorable v0.1.8 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mattn/go-sqlite3 v2.0.2+incompatible // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/moby/term v0.0.0-20201110203204-bea5bbe245bf // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/prometheus/statsd_exporter v0.15.0 // indirect
	github.com/robfig/cron v1.1.0 // indirect
	github.com/russross/blackfriday v1.5.2 // indirect
	github.com/ryszard/goskiplist v0.0.0-20150312221310-2dfbae5fcf46 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/shabbyrobe/gocovmerge v0.0.0-20180507124511-f6ea450bfb63 // indirect
	github.com/spf13/cobra v1.1.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/ugorji/go/codec v1.1.7 // indirect
	github.com/xanzy/ssh-agent v0.3.0 // indirect
	go.opencensus.io v0.22.4 // indirect
	go.opentelemetry.io/otel v0.13.0 // indirect
	go.uber.org/atomic v1.6.0 // indirect
	go.uber.org/multierr v1.5.0 // indirect
	go.uber.org/zap v1.15.0 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/term v0.0.0-20201126162022-7de9c90e9dd1 // indirect
	golang.org/x/text v0.3.6 // indirect
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e // indirect
	golang.org/x/tools v0.1.1 // indirect
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	google.golang.org/api v0.31.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20201110150050-8816d57aaa9a // indirect
	google.golang.org/grpc v1.33.1 // indirect
	google.golang.org/protobuf v1.26.0-rc.1 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.2 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apiserver v0.20.10 // indirect
	k8s.io/component-base v0.20.10 // indirect
	k8s.io/component-helpers v0.20.10 // indirect
	k8s.io/klog/v2 v2.5.0 // indirect
	k8s.io/kube-aggregator v0.20.4 // indirect
	k8s.io/kube-openapi v0.0.0-20210216185858-15cd8face8d6 // indirect
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920 // indirect
	knative.dev/networking v0.0.0-20200922180040-a71b40c69b15 // indirect
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
)

replace (
	k8s.io/api => k8s.io/api v0.20.10
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.10
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.10
	k8s.io/apiserver => k8s.io/apiserver v0.20.10
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.10
	k8s.io/client-go => k8s.io/client-go v0.20.10
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.10
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.10
	k8s.io/code-generator => k8s.io/code-generator v0.20.10
	k8s.io/component-base => k8s.io/component-base v0.20.10
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.10
	k8s.io/controller-manager => k8s.io/controller-manager v0.20.10
	k8s.io/cri-api => k8s.io/cri-api v0.20.10
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.10
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.10
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.10
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20200805222855-6aeccd4b50c6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.10
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.10
	k8s.io/kubectl => k8s.io/kubectl v0.20.10
	k8s.io/kubelet => k8s.io/kubelet v0.20.10
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.10
	k8s.io/metrics => k8s.io/metrics v0.20.10
	k8s.io/mount-utils => k8s.io/mount-utils v0.20.10
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.20.10
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.10
)
