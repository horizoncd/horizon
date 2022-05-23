module g.hz.netease.com/horizon

go 1.15

require (
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/alicebob/miniredis/v2 v2.15.1
	github.com/argoproj/argo-cd v1.7.8
	github.com/argoproj/argo-rollouts v0.9.2
	github.com/argoproj/gitops-engine v0.1.3-0.20200925215903-d25b8fd69f0d
	github.com/aws/aws-sdk-go v1.31.12
	github.com/gin-gonic/gin v1.7.7
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/golang/mock v1.6.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/mux v1.7.4
	github.com/hashicorp/go-retryablehttp v0.6.8
	github.com/johannesboyne/gofakes3 v0.0.0-20210819161434-5c8dfcfe5310
	github.com/mozillazg/go-pinyin v0.18.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/santhosh-tekuri/jsonschema/v5 v5.0.0
	github.com/satori/go.uuid v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/tektoncd/cli v0.3.1-0.20201026154019-cb027b2293d7
	github.com/tektoncd/pipeline v0.17.1-0.20201027063619-b7badedd0f65
	github.com/vmihailenco/msgpack/v5 v5.3.4
	github.com/xanzy/go-gitlab v0.50.4
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	gopkg.in/igm/sockjs-go.v3 v3.0.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gopkg.in/yaml.v2 v2.4.0
	gorm.io/driver/mysql v1.1.2
	gorm.io/driver/sqlite v1.1.5
	gorm.io/gorm v1.21.15
	gorm.io/plugin/prometheus v0.0.0-20210820101226-2a49866f83ee
	gorm.io/plugin/soft_delete v1.0.3
	k8s.io/api v0.18.9
	k8s.io/apimachinery v0.19.0
	k8s.io/cli-runtime v0.18.8
	k8s.io/client-go v11.0.1-0.20190816222228-6d55c1b1f1ca+incompatible
	k8s.io/klog v1.0.0 // indirect
	k8s.io/kubectl v0.18.8
	k8s.io/kubernetes v1.18.8
	knative.dev/pkg v0.0.0-20200922164940-4bf40ad82aab
	sigs.k8s.io/yaml v1.3.0
)

replace (
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
	k8s.io/api => k8s.io/api v0.18.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.8
	k8s.io/apiserver => k8s.io/apiserver v0.18.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.8
	k8s.io/client-go => k8s.io/client-go v0.18.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.8
	k8s.io/code-generator => k8s.io/code-generator v0.18.8
	k8s.io/component-base => k8s.io/component-base v0.18.8
	k8s.io/cri-api => k8s.io/cri-api v0.18.8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.8
	k8s.io/kubectl => k8s.io/kubectl v0.18.8
	k8s.io/kubelet => k8s.io/kubelet v0.18.8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.8
	k8s.io/metrics => k8s.io/metrics v0.18.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.8
)
