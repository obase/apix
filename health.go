package apix

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/consul/api"
	"github.com/obase/apix/consul"
	"github.com/obase/apix/grpc_health_v1"
	"google.golang.org/grpc"
	"net/http"
)

func registerServiceHttp(httpServer *gin.Engine, conf *Conf) {

	// 先初始化Consul
	consul.Init()

	httpServer.GET("/health", CheckHttpHealth)

	regs := &api.AgentServiceRegistration{
		Kind:    api.ServiceKind("http"),
		ID:      conf.HttpName() + "@" + conf.HttpAddr(),
		Name:    conf.HttpName(),
		Address: conf.HttpHost,
		Port:    conf.HttpPort,
		Tags:    []string{"http", conf.Name, conf.HttpName()},
		Check: &api.AgentServiceCheck{
			HTTP:                           fmt.Sprintf("http://%s/health", conf.HttpAddr()),
			Timeout:                        conf.ConsulCheckTimeoutHttp,
			Interval:                       conf.ConsulCheckIntervalHttp,
			DeregisterCriticalServiceAfter: conf.ConsulDeregisterServiceAfter,
		},
	}
	consul.RegisterService(regs)

	// 下述完全是兼容旧的服务注册逻辑
	regs.ID = conf.Name + "@" + conf.HttpAddr()
	regs.Name = conf.Name
	regs.Tags = []string{"http", conf.Name}
	consul.RegisterService(regs)

}

func registerServiceGrpc(grpcServer *grpc.Server, conf *Conf) {

	// 先初始化Consul
	consul.Init()

	service := &HealthService{}
	grpc_health_v1.RegisterHealthServer(grpcServer, service)
	regs := &api.AgentServiceRegistration{
		Kind:    api.ServiceKind("grpc"),
		ID:      conf.GrpcName() + "@" + conf.GrpcAddr(),
		Name:    conf.GrpcName(),
		Address: conf.GrpcHost,
		Port:    conf.GrpcPort,
		Tags:    []string{"grpc", conf.Name, conf.GrpcName()},
		Check: &api.AgentServiceCheck{
			GRPC:                           fmt.Sprintf("%v/%v", conf.GrpcAddr(), service),
			Timeout:                        conf.ConsulCheckTimeoutGrpc,
			Interval:                       conf.ConsulCheckIntervalGrpc,
			DeregisterCriticalServiceAfter: conf.ConsulDeregisterServiceAfter,
		},
	}
	consul.RegisterService(regs)
}

func deregisterService(conf *Conf) {
	consul.DeregisterService(conf.GrpcName() + "@" + conf.GrpcAddr())
	consul.DeregisterService(conf.HttpName() + "@" + conf.HttpAddr())
	// 兼容旧接口
	consul.DeregisterService(conf.Name + "@" + conf.HttpAddr())
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
