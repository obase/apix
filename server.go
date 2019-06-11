package apix

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
	"github.com/obase/conf"
	"github.com/obase/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

/*方法处理原型*/
type MethodFunc func(ctx context.Context, rdata []byte) (interface{}, error)
type RouteFunc func(engine *gin.Engine)

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
	conf         *Conf           // conf.yml中配置数据
	init         map[string]bool // file初始化标志
	serverOption []grpc.ServerOption
	middleFilter []gin.HandlerFunc
	services     []*Service
	routeFunc    RouteFunc
}

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

// 安装grpc的配置
func (server *Server) Setup(grpcServer *grpc.Server, httpServer *gin.Engine) {

	// 安装grpc相关配置
	if grpcServer != nil {
		for _, smeta := range server.services {
			grpcServer.RegisterService(smeta.serviceDesc, smeta.serviceImpl)
		}
	}

	// 安装http相关配置
	if httpServer != nil {
		var upgrader *websocket.Upgrader
		for _, smeta := range server.services {
			var router gin.IRouter
			if smeta.groupPath != "" {
				router = httpServer.Group(smeta.groupPath, smeta.groupFilter...)
			} else {
				router = httpServer
			}
			for _, mmeta := range smeta.methods {
				// POST handle
				if mmeta.handlePath != "" {
					handlers := append(mmeta.handleFilter, CreateHandleFunc(mmeta.adapter, mmeta.tag))
					router.POST(mmeta.handlePath, handlers...)
				}
				// GET socket
				if mmeta.socketPath != "" {
					if upgrader == nil {
						upgrader = CreateSocketUpgrader(server.conf)
					}
					handlers := append(mmeta.socketFilter, CreateSocketFunc(upgrader, mmeta.adapter, mmeta.tag))
					router.GET(mmeta.socketPath, handlers...)
				}
			}
		}
		if server.routeFunc != nil {
			// 附加额外的API设置,预防额外逻辑
			server.routeFunc(httpServer)
		}
	}

}
func (server *Server) Route(rf RouteFunc) {
	server.routeFunc = rf
}

func (server *Server) Serve() error {

	// 先初始化log
	log.Init()

	defer log.Flushf()

	var operations []func()
	var grpcServer *grpc.Server
	var httpServer *http.Server
	var ginEngine *gin.Engine

	// 创建grpc服务器
	if server.conf.GrpcPort > 0 {
		// 设置keepalive超时
		if server.conf.GrpcKeepAlivePeriod != "" {
			if period, err := time.ParseDuration(server.conf.GrpcKeepAlivePeriod); err != nil {
				return err
			} else {
				server.serverOption = append(server.serverOption, grpc.KeepaliveParams(keepalive.ServerParameters{
					Time: period,
				}))
			}
		}
		grpcServer = grpc.NewServer(server.serverOption...)
		// 注册grpc服务
		if server.conf.Name != "" {
			registerServiceGrpc(grpcServer, server.conf)
		}
		// 创建监听端口
		lis, err := net.Listen("tcp", server.conf.GrpcAddr())
		if err != nil {
			log.Errorf(context.Background(), "grpc server listen error: %v", err)
			log.Flushf()
			return err
		}
		// 启动grpc服务
		operations = append(operations, func() {
			if err = grpcServer.Serve(lis); err != nil {
				log.Errorf(context.Background(), "grpc server serve error: %v", err)
				log.Flushf()
				os.Exit(1)
			}
		})
	}

	// 创建http服务器
	if server.conf.HttpPort > 0 {
		ginEngine = gin.New()
		ginEngine.Use(server.middleFilter...)
		// 注册http检查
		if server.conf.Name != "" {
			registerServiceHttp(ginEngine, server.conf)
		}
		httpServer = &http.Server{Handler: ginEngine}
		// 创建监听端口
		lis, err := net.Listen("tcp", server.conf.HttpAddr())
		if err != nil {
			log.Errorf(context.Background(), "grpc server listen error: %v", err)
			log.Flushf()
			return err
		}
		kalis := &tcpKeepAliveListener{
			TCPListener: lis.(*net.TCPListener),
		}
		if server.conf.HttpKeepAlivePeriod != "" {
			if period, err := time.ParseDuration(server.conf.HttpKeepAlivePeriod); err != nil {
				return err
			} else {
				kalis.keepAlivePeriod = period
			}
		}
		operations = append(operations, func() {
			if err := httpServer.Serve(kalis); err != nil {
				log.Errorf(context.Background(), "http server serve error: %v", err)
				log.Flushf()
				os.Exit(1)
			}
		})
	}

	// 安装protobuf元配置
	server.Setup(grpcServer, ginEngine)

	// 后置执行操作
	for _, opt := range operations {
		go opt()
	}

	// 优雅关闭http与grpc服务
	sch := make(chan os.Signal, 1)
	signal.Notify(sch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
_EXIT:
	for {
		sig := <-sch

		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			ws := new(sync.WaitGroup)
			if httpServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					httpServer.Shutdown(context.Background())
				}(ws)
			}
			if grpcServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					grpcServer.GracefulStop()
				}(ws)
			}
			ws.Wait()
			break _EXIT
		}
	}

	// 反注册consul服务,另外还设定了超时反注册,双重保障
	if server.conf.Name != "" {
		deregisterService(server.conf)
	}

	return nil
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if ln.keepAlivePeriod > 0 {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(ln.keepAlivePeriod)
	} else {
		// 用回http.Server默认设置
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	return tc, nil
}

/*
使用pbx区别业务项目的api库
*/

func NewServerWith(c *Conf) *Server {

	server := new(Server)
	server.conf = MergeDefaultConfig(c)
	server.init = make(map[string]bool)

	return server
}

const CKEY = "service"

func NewServer() *Server {
	var cf *Conf
	if ok := conf.Scan(CKEY, &cf); ok {
		return NewServerWith(cf)
	}
	return NewServerWith(nil)
}
