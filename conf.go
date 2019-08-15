package apix

import "time"

/*服务配置,注意兼容性.Grpc服务添加前缀"grpc."*/
type Config struct {
	Name                string        `json:"name" bson:"name" yaml:"name"`                            // 服务名称
	HttpHost            string        `json:"httpHost" bson:"httpHost" yaml:"httpHost"`                // 默认本机扫描到的第一个私用IP
	HttpPort            int           `json:"httpPort" bson:"httpPort" yaml:"httpPort"`                // 若为空表示不启用http server
	HttpKeepAlive       time.Duration `json:"httpKeepAlive" bson:"httpKeepAlive" yaml:"httpKeepAlive"` // 默认不启用
	HttpCheckTimeout    string        `json:"httpCheckTimeout" bson:"httpCheckTimeout" yaml:"httpCheckTimeout"`
	HttpCheckInterval   string        `json:"httpCheckInterval" bson:"httpCheckInterval" yaml:"httpCheckInterval"`
	WbskReadBufferSize  int           `json:"wbskReadBufferSize" bson:"wbskReadBufferSize" yaml:"wbskReadBufferSize"`    // 默认4092
	WbskWriteBufferSize int           `json:"wbskWriteBufferSize" bson:"wbskWriteBufferSize" yaml:"wbskWriteBufferSize"` // 默认4092
	WbskNotCheckOrigin  bool          `json:"wbskNotCheckOrigin" bson:"wbskNotCheckOrigin" yaml:"wbskNotCheckOrigin"`    // 默认false
	GrpcHost            string        `json:"grpcHost" bson:"grpcHost" yaml:"grpcHost"`                                  // 默认本机扫描到的第一个私用IP
	GrpcPort            int           `json:"grpcPort" bson:"grpcPort" yaml:"grpcPort"`                                  // 若为空表示不启用grpc server
	GrpcKeepAlive       time.Duration `json:"grpcKeepAlive" bson:"grpcKeepAlive" yaml:"grpcKeepAlive"`                   // 默认不启用
	GrpcCheckTimeout    string        `json:"grpcCheckTimeout" bson:"grpcCheckTimeout" yaml:"grpcCheckTimeout"`
	GrpcCheckInterval   string        `json:"grpcCheckInterval" bson:"grpcCheckInterval" yaml:"grpcCheckInterval"`
}

// 合并默认值
func mergeConfig(conf *Config) *Config {

	if conf == nil {
		conf = &Config{}
	}

	// 补充默认逻辑
	if conf.HttpHost == "" {
		conf.HttpHost = PrivateAddress
	}
	if conf.GrpcHost == "" {
		conf.GrpcHost = PrivateAddress
	}
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
