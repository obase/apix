package apix

/*服务配置,注意兼容性.Grpc服务添加前缀"grpc."*/
type Config struct {
	Name                         string `json:"name" bson:"name" yaml:"name"`                                              // 服务名称
	HttpHost                     string `json:"httpHost" bson:"httpHost" yaml:"httpHost"`                                  // 默认本机扫描到的第一个私用IP
	HttpPort                     int    `json:"httpPort" bson:"httpPort" yaml:"httpPort"`                                  // 若为空表示不启用http server
	HttpKeepAlivePeriod          string `json:"httpKeepAlivePeriod" bson:"httpKeepAlivePeriod" yaml:"httpKeepAlivePeriod"` // 默认不启用
	GrpcHost                     string `json:"grpcHost" bson:"grpcHost" yaml:"grpcHost"`                                  // 默认本机扫描到的第一个私用IP
	GrpcPort                     int    `json:"grpcPort" bson:"grpcPort" yaml:"grpcPort"`                                  // 若为空表示不启用grpc server
	GrpcKeepAlivePeriod          string `json:"grpcKeepAlivePeriod" bson:"grpcKeepAlivePeriod" yaml:"grpcKeepAlivePeriod"` // 默认不启用
	WsReadBufferSize             int    `json:"wsReadBufferSize" bson:"wsReadBufferSize" yaml:"wsReadBufferSize"`          // 默认4092
	WsWriteBufferSize            int    `json:"wsWriteBufferSize" bson:"wsWriteBufferSize" yaml:"wsWriteBufferSize"`       // 默认4092
	WsNotCheckOrigin             bool   `json:"wsNotCheckOrigin" bson:"wsNotCheckOrigin" yaml:"wsNotCheckOrigin"`          // 默认false
	ConsulAddress                string `json:"consulAddress" bson:"consulAddress" yaml:"consulAddress"`                   // 默认127.0.0.1:8500, 如果设成0.0.0.0表示禁用consul服务
	ConsulCheckTimeoutHttp       string `json:"consulCheckTimeoutHttp" bson:"consulCheckTimeoutHttp" yaml:"consulCheckTimeoutHttp"`
	ConsulCheckIntervalHttp      string `json:"consulCheckIntervalHttp" bson:"consulCheckIntervalHttp" yaml:"consulCheckIntervalHttp"`
	ConsulCheckTimeoutGrpc       string `json:"consulCheckTimeoutGrpc" bson:"consulCheckTimeoutGrpc" yaml:"consulCheckTimeoutGrpc"`
	ConsulCheckIntervalGrpc      string `json:"consulCheckIntervalGrpc" bson:"consulCheckIntervalGrpc" yaml:"consulCheckIntervalGrpc"`
	ConsulDeregisterServiceAfter string `json:"consulDeregisterServiceAfter" bson:"consulDeregisterServiceAfter" yaml:"consulDeregisterServiceAfter"` // 默认30分钟
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
	if conf.ConsulCheckTimeoutHttp == "" {
		conf.ConsulCheckTimeoutHttp = "5s" // 兼容旧值
	}
	if conf.ConsulCheckIntervalHttp == "" {
		conf.ConsulCheckIntervalHttp = "6s" // 兼容旧值
	}
	if conf.ConsulCheckTimeoutGrpc == "" {
		conf.ConsulCheckTimeoutGrpc = "5s" // 兼容旧值
	}
	if conf.ConsulCheckIntervalGrpc == "" {
		conf.ConsulCheckIntervalGrpc = "6s" // 兼容旧值
	}
	if conf.ConsulDeregisterServiceAfter == "" {
		conf.ConsulDeregisterServiceAfter = "30m"
	}
	return conf
}
