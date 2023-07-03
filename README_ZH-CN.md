<img src="image/readme/horizon.svg">

[![license](https://img.shields.io/hexpm/l/plug)]() [![version](https://img.shields.io/badge/horizon-v2.0.1-yellow)]() [![Lint & Unit Test](https://github.com/horizoncd/horizon/actions/workflows/unit-test.yml/badge.svg)](https://github.com/horizoncd/horizon/actions/workflows/unit-test.yml) [![Test Coverage](https://api.codeclimate.com/v1/badges/5b9cc6ee71b84628a309/test_coverage)](https://codeclimate.com/github/horizoncd/horizon/test_coverage)

# Horizon
> [English](README.md) | 中文

Horizon是一款云原生应用的持续交付（CD）平台。平台团队可以让开发人员轻松、高效、最佳实践地将其代码和中间件部署到云端和Kubernetes上。Horizon受到了ArgoCD和AWS Proton的启发。

## 为什么选择Horizon

1. **标准化**：Kubernetes灵活而强大，但也很复杂。让所有开发人员全面了解Kubernetes并实践最佳实践很难。因此，Horizon引入了模板系统，使最佳实践得到控制，并提供高效的交付。例如，平台团队可以提供默认的资源规格模板（如0.5核心、512MB的tiny、1核心、1GB的small、2核心、4GB的middle等），而无需为开发人员选择限制或请求资源。
2. **安全性和可靠性**：安全性和可靠性一直是挑战。Horizon可使应用程序的每个更改持久化、可逆和可审计。这是通过我们的GitOps最佳实践实现的。Horizon引入了RBAC和成员系统，让您在细粒度的权限控制上实现最佳实践。
3. **多云**：Horizon提供了一个一致的应用程序平台，用于管理多云、混合云。
4. **基础设施无关**：Horizon不限制工作负载的类型。基本的Kubernetes工作负载和自定义的[CR](https://kubernetes.io/docs/concepts/extend-kubernetes/api-extension/custom-resources/)都得到了支持。
5. **高效性**：平台团队可快速基于Horizon模板完成端到端的最佳实践工作负载交付。
6. **生产级**：在[网易云音乐](https://music.163.com/)内，数千名开发人员每天使用Horizon部署工作负载。

## 特征

### GitOps

在Horizon中，Git是唯一的“真相之源”，Horizon将模板和所有值存储在Git存储库中。包括代码、镜像、环境变量、资源规格等所有更改都是持久、可逆和可审计的。

### Horizon模板

Horizon模板是基于helm和jsonschema的。平台团队可以默认提供基本实践（包括安全性、亲和力、优先级、资源等），并为用户提供定义的简单和统一接口，该接口由jsonschema文件定义。jsonschema用于在基于react-jsonschema-form的Horizon上提供用户友好的HTML表单。这是一种高度可扩展和灵活的方式，可以基于简单的模板系统制定自己的最佳实践。

### RBAC和成员

Horizon提供了一个类似于Gitlab的RBAC和成员系统。您可以轻松定义自己的平台成员和角色（就像Kubernetes的角色和角色绑定一样）。在我们的实践中，我们提供了类似于PE、Owner、Maintainer、Guest的角色。Owner与读取（列出pods、读取所有属性等）/写入（部署、构建部署、重启、发布、删除等）权限绑定在一起，guest只有读取权限。

### 集成便捷

Horizon提供OpenAPI、AccessToken、Oauth2.0、IDP Connector、Webhooks。它使得集成内部系统变得容易。

### 使用便捷

我们还提供了产品特性，如模板管理、Kubernetes管理、监控管理、环境管理。平台团队可以通过Horizon Web UI轻松设置。

### 架构

![https://horizoncd.github.io/assets/images/horizon-opensource-arch-d7a6d3d217198c2e20d377615c7e71db.jpg](https://horizoncd.github.io/assets/images/horizon-opensource-arch-d7a6d3d217198c2e20d377615c7e71db.jpg)

### Horizon-Core

Horizon Core Server是Rest Server，导出OpenAPI，该API由Web UI、CLI和其他系统使用。它还提供了以下功能：

+ Kubernetes和环境管理
+ 模板管理
+ PrivateToken、AccessToken管理
+ 组、应用程序、群集管理
+ CI、CD Pipeline管理
+ WekHook管理
+ 用户和成员管理
+ IDP管理

### Gitlab和ArgoCD

+ Gitlab：Gitlab存储应用程序的所有配置，是应用程序的“唯一真相之源”。
+ ArgoCD：ArgoCD是我们默认的GitOps引擎，用于将应用程序工作负载从git仓库同步到Kubernetes中。

### Tekton和S3

+ Tekton：用于我们的默认CI引擎的云原生管道，以从源代码自动构建镜像。
+ S3：完成的管道存储在S3中，您可以使用任何S3兼容服务，如Mino或Aws S3服务等。

### Grafana

为方便起见，我们将监控功能默认集成到Horizon中。只需配置您的Source Prometheus，Horizon就会自动检索指标以在Horizon-Web上显示指标仪表板。

### MySql和Redis

用于存储和缓存基本元信息，例如成员、用户、令牌、Webhook、IDP等。

## 常见问题

### Horizon与ArgoCD

ArgoCD是一款非常适合Kubernetes平台团队或熟悉Kubernetes的用户的工具，实际上Horizon使用argoCD作为默认的GitOps引擎。但是，我们认为对于应用程序开发人员团队来说，它不太用户友好。我们通过核心功能，如群组、成员和RBAC、模板等，使Horizon更加用户友好。

### Horizon与OpenShift

我们认为Horizon和OpenShift解决了同样的问题。它们都为您提供了在Kubernetes和云上构建、部署和运行应用程序的能力。但是，它们看起来根本不同，这主要是因为OpenShift更多地是Kubernetes的扩展和增强，但现在Horizon旨在成为基于Kubernetes和云的持续交付平台。

## Horizon GitOps

GitOps是应用程序交付的最佳实践，Horizon遵循GitOps最佳实践。我们使用Git使应用程序的每个更改都稳定、可靠、安全、可审计和可逆转。

## Horizon用法

在网易云音乐内，平台团队基于Horizon向用户提供各种服务模板，包括Web服务器、无服务器（Knative应用程序）、中间件等。700多个研发人员每天基于Horizon进行数百个构建和部署。

## 开始使用Horizon

+ 按照[安装指南](https://horizoncd.github.io/docs/tutorials/how-to-install)进行操作。
+ 通过[部署第一个工作负载](https://horizoncd.github.io/docs/tutorials/how-to-deploy-your-first-workload)开始入门。
+ 在[horizoncd.github.io](https://horizoncd.github.io/docs/user-guide/common-user/group)上查看其他文档。

## 贡献

我们欢迎社区的贡献！以下是您可以帮助改进此项目的一些方法：

+ 通过在我们的[问题跟踪器](https://github.com/horizoncd/horizon/issues)中打开问题报告和请求新功能来**报告错误和请求新功能**。
+ 通过打开拉取请求**提交代码贡献**。提交之前，请确保遵循我们的[贡献指南](./CONTRIBUTING.md)和[行为准则](./CODE-OF-CONDUCT.md)。
+ 通过建议更改或提交拉取请求来**改进文档**。

感谢所有的[贡献者](https://github.com/horizoncd/horizon/contributors)为帮助使这个项目成功！

## 联系我们

您可以通过以下方式与我们联系：

+ [Discussions](https://github.com/horizoncd/horizon/discussions)

+ [Slack](https://join.slack.com/t/horizoncd/shared_invite/zt-1sehbmzcx-dgIwaExNR4fZKXppj5kmgQ)

+ 微信群

  添加管理员微信，您将被邀请加入该群。
  ![wechat](image/readme/wechat.jpg)