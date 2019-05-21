package pbx

import (
	"github.com/hashicorp/consul/api"
	"sync"
)

var once sync.Once

var client *api.Client

func reset(consulAgent string) {
	var err error
	config := api.DefaultConfig()
	if consulAgent != "" {
		config.Address = conf.ConsulAgent
	}
	if consulClient, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
		Errorf(context.Background(), "Create consul client failed. error: %v", err)
		panic(err)
	}
}

func Client() *api.Client {

}
