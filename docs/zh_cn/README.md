# conf-agent 

## 1. conf-agent 说明
conf-agent 从 api-server 获取最新的配置并触发bfe热加载。

## 2. 获取方式
获取 `conf-agent` 工具。获取 `conf-agent` 有多种方式：
- 在relase页面下载对应平台的可执行文件
- 通过 `go get` 工具本地编译
- 下载本仓库，执行 `make` （需要go开发环境）

## 3. 配置说明
在api-server有对应导出接口的前提下，conf-agent通过配置能够支持所有的module的配置拉取和热加载。

配置详见[配置详情](./config.md)

## 4. 部署和启动
和 bfe同机部署

启动命令为:

```
./conf-agent -c ../conf/ -cf conf-agent.toml
```

## 5. 实现原理
详见[实现原理](./implementation.md)