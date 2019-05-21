package pbx

import (
	"context"
	. "github.com/hashicorp/consul/api"
	"github.com/obase/conf"
	"github.com/obase/log"
	"github.com/pkg/errors"
)

var client *Client

func init() {
	consulAgent, _ := conf.GetString("service.consulAgent")
	// 默认127.0.0.1:8500, 如果设置为0.0.0.0或-表示不启用consul
	if consulAgent != "0.0.0.0" && consulAgent != "-" {
		config := DefaultConfig()
		if consulAgent != "" {
			config.Address = consulAgent
		}
		var err error
		if client, err = NewClient(config); err != nil { // 兼容旧的逻辑
			log.Errorf(context.Background(), "Create consul client failed. error: %v", err)
		}
	}
}

var ErrNilClient = errors.New("consul client nil")

func RegisterService(service *AgentServiceRegistration) error {
	if client != nil {
		return client.Agent().ServiceRegister(service)
	}
	return ErrNilClient
}
