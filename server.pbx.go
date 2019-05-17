package pbx

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"net"
	"os"
)

/*处理引擎*/
type Server struct {
	*Conf // conf.yml中配置数据
	*Meta // proto中的插件数据
}

func (server *Server) Init(f func(server *Server)) {
	k := fmt.Sprintf("%p", f)
	if _, ok := server.init[k]; ok {
		return
	}
	f(server)
	server.init[k] = true
}

// 安装grpc的配置
func (server *Server) Setup(grpcServer *grpc.Server, httpServer *gin.Engine) {

	// 安装grpc相关配置
	if grpcServer != nil {
		for _, smeta := range server.Meta.services {
			grpcServer.RegisterService(smeta.serviceDesc, smeta.serviceImpl)
		}
	}

	// 安装http相关配置
	if httpServer != nil {
		var upgrader *websocket.Upgrader
		for _, smeta := range server.Meta.services {
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
						upgrader = CreateSocketUpgrader(server.Conf)
					}
					handlers := append(mmeta.socketFilter, CreateSocketFunc(upgrader, mmeta.adapter, mmeta.tag))
					router.GET(mmeta.socketPath, handlers...)
				}
			}
		}
	}

}

func (server *Server) SetupAndServe() {

	var operations []func()
	var grpcServer *grpc.Server
	var httpServer *gin.Engine

	// 创建grpc服务器
	if server.Conf.GrpcPort > 0 {
		grpcServer = grpc.NewServer(server.Meta.serverOption...)
		operations = append(operations, func() {
			// 注册grpc服务
			RegisterGrpcHealthCheck(grpcServer, server.Conf)
			// 启动grpc服务
			lis, err := net.Listen("tcp", server.GrpcAddr())
			if err != nil {
				Errorf(context.Background(), "grpc server listen error: %v", err)
				os.Exit(1)
			}
			if err = grpcServer.Serve(lis); err != nil {
				Errorf(context.Background(), "grpc server serve error: %v", err)
				os.Exit(1)
			}
		})
	}

	// 创建http服务器
	if server.Conf.HttpPort > 0 {
		httpServer = gin.New()
		httpServer.Use(server.Meta.middleFilter...)
		operations = append(operations, func() {
			// 注册http检查
			RegisterHttpHealthCheck(httpServer, server.Conf)
			// 启动http服务
			if err := httpServer.Run(server.HttpAddr()); err != nil {
				Errorf(context.Background(), "http server serve error: %v", err)
				os.Exit(1)
			}
		})
	}

	// 安装protobuf元配置
	server.Setup(grpcServer, httpServer)

	// 后置执行操作
	if n := len(operations); n > 0 {
		for i := n - 1; i > 0; i-- {
			go operations[i]()
		}
		operations[0]()
	}

	// 信号处理(TBD)
}

/*
使用pbx区别业务项目的api库
*/

func NewServerWith(c *Conf) *Server {
	return &Server{
		Conf: MergeDefltConf(c),
		Meta: NewMeta(),
	}
}

func NewServer() *Server {
	return NewServerWith(NewConf())
}
