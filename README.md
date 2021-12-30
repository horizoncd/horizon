## 配置 git pre-commit hooks
- 安装 pre-commit 包管理工具
### Mac
```shell
brew install pre-commit
```
- Install the git hook scripts
```shell
pre-commit install
```

所有commits只有在通过根目录下 **.pre-commit-config.yaml** 文件定义的hooks后，才可顺利提交
