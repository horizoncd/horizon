module github.com/horizoncd/horizon

go 1.15

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
	github.com/google/go-github/v41 v41.0.0
	github.com/google/uuid v1.2.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/sessions v1.2.0
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/igm/sockjs-go v3.0.2+incompatible // indirect
	github.com/johannesboyne/gofakes3 v0.0.0-20210819161434-5c8dfcfe5310
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/quasoft/memstore v0.0.0-20191010062613-2bce066d2b0b
	github.com/rbcervilla/redisstore/v8 v8.1.0
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tektoncd/cli v0.3.1-0.20201026154019-cb027b2293d7
	github.com/tektoncd/pipeline v0.17.1-0.20201027063619-b7badedd0f65
	github.com/tektoncd/triggers v0.8.2-0.20201007153255-cb1879311818
	github.com/xanzy/go-gitlab v0.50.4
	golang.org/x/net v0.0.0-20220107192237-5cfca573fb4d
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
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
