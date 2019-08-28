package apix

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
	"github.com/obase/httpx/cache"
	"github.com/obase/httpx/ginx"
	"github.com/obase/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"net/http"
	"os"
)

/*方法处理原型*/
type MethodFunc func(ctx context.Context, rdata []byte) (interface{}, error)

/*封装错误类型*/
func ParsingRequestError(err error, tag string) error {
	return &api.Response{
		Code: api.PARSING_REQUEST_ERROR,
		Msg:  err.Error(),
		Tag:  tag,
	}
}

/*
扩展逻辑服务器:
1. 支持proto静态注册
2. 支持api + conf.yml动态注册
3. 支持httpPlugin, httpCache机制
*/
type XServer struct {
	*ginx.Server                 // 扩展gin.Server
	init         map[string]bool // file初始化标志
	serverOption []grpc.ServerOption
	middleFilter []gin.HandlerFunc
	services     []*Service
	routesFunc   func(server *ginx.Server)
	registFunc   func(server *grpc.Server)
}

/*用于apigen工具的方法*/
func (s *XServer) Init(f func(server *XServer)) {
	k := fmt.Sprintf("%p", f)
	if _, ok := s.init[k]; ok {
		return
	}
	f(s)
	s.init[k] = true
}

func (s *XServer) ServerOption(so grpc.ServerOption) {
	s.serverOption = append(s.serverOption, so)
}

func (s *XServer) MiddleFilter(mf gin.HandlerFunc) {
	s.middleFilter = append(s.middleFilter, mf)
}

func (s *XServer) Service(desc *grpc.ServiceDesc, impl interface{}) *Service {
	gs := &Service{
		serviceDesc: desc,
		serviceImpl: impl,
	}
	s.services = append(s.services, gs)
	return gs
}

type Service struct {
	serviceImpl interface{}
	serviceDesc *grpc.ServiceDesc
	groupPath   string
	groupFilter []gin.HandlerFunc
	methods     []*Method
}

func (gs *Service) GroupPath(gpath string) {
	gs.groupPath = gpath
}

func (gs *Service) GroupFilter(gf gin.HandlerFunc) {
	gs.groupFilter = append(gs.groupFilter, gf)
}

func (gs *Service) Method(tag string, adapt MethodFunc) *Method {
	gm := &Method{
		tag:     tag,
		adapter: adapt,
	}
	gs.methods = append(gs.methods, gm)
	return gm
}

type Method struct {
	tag          string
	adapter      MethodFunc        // 对应方法的AdapterFunc
	handlePath   string            // 对应方法的Handler path
	handleFilter []gin.HandlerFunc // 对应方法的Handler Filter
	socketPath   string            // 对应方法的Socket path
	socketFilter []gin.HandlerFunc // 对应方法的Handler Filter
}

func (gm *Method) HandlePath(path string) {
	gm.handlePath = path
}

func (gm *Method) HandleFilter(hf gin.HandlerFunc) {
	gm.handleFilter = append(gm.handleFilter, hf)
}

func (gm *Method) SocketPath(path string) {
	gm.socketPath = path
}

func (gm *Method) SocketFilter(hf gin.HandlerFunc) {
	gm.socketFilter = append(gm.socketFilter, hf)
}

/* 补充gin的IRouter路由信息*/
func (server *XServer) Routes(rf func(server *ginx.Server)) {
	server.routesFunc = rf
}

func (server *XServer) Regist(rf func(server *grpc.Server)) {
	server.registFunc = rf
}

func (server *XServer) Serve() error {
	return server.ServeWith(LoadConfig())
}

func (server *XServer) ServeWith(config *Config) error {

	config = mergeConfig(config)

	// 没有配置任何启动,直接退出. 注意: 没有默认80之类的设置
	if config.GrpcPort == 0 && config.HttpPort == 0 {
		return nil
	}

	var (
		operations   []func()
		grpcServer   *grpc.Server
		grpcListener net.Listener
		httpServer   *http.Server
		httpListener net.Listener
		httpCache    cache.Cache
		err          error
	)

	defer func() {
		log.Flush()
		// 反注册consul服务,另外还设定了超时反注册,双重保障
		if config.Name != "" {
			deregisterService(config)
		}
		// 退出需要明确关闭
		if grpcListener != nil {
			grpcListener.Close()
		}
		if httpListener != nil {
			httpListener.Close()
		}
		if httpCache != nil {
			httpCache.Close()
		}
	}()

	// 创建grpc服务器
	if config.GrpcPort > 0 {
		// 设置keepalive超时
		if config.GrpcKeepAlive != 0 {
			server.serverOption = append(server.serverOption, grpc.KeepaliveParams(keepalive.ServerParameters{
				Time: config.GrpcKeepAlive,
			}))
		}
		grpcServer = grpc.NewServer(server.serverOption...)
		// 安装grpc相关配置
		for _, smeta := range server.services {
			grpcServer.RegisterService(smeta.serviceDesc, smeta.serviceImpl)
		}
		if server.registFunc != nil {
			server.registFunc(grpcServer) // 附加额外的Grpc设置,预防额外逻辑
		}
		// 注册grpc服务
		if config.Name != "" {
			registerServiceGrpc(grpcServer, config)
		}
		// 创建监听端口
		grpcListener, err = graceListenGrpc(config.GrpcHost, config.GrpcPort)
		if err != nil {
			log.Error(nil, "grpc server listen error: %v", err)
			log.Flush()
			return err
		}
		// 启动grpc服务
		operations = append(operations, func() {
			if err = grpcServer.Serve(grpcListener); err != nil {
				log.Error(nil, "grpc server serve error: %v", err)
				log.Flush()
				os.Exit(1)
			}
		})
	}

	// 创建http服务器
	if config.HttpPort > 0 {
		server.Server.Use(server.middleFilter...)
		// 安装http相关配置
		var upgrader *websocket.Upgrader
		var httpRouter ginx.IRouter = server.Server // 设置为顶层
		for _, smeta := range server.services {
			if smeta.groupPath != "" {
				httpRouter = httpRouter.Group(smeta.groupPath, smeta.groupFilter...)
			}
			for _, mmeta := range smeta.methods {
				// POST handle
				if mmeta.handlePath != "" {
					handlers := append(mmeta.handleFilter, CreateHandleFunc(mmeta.adapter, mmeta.tag))
					httpRouter.POST(mmeta.handlePath, handlers...)
				}
				// GET socket
				if mmeta.socketPath != "" {
					if upgrader == nil {
						upgrader = CreateSocketUpgrader(server.Config)
					}
					handlers := append(mmeta.socketFilter, CreateSocketFunc(upgrader, mmeta.adapter, mmeta.tag))
					httpRouter.GET(mmeta.socketPath, handlers...)
				}
			}
		}
		if server.routeFunc != nil {
			// 附加额外的API设置,预防额外逻辑
			server.routeFunc(httpRouter)
		}
		// 注册http检查
		if config.Name != "" {
			registerServiceHttp(server.Server, config)
		}
		// 创建监听端口
		httpListener, err = graceListenHttp(config.HttpHost, config.HttpPort, config.HttpKeepAlive)
		if err != nil {
			log.Error(context.Background(), "http server listen error: %v", err)
			log.Flush()
			return err
		}
		operations = append(operations, func() {
			if err := httpServer.Serve(httpListener); err != nil {
				log.Error(nil, "http server serve error: %v", err)
				log.Flush()
				os.Exit(1)
			}
		})
	}

	// 后置执行操作
	for _, opt := range operations {
		go opt()
	}

	// 优雅关闭http与grpc服务
	graceShutdownOrRestart(grpcServer, grpcListener, httpServer, httpListener)

	return nil
}
