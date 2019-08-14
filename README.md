# package api apix
包含api协定框架的元数据. 用户实现service, 轻松提供http, websocket, grpc等多种访问渠道. 


## 支持优雅关闭/重启:
1. graceful shutdown: windows, linux, darwin
```
kill -HUP/-INT/-TERM <pid>, 或者kill <pid>
```
2. graceful restart: linux, darwing
```
kill -USR2 <pid>
```

## api框架的目录结构:
```
$project
   |__src
   |  |__api:     用户编写proto文件,并使用apigen维护基础代码.
   |  |__service: 用户实现业务接口的逻辑代码.
   |  |__main.go: 用户注册service实例
   |
   |__conf.yml: 项目配置文件
      
最基本的项目结构, 另根据需求添加dao, model等业务package
```

## api框架的辅助工具apigen
apigen用于api框架自动代码生成工具, 能大大减少代码编写量!

- windows(64位)版本: https://obase.github.io/apigen/windows/apigen.exe
- linux(64位)版本:   https://obase.github.io/apigen/linux/apigen
- darwin(64位)版本:  https://obase.github.io/apigen/darwin/apigen

其他版本请从源码编译:
- go get -u github.com/obase/apigen
- go install github.com/obase/apigen

在$GOPATH/bin查找编译后可执行文件apigen

## api框架的使用步骤
1. 创建项目目录结构(见上), 所有接口proto文件必须放在api及其子包里.
2. 打开DOS或Shell, 进入src目录, 即api父目录.
```
cd $project/src
```
3. 执行apigen命令,自动生成proto的代码文件并存于api对应目录内.
```
apigen
```
强烈建议不要手工修改$project/src/api目录里面的*.pb.go文件内容, 应该使用apigen工具维护.


## api框架的局限

基于性能考虑, apix使用标准encoding/json(而非grpc的jsonpb)处理protobuf的json. 经测试不支持protobuf的Any特性!

# Installation
- go get
```
go get -u github.com/gin-gonic/gin
go get -u github.com/golang/protobuf
go get -u github.com/gorilla/websocket
go get -u github.com/obase/api
go get -u github.com/obase/center 
go get -u github.com/obase/conf 
go get -u github.com/obase/log
go get -u google.golang.org/grpc 
```
- go mod
```
go mod edit -require=github.com/obase/apix@latest
```
强烈建议go mod, 自动级联下载所需依赖

# Configuration
```
# 服务元数据
service:
  # 服务名称, 自动注册<name>, <name>.http, <name>.grpc三种服务
  name: "demo"
  # Http请求(post请求及websocket请求)主机, 如果为空, 默认本机首个私有IP
  httpHost: "127.0.0.1"
  # Http请求(post请求及websocket请求)端口, 如果为空, 则不启动Http服务器
  httpPort: 8000
  # consul健康检查超时及间隔. 默认5s与6s
  httpCheckTimeout: "5s"
  httpCheckInterval: "6s"
  # Grpc请求主机, 如果为空, 默认本机首个私有IP
  grpcHost: "127.0.0.1"
  # Grpc请求端口, 如果为空, 则不启动Grpc服务器
  grpcPort: 8100
  # consul健康检查超时及间隔
  grpcCheckTimeout: "5s"
  grpcCheckInterval: "6s"
  # 启动模式: DEBUG, TEST, RELEASE
  mode: "DEBUG"
  # Weboscket读缓存大小
  wsReadBufferSize: 8092
  # Websocket写缓存大小
  wsWriteBufferSize: 8092
  # Websocket不校验origin
  wsNotCheckOrigin: false
  # consult 配置地址, 默认不启动.也支持
  # center:
  #   address: "127.0.0.1:8500"
  #   timeout: "30s"
  center: "127.0.0.1"
```

# Index 
- type RouteFunc
```
type RouteFunc func(engine *gin.Engine)
```
- type Server 
```
/*处理引擎*/
type Server struct {
	conf         *Config         // conf.yml中配置数据
	init         map[string]bool // file初始化标志
	serverOption []grpc.ServerOption
	middleFilter []gin.HandlerFunc
	services     []*Service
	routeFunc    RouteFunc
}
```
- func NewServerWith
```
func NewServerWith(c *Config) *Server
```
由用户指定全部配置

- func NewServer()
```
func NewServer() *Server 
```
由conf.yml读取全瓿配置

# Examples
proto
```
syntax = "proto3";

package api;

import "github.com/obase/api/x.proto";

// grpc的白名单过滤器
option (server_option) = {pack:"github.com/obase/demo/system" func:"AccessGuarderGrpc"};
// http的Logger过滤器
option (middle_filter) = {pack:"github.com/obase/demo/system" func:"AccessLoggerHttp"};

service IPlayer {
    // post分组, 配套还有group_filter
    option (group) = {path:"/player"};

    rpc Add (Player) returns (Player) {
        // post请求, 因为配置了group path, 所以路径为/player/add
        option (handle) = {path:"/add"};
        // websocket请求, 因为配置了group path, 所以路径为/player/add
        option (socket) = {path:"/add"};
    }
    rpc Del (Player) returns (void){
        // post请求, 因为配置了group path, 所以路径为/player/del
        option (handle) = {path:"/del"};
        // websocket请求, 因为配置了group path, 所以路径为/player/del
        option (socket) = {path:"/del"};
    }
    rpc Get (Player) returns (Player){
        // post请求, 因为配置了group path, 所以路径为/player/get
        option (handle) = {path:"/get"};
        // websocket请求, 因为配置了group path, 所以路径为/player/get
        option (socket) = {path:"/get"};
    }
    rpc List (void) returns (PlayerList){
        // post请求, 因为配置了group path, 所以路径为/player/list
        option (handle) = {path:"/list"};
        // websocket请求, 因为配置了group path, 所以路径为/player/list
        option (socket) = {path:"/list"};
    }
}

// 默认字段
message Player {
    string id = 1;
    string name = 2;
    string globalRoleId = 3;
}

message PlayerList {
    repeated Player players = 1;
}

service ICorps {
    // post分组, 配套还有group_filter
    option (group) = {path:"/corps"};

    rpc Add (Corps) returns (void) {
        // post请求, 因为配置了group path, 所以路径为/corps/add
        option (handle) = {path:"/add"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/add
        option (socket) = {path:"/add"};
    }
    rpc Del (Corps) returns (void){
        // post请求, 因为配置了group path, 所以路径为/corps/del
        option (handle) = {path:"/del"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/del
        option (socket) = {path:"/del"};
    }
    rpc Get (Corps) returns (Corps){
        // post请求, 因为配置了group path, 所以路径为/corps/get
        option (handle) = {path:"/get"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/get
        option (socket) = {path:"/get"};
    }
    rpc List (void) returns (CorpsList){
        // post请求, 因为配置了group path, 所以路径为/corps/list
        option (handle) = {path:"/list"};
        // websocket请求, 因为配置了group path, 所以路径为/corps/list
        option (socket) = {path:"/list"};
    }
}

message Corps {
    string id = 1;
    string name = 2;
    string logo = 3;
    fixed32 type = 4; // 2, 3, 5
}

message CorpsList {
    repeated Corps corps = 1;
}

```

codes:
```
func main() {
	server := apix.NewServerWith(&apix.Conf{
		HttpHost:         "127.0.0.1",
		HttpPort:         8000,
		GrpcHost:         "127.0.0.1",
		GrpcPort:         9000,
		WsNotCheckOrigin: true, //不检查websocket的Origin,方便测试
	})
	// 注册服务
	api.RegisterICorpsService(server, &service.ICorpsService{})
	api.RegisterIPlayerService(server, &service.IPlayerService{})

	// 启动服务
	if err := server.Serve(); err != nil {
		panic(err)
	}
}

```