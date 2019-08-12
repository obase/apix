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
	"strconv"
)

func registerServiceHttp(httpServer *gin.Engine, conf *Config) {
	defer log.Flush()
	httpServer.GET("/health", CheckHttpHealth)

	suffix := "@" + conf.HttpHost + ":" + strconv.Itoa(conf.HttpPort)
	myname := center.HttpName(conf.Name)
	regs := &center.Service{
		Id:   myname + suffix,
		Kind: "http",
		Name: myname,
		Host: conf.HttpHost,
		Port: conf.HttpPort,
	}

	chks := &center.Check{
		Type:     "http",
		Target:   fmt.Sprintf("http://%s:%v/health", conf.HttpHost, conf.HttpPort),
		Timeout:  conf.ConsulCheckTimeoutHttp,
		Interval: conf.ConsulCheckIntervalHttp,
	}

	if err := center.Register(regs, chks); err == nil {
		log.Info(nil, "register service success, %v", *regs)
	} else {
		log.Error(nil, "register service error, %v, %v", *regs, err)
	}

	// 下述完全是兼容旧的服务注册逻辑
	regs.Id = conf.Name + suffix
	regs.Name = conf.Name
	if err := center.Register(regs, chks); err == nil {
		log.Info(nil, "register service success, %v", *regs)
	} else {
		log.Error(nil, "register service error, %v, %v", *regs, err)
	}
}

func registerServiceGrpc(grpcServer *grpc.Server, conf *Config) {

	defer log.Flush()
	service := &HealthService{}
	grpc_health_v1.RegisterHealthServer(grpcServer, service)

	suffix := "@" + conf.HttpHost + ":" + strconv.Itoa(conf.HttpPort)
	myname := center.GrpcName(conf.Name)
	regs := &center.Service{
		Id:   myname + suffix,
		Kind: "grpc",
		Name: myname,
		Host: conf.GrpcHost,
		Port: conf.GrpcPort,
	}
	chks := &center.Check{
		Type:     "grpc",
		Target:   fmt.Sprintf("%s:%v/%v", conf.GrpcHost, conf.GrpcPort, service),
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
	// 统一删除
	suffix := "@" + conf.HttpHost + ":" + strconv.Itoa(conf.HttpPort)
	center.Deregister(conf.Name + suffix)
	center.Deregister(center.HttpName(conf.Name) + suffix)
	center.Deregister(center.GrpcName(conf.Name) + suffix)
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
