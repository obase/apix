package apix

import (
	"github.com/obase/conf"
	"github.com/obase/httpx/cache"
	"github.com/obase/httpx/ginx"
	"time"
)

/*服务配置,注意兼容性.Grpc服务添加前缀"grpc."*/
type Config struct {
	Name                string            `json:"name" bson:"name" yaml:"name"`                                              // 注册服务名,如果没有则不注册
	HttpCheckTimeout    string            `json:"httpCheckTimeout" bson:"httpCheckTimeout" yaml:"httpCheckTimeout"`          // 注册服务心跳检测超时
	HttpCheckInterval   string            `json:"httpCheckInterval" bson:"httpCheckInterval" yaml:"httpCheckInterval"`       // 注册服务心跳检测间隔
	HttpHost            string            `json:"httpHost" bson:"httpHost" yaml:"httpHost"`                                  // Http暴露主机,默认首个私有IP
	HttpPort            int               `json:"httpPort" bson:"httpPort" yaml:"httpPort"`                                  // Http暴露端口, 默认80
	HttpKeepAlive       time.Duration     `json:"httpKeepAlive" bson:"httpKeepAlive" yaml:"httpKeepAlive"`                   // Keepalive
	HttpCertFile        string            `json:"httpCertFile" bson:"httpCertFile" yaml:"httpCertFile"`                      // 启用TLS
	HttpKeyFile         string            `json:"httpKeyFile" bson:"httpKeyFile" yaml:"httpKeyFile"`                         // 启用TLS
	HttpCache           *cache.Config     `json:"httpCache" bson:"httpCache" yaml:"httpCache"`                               // 是否启用Redis缓存
	HttpPlugin          map[string]string `json:"httpPlugin" bson:"httpPlugin" yaml:"httpPlugin"`                            // 默认参数
	HttpEntry           []ginx.Entry      `json:"httpEntry" bson:"httpEntry" yaml:"httpEntry"`                               // 代理入口配置
	WbskReadBufferSize  int               `json:"wbskReadBufferSize" bson:"wbskReadBufferSize" yaml:"wbskReadBufferSize"`    // 默认4092
	WbskWriteBufferSize int               `json:"wbskWriteBufferSize" bson:"wbskWriteBufferSize" yaml:"wbskWriteBufferSize"` // 默认4092
	WbskNotCheckOrigin  bool              `json:"wbskNotCheckOrigin" bson:"wbskNotCheckOrigin" yaml:"wbskNotCheckOrigin"`    // 默认false

	GrpcHost          string        `json:"grpcHost" bson:"grpcHost" yaml:"grpcHost"`                // 默认本机扫描到的第一个私用IP
	GrpcPort          int           `json:"grpcPort" bson:"grpcPort" yaml:"grpcPort"`                // 若为空表示不启用grpc server
	GrpcKeepAlive     time.Duration `json:"grpcKeepAlive" bson:"grpcKeepAlive" yaml:"grpcKeepAlive"` // 默认不启用
	GrpcCheckTimeout  string        `json:"grpcCheckTimeout" bson:"grpcCheckTimeout" yaml:"grpcCheckTimeout"`
	GrpcCheckInterval string        `json:"grpcCheckInterval" bson:"grpcCheckInterval" yaml:"grpcCheckInterval"`
}

const CKEY = "service"

func LoadConfig() *Config {
	var config *Config
	if ok := conf.Scan(CKEY, &config); !ok {
		return nil
	}
	return config
}

// 合并默认值
func mergeConfig(conf *Config) *Config {

	if conf == nil {
		conf = &Config{}
	}

	// 补充默认逻辑
	if conf.HttpCheckTimeout == "" {
		conf.HttpCheckTimeout = "5s"
	}
	if conf.HttpCheckInterval == "" {
		conf.HttpCheckInterval = "6s"
	}
	if conf.GrpcCheckTimeout == "" {
		conf.GrpcCheckTimeout = "5s"
	}
	if conf.GrpcCheckInterval == "" {
		conf.GrpcCheckInterval = "6s"
	}
	return conf
}
