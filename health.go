package apix

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/apix/grpc_health_v1"
	"github.com/obase/center"
	"github.com/obase/log"
	"google.golang.org/grpc"
	"net/http"
)

func registerServiceHttp(httpServer *gin.Engine, conf *Config) {

	httpServer.GET("/health", CheckHttpHealth)
	regs := &center.Service{
		Id:   conf.HttpName() + "@" + conf.HttpAddr(),
		Kind: "http",
		Name: conf.HttpName(),
		Host: conf.HttpHost,
		Port: conf.HttpPort,
	}

	chks := &center.Check{
		Type:     "HTTP",
		Target:   fmt.Sprintf("http://%s/health", conf.HttpAddr()),
		Timeout:  conf.ConsulCheckTimeoutHttp,
		Interval: conf.ConsulCheckIntervalHttp,
	}

	if err := center.Register(regs, chks); err == nil {
		log.Info(nil, "register service success, %v", *regs)
	} else {
		log.Error(nil, "register service error, %v, %v", *regs, err)
	}

	// 下述完全是兼容旧的服务注册逻辑
	regs.Id = conf.Name + "@" + conf.HttpAddr()
	regs.Name = conf.Name
	if err := center.Register(regs, chks); err == nil {
		log.Info(nil, "register service success, %v", *regs)
	} else {
		log.Error(nil, "register service error, %v, %v", *regs, err)
	}
}

func registerServiceGrpc(grpcServer *grpc.Server, conf *Config) {

	service := &HealthService{}
	grpc_health_v1.RegisterHealthServer(grpcServer, service)
	regs := &center.Service{
		Id:   conf.GrpcName() + "@" + conf.GrpcAddr(),
		Kind: "grpc",
		Name: conf.GrpcName(),
		Host: conf.GrpcHost,
		Port: conf.GrpcPort,
	}
	chks := &center.Check{
		Type:     "GRPC",
		Target:   fmt.Sprintf("%v/%v", conf.GrpcAddr(), service),
		Timeout:  conf.ConsulCheckTimeoutHttp,
		Interval: conf.ConsulCheckIntervalHttp,
	}

	if err := center.Register(regs, chks); err == nil {
		log.Info(nil, "register service success, %v", *regs)
	} else {
		log.Error(nil, "register service error, %v, %v", *regs, err)
	}
}

func deregisterService(conf *Config) {
	center.Deregister(conf.GrpcName() + "@" + conf.GrpcAddr())
	center.Deregister(conf.HttpName() + "@" + conf.HttpAddr())
	// 兼容旧接口
	center.Deregister(conf.Name + "@" + conf.HttpAddr())
}

func CheckHttpHealth(ctx *gin.Context) {
	ctx.String(http.StatusOK, "OK")
}

type HealthService struct {
}

func (hs *HealthService) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (rsp *grpc_health_v1.HealthCheckResponse, err error) {
	rsp = &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}
	return
}
