#!/bin/bash

VERSION=
K8S_VERSION="v1.19.16"

CHINA=false
FULL=false
UPGRADE=false
CLEAN=false
CONTAINER_NAME=""
HELM_SET=""
STORAGE_CLASS=""

HTTP_PORT=80

CLOUD=false
MINIKUBE=false
KIND=false

INTERNAL_GITLAB_ENABLED=false
GITLAB="core.args.gitOpsRepoDefaultBranch=main"
GITLAB="$GITLAB,config.gitopsRepoConfig.rootGroupPath=horizoncd1"
GITLAB="$GITLAB,config.gitopsRepoConfig.url=https://gitlab.com"
GITLAB="$GITLAB,config.gitopsRepoConfig.token=glpat-2n6qmgCah_Yz4kErMC5V"
GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.url=https://gitlab.com"
GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.password=glpat-2n6qmgCah_Yz4kErMC5V"
GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.username=root"
GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.name=gitops-creds"

GITHUB_TOKEN=""

# Install horizon of the script
#
# Pre-requisites:
# - kubectl
# - helm
# - docker
# - kind or minikube

# Check if the binary is installed
# If not, return false, else return true
function checkbinary() {
    if command -v "$1" &> /dev/null
    then
        echo "The binary $1 is installed"
        return 0
    else
        echo "The binary $1 is not installed"
        return 1
    fi
}

function mustinstalled() {
    if ! command -v "$1" &> /dev/null
    then
        echo "The binary $1 is not installed"
        exit
    else
        echo "The binary $1 is installed"
    fi
}

function installhelm() {
    echo "Installing helm"
    curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash
}

function installkubectl() {
    echo "Installing kubectl"
    curl -LO https://storage.googleapis.com/kubernetes-release/release/v1.21.0/bin/linux/amd64/kubectl
    chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin/kubectl
}

function installkind() {
    echo "Installing kind"
    curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.1/kind-linux-amd64
    chmod +x ./kind
    sudo mv ./kind /usr/local/bin/kind
}

function installminikube() {
    echo "Installing minikube"
    curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
    chmod +x minikube
    sudo mv minikube /usr/local/bin/
}

function checkprerequesites() {
    mustinstalled docker

    # If kind is not installed, install kind
    if $KIND && ! checkbinary kind
    then
        installkind
    fi

    # If minikube is not installed, install minikube
    if $MINIKUBE && ! checkbinary minikube
    then
        installminikube
    fi

    # If kubectl is not installed, install kubectl
    if ! checkbinary kubectl
    then
        installkubectl
    fi

    # If helm is not installed, install helm
    if ! checkbinary helm
    then
        installhelm
    fi
}

function cmdhelp() {
    echo "Usage: $0 [options]"
    echo "Options:"
    echo "  -c,  --cloud, install on cloud"
    # install for user from China
    echo "  -cn, --china, install with image mirror in China"
    echo "       --clean"
    echo "  -e,  --gitlab-external <defaultBranch> <rootGroupID> <gitlabURL> <token>, create with external gitlab"
    echo "  -f,  --full, install full horizon"
    echo "       --github-token <token>, specify the github token for source code repo"
    echo "  -g,  --gitlab-internal, create with internal gitlab"
    echo "  -h,  --help"
    echo "       --http-port <HTTP PORT>, specify the http port to use, only take effect when -k/-m is set"
    echo "  -i,  --init, deploy init job into cluster"
    echo "  -k,  --kind, create cluster by kind"
    echo "  -kv, --k8s-version <K8S_VERSION>, specify the version of k8s to install, default is $K8S_VERSION, only take effect when -k/-m is set"
    echo "  -m,  --minikube, create cluster by minikube"
    echo "  -s,  --storage-class <STORAGE_CLASS>, specify the storage class to use, only take effect when -c/--cloud is set"
    echo "       --set <HELM SETS>, equals to helm install/upgrade --set ..."
    echo "  -u,  --upgrade, equals to helm upgrade"
    echo "  -v,  --version <VERSION>, specify the version of horizon to install"
}

function kindcreatecluster() {
    echo "Creating kind cluster"

    CONTAINER_NAME="horizon-control-plane"

echo \
"kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: \"ingress-ready=true\"
    extraPortMappings:
      - containerPort: 80
        hostPort: $HTTP_PORT
        protocol: TCP" > /tmp/kind.yaml


    kind create cluster --image=kindest/node:$K8S_VERSION --name=horizon --config=/tmp/kind.yaml || exit 1

    docker exec $CONTAINER_NAME bash -c \
$'echo \'[plugins."io.containerd.grpc.v1.cri".registry.configs."horizon-registry.horizoncd.svc.cluster.local".tls]
  insecure_skip_verify = true\' >> /etc/containerd/config.toml'

    docker exec $CONTAINER_NAME systemctl restart containerd

    kubectl config use-context kind-horizon || exit 1
}

function minikubecreatecluster() {
    echo "Creating minikube cluster"

    CONTAINER_NAME="minikube"

    MEMORY=8000
    CPUS=4
    if ! $FULL
    then
        MEMORY=4000
        CPUS=2
    fi

    minikube start --container-runtime=docker --driver=docker \
        --kubernetes-version=$K8S_VERSION --cpus=$CPUS --memory=$MEMORY --ports=$HTTP_PORT:80 || exit 1

    kubectl config use-context minikube || exit 1
}

function installingress() {
    # install ingress-nginx by helm
    echo "Installing ingress-nginx"
    helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx 2> /dev/null

    cmd="helm install my-ingress-nginx -n ingress-nginx ingress-nginx/ingress-nginx"
    cmd="$cmd --version 4.1.4"
    cmd="$cmd --set controller.hostNetwork=true"
    cmd="$cmd --set controller.watchIngressWithoutClass=true"
    cmd="$cmd --set controller.livenessProbe.initialDelaySeconds=0"
    cmd="$cmd --set controller.readinessProbe.initialDelaySeconds=0"
    cmd="$cmd --set controller.admissionWebhooks.enabled=false"

    if $CHINA
    then
        cmd="$cmd --set controller.image.registry=registry.cn-hangzhou.aliyuncs.com"
        cmd="$cmd --set controller.image.image=horizoncd/registry.k8s.io.ingress-nginx.controller"
    fi

    cmd="$cmd --create-namespace"

    eval "$cmd" > /dev/null

    # wait for ingress-nginx to be ready
    echo "Waiting for ingress-nginx to be ready"
    while ! kubectl wait --for=condition=ready pod -l app.kubernetes.io/instance=my-ingress-nginx,app.kubernetes.io/name=ingress-nginx --timeout=600s -n ingress-nginx 2> /dev/null; do	
      sleep 1	
    done
}

# print progress bar by count of how many pods was ready
# 'kubectl get pods -n horizoncd'
function progressbar() {
  while true; do
   # get count of pods
    total=$(kubectl get pods -nhorizoncd --field-selector=status.phase!=Failed 2> /dev/null | grep -v NAME -c | awk '{print $1}')

    # get count of pods that are ready
    ready=$(kubectl get pods -nhorizoncd --field-selector=status.phase=Running 2> /dev/null | \
      awk '{print $2}' | grep -v READY | awk -F/ '$1 == $2 {print}' | wc -l)
    succeeded=$(kubectl get pods -nhorizoncd --field-selector=status.phase=Succeeded 2> /dev/null | \
                   grep -v NAME -c)
    ready=$((ready + succeeded))

    # calculate progress bar length
    bar_length=50
    completed=$((ready * bar_length / total))

    # print progress bar
    echo -ne '['
    for ((i=0; i<bar_length; i++)); do
      if ((i < completed)); then
        echo -ne '#'
      else
        echo -ne '-'
      fi
    done
    echo -ne "] (${ready}/${total})\r"

    # sleep 1 second, then clear the line
    sleep 1
    echo -ne "\033[K"

    # if all pods are ready, break the loop
    if ((ready == total)); then
      break
    fi
  done
}

function install() {
    helm repo add horizon https://horizoncd.github.io/helm-charts

    echo "Update helm repo"
    helm repo update horizon

    cmd="helm"

    if $UPGRADE
    then
        cmd="$cmd upgrade"
    else
        cmd="$cmd install"
    fi

    if [ -n "$GITHUB_TOKEN" ]
    then
        cmd="$cmd --set $GITHUB_TOKEN"
    fi

    if $FULL
    then
        cmd="$cmd --set tags.full=true,tags.minimal=false"
    else
        cmd="$cmd --set chartmuseum.env.open.STORAGE=local"
    fi

    if $CLOUD
    then
        cmd="$cmd --set mysql.primary.persistence.size=20Gi,gitlab.persistence.size=20Gi,harbor.persistence.persistentVolumeClaim.databse.size=20Gi,harbor.persistence.persistentVolumeClaim.jobservice.size=20Gi,minio.persistence.size=20Gi,harbor.persistence.persistentVolumeClaim.database.size=20Gi"

        if [ -n "$STORAGE_CLASS" ]
        then
          cmd="$cmd --set gitlab.persistence.storageClass=$STORAGE_CLASS,minio.global.storageClass=$STORAGE_CLASS,harbor.persistence.persistentVolumeClaim.database.storageClass=$STORAGE_CLASS,persistence.persistentVolumeClaim.jobservice.jobLog.storageClass=$STORAGE_CLASS,harbor.persistence.persistentVolumeClaim.jobservice.scanDataExports.storageClass=$STORAGE_CLASS,mysql.global.storageClass=$STORAGE_CLASS"
        fi
    fi

    if [ -n "$HELM_SET" ]
    then
        cmd="$cmd --set $HELM_SET"
    fi

    if $INTERNAL_GITLAB_ENABLED
    then
        cmd="$cmd --set gitlab.enabled=true"
    else
        cmd="$cmd --set $GITLAB,gitlab.enabled=false"
    fi

    cmd="$cmd horizon horizon/horizon -n horizoncd"

    if [ -n "$VERSION" ]
    then
        cmd="$cmd --version $VERSION"
    fi

    if $UPGRADE
    then
        cmd="$cmd --timeout 30s"
    else
        cmd="$cmd --create-namespace --timeout 60s"
    fi

    if $CHINA
    then
        cmd="$cmd -f https://raw.githubusercontent.com/horizoncd/helm-charts/main/horizon-cn-values.yaml"
    fi

    if $UPGRADE
    then
        echo "Upgrading horizon"
    else
        echo "Installing horizon"
    fi

    eval "$cmd" 1> /dev/null

    progressbar
}

function clean() {
#    echo "Cleaning horizon"
#    helm uninstall horizon -n horizoncd
#    kubectl delete ns horizoncd
#
#    echo "Cleaning ingress-nginx"
#    helm uninstall my-ingress-nginx -n ingress-nginx
#    kubectl delete ns ingress-nginx

    echo "Cleaning kind cluster"
    if $MINIKUBE
    then
        minikube delete
    else
        kind delete cluster --name horizon
    fi
}

function applyinitjobtok8s(){
    echo "Applying init job to k8s"

    image="horizoncd/init:v1.0.1"
    if $CHINA
    then
        image="registry.cn-hangzhou.aliyuncs.com/horizoncd/horizoncd.init:v1.0.1"
    fi

    INDENT="    "
    if $KIND 
    then
      kubeconfig=$(docker exec horizon-control-plane cat /etc/kubernetes/admin.conf | sed "2,\$s/^/$INDENT/")
    else
      kubeconfig=$(docker exec minikube cat /etc/kubernetes/admin.conf | sed "2,\$s/^/$INDENT/")
    fi

    cat <<EOF | kubectl apply -nhorizoncd -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: horizon-init
data:
  script: |
    import subprocess
    import requests
    import pymysql
    import os

    host = os.environ.get('MYSQL_HOST', '127.0.0.1')
    port = os.environ.get('MYSQL_PORT', '3306')
    username = os.environ.get('MYSQL_USER', 'root')
    password = os.environ.get('MYSQL_PASSWORD', '123456')
    db = os.environ.get('MYSQL_DATABASE', 'horizon')

    connection = pymysql.connect(host=host, user=username,
                                 password=password, database=db, port=int(port), cursorclass=pymysql.cursors.DictCursor)

    sql_registry = "insert into tb_registry (id, name, server, token, path, insecure_skip_tls_verify, kind) VALUES (1, 'local', 'https://horizon-registry.horizoncd.svc.cluster.local', 'YWRtaW46SGFyYm9yMTIzNDU=', 'library', 1, 'harbor')"
    sql_kubernetes = '''INSERT INTO tb_region (id, name, display_name, server, certificate, ingress_domain, prometheus_url, disabled, registry_id) VALUES (1, 'local', 'local','https://kubernetes.default.svc', '$kubeconfig','', '', 0, 1)'''

    sql_tag = "INSERT INTO tb_tag (id, resource_id, resource_type, tag_key, tag_value) VALUES (1, 1, 'regions', 'cloudnative-kubernetes-groups', 'public')"
    sql_environment = "INSERT INTO tb_environment (id, name, display_name, auto_free) VALUES (1, 'local', 'local', 0)"
    sql_environment_region = "INSERT INTO tb_environment_region (id, environment_name, region_name, is_default, disabled) VALUES (1, 'local', 'local', 0, 0)"
    sql_group = "INSERT INTO tb_group (id, name, path, description, visibility_level, parent_id, traversal_ids, region_selector) VALUES (1,'horizon', 'horizon', '', 'private', 0, 1, '- key: cloudnative-kubernetes-groups\n  values:\n    - public\n  operator: ""')"
    sql_template = "INSERT INTO tb_template (id, name, description, repository, group_id, chart_name, only_admin, only_owner, without_ci) VALUES (1, 'deployment', '', 'https://github.com/horizoncd/deployment.git', 0, 'deployment', 0, 0, 1)"
    sql_template_release = "INSERT INTO tb_template_release (id, template_name, name, description, recommended, template, chart_name, only_admin, chart_version, sync_status, failed_reason, commit_id, last_sync_at, only_owner) VALUES (1, 'deployment', 'v0.0.1', '', 1, 1, 'deployment', 0, 'v0.0.1-5e5193b355961b983cab05a83fa22934001ddf4d', 'status_succeed', '', '5e5193b355961b983cab05a83fa22934001ddf4d', '2023-03-22 17:28:38', 0)"
    sql_application = "INSERT INTO tb_application (id, group_id, name, description, priority, git_url, git_subfolder, git_branch, git_ref, git_ref_type, template, template_release, created_by, updated_by) VALUES (1, 1, 'demo', 'example demo app', 'P0', 'https://github.com/horizoncd/springboot-source-demo.git', '', NULL, 'master', 'branch', 'deployment', 'v0.0.1', 1, 1)"
    sql_member = "INSERT INTO tb_member (id, resource_type, resource_id, role, member_type, membername_id, granted_by, created_by) values (1, 'applications', 1, 'owner', 0, 1, 1, 1)"

    sqls = [sql_registry, sql_kubernetes, sql_tag, sql_environment,
            sql_environment_region, sql_group, sql_template, sql_template_release, sql_application, sql_member]

    with connection:
        with connection.cursor() as cursor:
            for sql in sqls:
                try:
                    cursor.execute(sql)
                except Exception as e:
                    print("Error:", e)
                    print("sql:", sql)
        connection.commit()

    user = "horizoncd"
    repo = "deployment"
    format = "tarball"
    branch = "main"

    url = f"https://github.com/{user}/{repo}/{format}/{branch}"

    response = requests.get(url, stream=True)

    chart_file_path = "/tmp/deployment.tgz"

    with open(chart_file_path, "wb") as f:
        for chunk in response.iter_content(chunk_size=1024):
            if chunk:
                f.write(chunk)


    chartmuseum_url = os.environ.get("CHARTMUSEUM_URL", "http://127.0.0.1:8080")

    version = "v0.0.1-5e5193b355961b983cab05a83fa22934001ddf4d"

    command = "helm-cm-push --version {} {} {}".format(version, chart_file_path, chartmuseum_url)
    result = subprocess.run(command, shell=True,
                            stdout=subprocess.PIPE, stderr=subprocess.PIPE)

    if result.returncode == 0:
        print("Chart upload success!")
    else:
        print(f"Chart upload failed: {result.stderr.decode('utf-8')}")
        exit(1)
---
apiVersion: batch/v1
kind: Job
metadata:
  name: horizon-init
spec:
  template:
    spec:
      containers:
      - name: init
        image: $image
        command: ["python","/init/script.py"]
        env:
          - name: MYSQL_HOST
            value: "horizon-mysql.horizoncd.svc.cluster.local"
          - name: MYSQL_PORT
            value: "3306"
          - name: MYSQL_USER
            value: "root"
          - name: MYSQL_PASSWORD
            value: "horizon"
          - name: MYSQL_DATABASE
            value: "horizon"
          - name: CHARTMUSEUM_URL
            value: "http://horizon-chartmuseum.horizoncd.svc.cluster.local:8080"
        volumeMounts:
          - name: init-script
            mountPath: /init
      restartPolicy: Never

      volumes:
        - name: init-script
          configMap:
            name: horizon-init
            items:
              - key: script
                path: script.py
  backoffLimit: 1
EOF

    kubectl wait --for=condition=complete --timeout=60s -nhorizoncd job/horizon-init
}

function parseinput() {
    if [ $# -eq 0 ]
    then
        cmdhelp
        exit
    fi

    while [ $# -gt 0 ]
    do
        case $1 in
            -h|--help)
                cmdhelp
                exit
                ;;
            -k|--kind)
                KIND=true
                shift
                ;;
            -u|--upgrade)
                UPGRADE=true
                shift
                ;;
            -m|--minikube)
                MINIKUBE=true
                shift
                ;;
            --set)
                HELM_SET=$2
                shift 2
                ;;
            -v|--version)
                VERSION=$2
                shift 2
                ;;
            -kv|--k8s-version)
                K8S_VERSION=$2
                shift 2
                ;;
            -g|--gitlab-internal)
                INTERNAL_GITLAB_ENABLED=true
                shift
                ;;
            --github-token)
                GITHUB_TOKEN="config.gitRepos[0].url=https://github.com"
                GITHUB_TOKEN="$GITHUB_TOKEN,config.gitRepos[0].kind=github"
                GITHUB_TOKEN="$GITHUB_TOKEN,config.gitRepos[0].token=$2"
                shift 2
                ;;
            -s|--storage-class)
                STORAGE_CLASS=$2
                shift 2
                ;;
            -e|--external-gitlab)
                GITLAB="core.args.gitOpsRepoDefaultBranch=$2"
                GITLAB="$GITLAB,config.gitopsRepoConfig.rootGroupPath=$3"
                GITLAB="$GITLAB,config.gitopsRepoConfig.url=$4"
                GITLAB="$GITLAB,config.gitopsRepoConfig.token=$5"
                GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.url=$4"
                GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.password=$5"
                GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.username=root"
                GITLAB="$GITLAB,argo-cd.configs.credentialTemplates.gitops-creds.name=gitops-creds"
                shift 5
                ;;
            -cn|--china)
                CHINA=true
                shift
                ;;
            -f|--full)
                FULL=true
                shift
                ;;
            -i|--init)
                applyinitjobtok8s
                echo "Horizon is initialized successfully!"
                exit
                ;;
            -c|--cloud)
                CLOUD=true
                shift
                ;;
            --clean)
                CLEAN=true
                shift
                ;;
            --http-port)
                HTTP_PORT=$2
                shift 2
                ;;
            *)
                echo "Invalid option $1"
                cmdhelp
                exit 1
                ;;
        esac
    done

    if $CLEAN
    then
      clean
      exit
    fi

    if $UPGRADE
    then
      install
      exit
    fi

    checkprerequesites

    if ! $KIND && ! $MINIKUBE && ! $CLOUD
    then
        echo "Please specify the cluster type. kind or minikube or cloud."
        cmdhelp
        exit 1
    elif $KIND
    then
        kindcreatecluster
    elif $MINIKUBE
    then
        minikubecreatecluster
    fi

    if ! $CLOUD
    then
        installingress
    fi

    install

    if ! $CLOUD
    then
        applyinitjobtok8s
    fi

    nameserver=$(kubectl get service -n kube-system kube-dns -o jsonpath="{.spec.clusterIP}")
    docker exec $CONTAINER_NAME bash -c "echo \"nameserver $nameserver\" > /etc/resolv.conf"

    echo 'Horizon is installed successfully!'
    echo "Please access the Horizon UI at http://horizon.h8r.site:$HTTP_PORT with username: admin@cloudnative.com and password: 123456"
}

parseinput "$@"
