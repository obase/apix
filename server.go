package apix

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
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

/*处理引擎*/
type Server struct {
	*Config                      // conf.yml中配置数据
	init         map[string]bool // file初始化标志
	serverOption []grpc.ServerOption
	middleFilter []gin.HandlerFunc
	services     []*Service
	routeFunc    func(router gin.IRouter)
	registFunc   func(server *grpc.Server)
}

/*用于apigen工具的方法*/
func (s *Server) Init(f func(server *Server)) {
	k := fmt.Sprintf("%p", f)
	if _, ok := s.init[k]; ok {
		return
	}
	f(s)
	s.init[k] = true
}

func (s *Server) ServerOption(so grpc.ServerOption) {
	s.serverOption = append(s.serverOption, so)
}

func (s *Server) MiddleFilter(mf gin.HandlerFunc) {
	s.middleFilter = append(s.middleFilter, mf)
}

func (s *Server) Service(desc *grpc.ServiceDesc, impl interface{}) *Service {
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
func (server *Server) Route(rf func(router gin.IRouter)) {
	server.routeFunc = rf
}

func (server *Server) Regist(rf func(server *grpc.Server)) {
	server.registFunc = rf
}

/*安装*/
func (server *Server) Setup(grpcServer *grpc.Server, httpRouter gin.IRouter) {

	// 安装grpc相关配置
	if grpcServer != nil {
		for _, smeta := range server.services {
			grpcServer.RegisterService(smeta.serviceDesc, smeta.serviceImpl)
		}
		if server.registFunc != nil {
			// 附加额外的Grpc设置,预防额外逻辑
			server.registFunc(grpcServer)
		}
	}

	// 安装http相关配置
	if httpRouter != nil {
		var upgrader *websocket.Upgrader
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
	}

}

func (server *Server) Serve() error {

	if server.Config.GrpcPort == 0 && server.Config.HttpPort == 0 {
		return nil
	}

	var (
		operations   []func()
		grpcServer   *grpc.Server
		grpcListener net.Listener
		httpServer   *http.Server
		httpListener net.Listener
		httpRouter   *gin.Engine
		err          error
	)

	defer func() {
		log.Flush()
		// 反注册consul服务,另外还设定了超时反注册,双重保障
		if server.Config.Name != "" {
			deregisterService(server.Config)
		}
		// 退出需要明确关闭
		if grpcListener != nil {
			grpcListener.Close()
		}
		if httpListener != nil {
			httpListener.Close()
		}
	}()

	// 创建grpc服务器
	if server.Config.GrpcPort > 0 {
		// 设置keepalive超时
		if server.Config.GrpcKeepAlive != 0 {
			server.serverOption = append(server.serverOption, grpc.KeepaliveParams(keepalive.ServerParameters{
				Time: server.Config.GrpcKeepAlive,
			}))
		}
		grpcServer = grpc.NewServer(server.serverOption...)
		// 注册grpc服务
		if server.Config.Name != "" {
			registerServiceGrpc(grpcServer, server.Config)
		}
		// 创建监听端口
		grpcListener, err = graceListenGrpc(server.Config.GrpcHost, server.Config.GrpcPort)
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
	if server.Config.HttpPort > 0 {
		gin.SetMode(gin.ReleaseMode)
		httpRouter = gin.New()
		httpRouter.Use(server.middleFilter...)
		// 注册http检查
		if server.Config.Name != "" {
			registerServiceHttp(httpRouter, server.Config)
		}
		httpServer = &http.Server{Handler: httpRouter}
		// 创建监听端口
		httpListener, err = graceListenHttp(server.Config.HttpHost, server.Config.HttpPort, server.Config.HttpKeepAlive)
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

	// 安装protobuf元配置
	server.Setup(grpcServer, httpRouter)

	// 后置执行操作
	for _, opt := range operations {
		go opt()
	}

	// 优雅关闭http与grpc服务
	graceShutdownOrRestart(grpcServer, grpcListener, httpServer, httpListener)

	return nil
}

/*
使用pbx区别业务项目的api库
*/

func NewServer() *Server {
	return NewServerWith(LoadConfig())
}

func NewServerWith(c *Config) *Server {
	return &Server{
		Config: mergeConfig(c),
		init:   make(map[string]bool),
	}
}
