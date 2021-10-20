# conf-agent 

## conf-agent 说明
conf-agent 从 api-server 获取最新的配置并触发bfe热加载。

## 获取方式
获取 `conf-agent` 工具。获取 `conf-agent` 有多种方式：
- 在relase页面下载对应平台的可执行文件
- 通过 `go get` 工具本地编译
- 下载本仓库，执行 `make` （需要go开发环境）

## 配置说明
在api-server有对应导出接口的前提下，conf-agent通过配置能够支持所有的module的配置拉取和热加载。

配置详见[配置详情](./config.md)
- 访问 API Server 需要 在配置文件中配置 Token，获取方式见[通过Dashboard获取Token](https://github.com/bfenetworks/dashboard/blob/develop/docs/zh-cn/user-guide/system-view/user-management.md#token%E7%AE%A1%E7%90%86)
- BFECluster 配置为在控制台配置的有效 BFE Cluster，配置方式见[通过Dashboard配置BFECluster](https://github.com/bfenetworks/dashboard/blob/develop/docs/zh-cn/user-guide/system-view/bfe-cluster-and-pool.md#bfe%E9%9B%86%E7%BE%A4%E7%9A%84%E9%85%8D%E7%BD%AE)

## 部署和启动
和 bfe同机部署

启动命令为:

```
./conf-agent -c ../conf/ -cf conf-agent.toml
```

## 实现原理
详见[实现原理](./implementation.md)


## 关于BFE
- 官网：https://www.bfe-networks.net
- 书籍：[《深入理解BFE》](https://github.com/baidu/bfe-book) ：介绍网络接入的相关技术原理，说明BFE的设计思想，以及如何基于BFE搭建现代化的网络接入平台。现已开放全文阅读。
