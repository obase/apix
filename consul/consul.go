package consul

import (
	"context"
	"errors"
	"github.com/hashicorp/consul/api"
	"github.com/obase/conf"
	"github.com/obase/log"
	"sync"
)

var client *api.Client
var once sync.Once

var ErrInvalidClient = errors.New("consul client invalid")

func Init() {
	once.Do(func() {
		consulAddress, _ := conf.GetString("service.consulAddress")
		// 默认127.0.0.1:8500, 如果设置为0.0.0.0或-表示不启用consul
		config := api.DefaultConfig()
		if consulAddress != "" {
			config.Address = consulAddress
		}
		var err error
		if client, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
			log.Errorf(context.Background(), "Connect consul agent error: %s, %v", consulAddress, err)
			log.Flushf()
		} else {
			if _, err = client.Agent().Services(); err != nil {
				log.Errorf(context.Background(), "Connect consul agent error: %s, %v", consulAddress, err)
				log.Flushf()
			} else {
				log.Inforf(context.Background(), "Connect consul agent success: %s", consulAddress)
				log.Flushf()
			}
		}
	})
}

func RegisterService(service *api.AgentServiceRegistration) (err error) {
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

func DeregisterService(serviceId string) {
	if err := client.Agent().ServiceDeregister(serviceId); err != nil {
		log.Errorf(context.Background(), "Deregister consul service error: %v, %v", serviceId, err)
		log.Flushf()
	} else {
		log.Inforf(context.Background(), "Deregister consul service success: %v", serviceId)
	}
}

func DiscoveryService(lastIndex uint64, service string, tags ...string) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	if client != nil {
		return client.Health().ServiceMultipleTags(service, tags, true, &api.QueryOptions{
			WaitIndex: lastIndex,
		})
	}
	return nil, nil, nil
}
