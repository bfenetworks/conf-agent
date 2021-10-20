# 配置说明

- 配置 使用 `toml` 数据格式
- 配置分为3部分：
    - Logger：日志相关，必填，将按照配置初始化文件日志对象
    - Basic：基础配置，为Reloader配置的缺省配置，当Reloader没有配置时，会使用Basic配置作为Reloader配置。建议配置Basic配置，Reloader配置只在需要的时候进行个性化配置
    - Reloaders: reload列表。


## 1 Logger配置
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| LogDir | string | 日志文件目录 | Y | - | |
| LogName | string | 日志文件名 | Y | - | |
| LogLevel | string | 日志等价 | Y | - |  可选： DEBUG TRACE INFO WARNING ERROR CRITICAL|
| RotateWhen | string | 日志文件切割策略 | Y | - | 可选：M：每分钟 H：每小时 D：每天 MIDNIGHT：午夜切割 |
| BackupCount | int | 日志文件保留格式 | Y | - | |
| Format | string | 日志消息格式 | Y | - | |
| StdOut | bool | 日志内容是否控制台输出 | N | - | |


## 2 Basic配置
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| BFECluster              | string | 当前所在的BFE集群名 | Y |  |  |
| BFEConfDir              | string | bfe配置目录位置 | N | /home/work/bfe/conf |  |
| BFEMonitorPort          | int | BFE监控端口号，配置加载时将调用 | N | 8421 |  |
| BFEReloadTimeoutMs      | int | BFE reload 超时设置 | N | 1500 |  |
| ReloadIntervalMs             | int | 拉取时间间隔 | N | 10000 |  |
| ConfServer              | string | APIServer服务器，用来拉取配置 | Y | - |  |
| ConfTaskHeaders        | map\<string\>string  | 配置请求Header, Api Server 当前会对请求鉴权，需要设置 Authorization 头， [通过Dashboard获取Token](https://github.com/bfenetworks/dashboard/blob/develop/docs/zh-cn/user-guide/system-view/user-management.md#token%E7%AE%A1%E7%90%86) | N | - |  |
| ConfTaskTimeoutMs      | int | 配置拉取超时 | Y | 2500 |  |
| ExtraFileServer         | string | 静态文件服务器，用来拉取静态文件 | Y | - |  |
| ExtraFileTaskHeaders   | map\<string\>string  | 静态文件请求Header, Api Server 当前会对请求鉴权，需要设置 Authorization 头， [通过Dashboard获取Token](https://github.com/bfenetworks/dashboard/blob/develop/docs/zh-cn/user-guide/system-view/user-management.md#token%E7%AE%A1%E7%90%86) | N | - |  |
| ExtraFileTaskTimeoutMs | int | 静态文件拉取超时 | Y | 2500 |  |

## 3 Reloaders配置

Reloaders 是个 map\<string\>Reloader 数据类型，key为名字，value为详细配置。

每个Reloader配置为：
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| ConfDir          | string | 模块配置本地目录 | N | 同模块名 | 模块的配置将保留在 {BFEConfDir}/{ConfDir}/下 |
| BFEReloadAPI  | string | bfe reload API | Y | - | 见 [数据面reload](https://www.bfe-networks.net/zh_cn/operation/reload/) |
| BFEReloadTimeoutMs  |  |  | N  |  | 同 Basic.BFEReloadTimeoutMs，若未设置使用 Basic 设置 |
| ReloadIntervalMs  |  |  | N  |  | 同 Basic.ReloadIntervalMs，若未设置使用 Basic 设置 |
| CopyFiles          | []string | 保留的文件列表 | N | - | 有些配置当前不会通过api server 的配置导出的接口更新，但是bfe冷启动时必须读取。对于这些文件，需要从默认文件夹copy到最新的配置文件夹当做初始化配置。 |
| NormalFileTasks  | []NormalFileTask |  | N  |  | 普通配置文件任务列表。详细说明见后续说明 |
| MultiKeyFileTasks  | []MultiKeyFileTask |  | N  |  | 多个Key配置文件任务列表。详细说明见后续说明 |
| ExtraFileTasks  | []ExtraFileTask |  | N  |  | 有扩展文件的配置文件任务列表。详细说明见后续说明 |

文件任务的定义如下：
- NormalFileTask: 一个API对应一个本地配置文件的形式。
- MultiKeyFileTask: 一个API对应多个本地配置文件的形式。其中API返回多个配置文件的内容，每个key都对应一个本地配置文件。
- ExtraFileTask: 和 NormalFileTask类似。但是配置文件内容需要再解析，得到其依赖的扩展文件。扩展文件也需要下载到本地文件。

一个reloader可以定义多种类型的任务，每种类型的任务格式也可以是多个的。
一个reloader必须至少定义一个任务。

### 3.1 Reloader.NormalFileTasks
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| ConfAPI          | string | APIServer 配置导出的 API | Y | - |  |
| ConfFileName    | string | 文件本地保存的文件名 | Y | - | 最终文件名为： {BFEConfDir}/{ConfDir}_{version}/{ConfFileName} |
| ConfServer  |  |  | N  |  | 同 Basic.ConfServer，若未设置使用 Basic 设置 |
| ConfTaskHeaders  |  |  | N  |  | 同 Basic.ConfTaskHeaders，若未设置使用 Basic 设置 |
| ConfTaskTimeoutMs  |  |  | N  |  | 同 Basic.ConfTaskTimeoutMs，若未设置使用 Basic 设置 |

### 3.2 Reloader.MultiKeyFileTasks
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| ConfAPI          | string | APIServer 配置导出的 API | Y | - |  |
| Key2ConfFile | map\<string\>string | 配置对象和文件本地保存的文件名的映射 | Y | - | |
| ConfServer  |  |  | N  |  | 同 Basic.ConfServer，若未设置使用 Basic 设置 |
| ConfTaskHeaders  |  |  | N  |  | 同 Basic.ConfTaskHeaders，若未设置使用 Basic 设置 |
| ConfTaskTimeoutMs  |  |  | N  |  | 同 Basic.ConfTaskTimeoutMs，若未设置使用 Basic 设置 |


### 3.3 Reloader.ExtraFileTasks
| Key | 数据类型 | 含义  | 必填 | 默认值 | 说明 | 
| - | - | - | - | - | - |
| ExtraFileJSONPaths    | []string | 扩展文件名的JsonPath | N | - | [JsonPath语法](https://goessner.net/articles/JsonPath/), 对于有附件的配置，需要配置 |
| ConfAPI          | string | APIServer 配置导出的 API | Y | - |  |
| ConfFileName    | string | 文件本地保存的文件名 | Y | - | 最终文件名为： {BFEConfDir}/{ConfDir}_{version}/{ConfFileName} |
| ConfServer  |  |  | N  |  | 同 Basic.ConfServer，若未设置使用 Basic 设置 |
| ConfTaskHeaders  |  |  | N  |  | 同 Basic.ConfTaskHeaders，若未设置使用 Basic 设置 |
| ConfTaskTimeoutMs  |  |  | N  |  | 同 Basic.ConfTaskTimeoutMs，若未设置使用 Basic 设置 
| ExtraFileServer  |  |  | N  |  | 同 Basic.ExtraFileServer ，若未设置使用 Basic 设置 |
| ExtraFileTaskHeaders  |  |  | N  |  | 同 Basic.ExtraFileTaskHeaders ，若未设置使用 Basic 设置 |
| ExtraFileTaskTimeoutMs  |  |  | N  |  | 同 Basic.ExtraFileTaskTimeoutMs ，若未设置使用 Basic 设置 |