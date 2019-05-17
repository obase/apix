package pbx

import (
	"fmt"
)

/*服务配置,注意兼容性.Grpc服务添加前缀"grpc."*/
type Conf struct {
	Name              string `json:"name"`
	HttpHost          string `json:"httpHost"`
	HttpPort          int    `json:"httpPort"`
	HttpCheckTimeout  string `json:"httpCheckTimeout"`
	HttpCheckInterval string `json:"httpCheckInterval"`
	GrpcHost          string `json:"grpcHost"`
	GrpcPort          int    `json:"grpcPort"`
	GrpcCheckTimeout  string `json:"grpcCheckTimeout"`
	GrpcCheckInterval string `json:"grpcCheckInterval"`
	Mode              string `json:"mode"`
	WsReadBufferSize  int    `json:"wsReadBufferSize"`
	WsWriteBufferSize int    `json:"wsWriteBufferSize"`
	WsNotCheckOrigin  bool   `json:"wsNotCheckOrigin"`
	CenterAddress     string `json:"centerAddress"`
}

// 兼容旧的命名逻辑, 不加任何后缀
func (c *Conf) HttpName() string {
	return c.Name + ".http"
}

// grpc后缀表示grpc服务
func (c *Conf) GrpcName() string {
	return c.Name + ".grpc"
}

func (c *Conf) HttpAddr() string {
	return fmt.Sprintf("%v:%v", c.HttpHost, c.HttpPort)
}

func (c *Conf) GrpcAddr() string {
	return fmt.Sprintf("%v:%v", c.GrpcHost, c.GrpcPort)
}

func NewConf() *Conf {
	return &Conf{}
}

// 合并默认值
func MergeDefltConf(conf *Conf) *Conf {
	if conf == nil {
		conf = NewConf()
	}
	// 补充默认逻辑
	if conf.HttpHost == "" {
		conf.HttpHost = PrivateAddress
	}
	if conf.GrpcHost == "" {
		conf.GrpcHost = PrivateAddress
	}
	if conf.HttpCheckTimeout == "" {
		conf.HttpCheckTimeout = "5s" // 兼容旧值
	}
	if conf.HttpCheckInterval == "" {
		conf.HttpCheckInterval = "6s" // 兼容旧值
	}
	if conf.GrpcCheckTimeout == "" {
		conf.GrpcCheckTimeout = "5s" // 兼容旧值
	}
	if conf.GrpcCheckInterval == "" {
		conf.GrpcCheckInterval = "6s" // 兼容旧值
	}
	return conf
}
