package consul

import (
	"context"
	"errors"
	. "github.com/hashicorp/consul/api"
	"github.com/obase/conf"
	"github.com/obase/log"
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
			log.Errorf(context.Background(), "Connect consul agent error: %s, %v", consulAgent, err)
			log.Flushf()
		} else {
			if _, err = client.Agent().Services(); err != nil {
				log.Errorf(context.Background(), "Connect consul agent error: %s, %v", consulAgent, err)
				log.Flushf()
			} else {
				log.Inforf(context.Background(), "Connect consul agent success: %s", consulAgent)
			}
		}

	}
}

var ErrInvalidClient = errors.New("consul client invalid")

func RegisterService(service *AgentServiceRegistration) (err error) {
	if client != nil {
		if err = client.Agent().ServiceRegister(service); err != nil {
			log.Errorf(context.Background(), "Register consul service error: %v, %v", service, err)
			log.Flushf()
			return err
		} else {
			log.Inforf(context.Background(), "Register consul service success: %v", service)
		}
	}
	return ErrInvalidClient
}

func DiscoveryService(lastIndex uint64, service string, tags ...string) ([]*ServiceEntry, *QueryMeta, error) {
	if client != nil {
		return client.Health().ServiceMultipleTags(service, tags, true, &QueryOptions{
			WaitIndex: lastIndex,
		})
	}
	return nil, nil, nil
}
