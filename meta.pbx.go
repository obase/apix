package pbx

import (
	"context"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

/*方法处理原型*/
type MethodFunc func(ctx context.Context, rdata []byte) (interface{}, error)

/*对应proto里面option数据结构,在各个RegisterXXXService()自动注册*/
type Meta struct {
	init         map[string]bool // file初始化标志
	serverOption []grpc.ServerOption
	middleFilter []gin.HandlerFunc
	services     []*Service
}

func NewMeta() *Meta {
	return &Meta{
		init: make(map[string]bool),
	}
}

func (m *Meta) ServerOption(so grpc.ServerOption) {
	m.serverOption = append(m.serverOption, so)
}

func (m *Meta) MiddleFilter(mf gin.HandlerFunc) {
	m.middleFilter = append(m.middleFilter, mf)
}

func (m *Meta) Service(desc *grpc.ServiceDesc, impl interface{}) *Service {
	gs := &Service{
		serviceDesc: desc,
		serviceImpl: impl,
	}
	m.services = append(m.services, gs)
	return gs
}

type Service struct {
	serviceImpl interface{}
	serviceDesc *grpc.ServiceDesc
	groupPath   string
	groupFilter []gin.HandlerFunc
	methods     []*Method
}

func (gs *Service) GroupPath(gpath string) {
	gs.groupPath = gpath
}

func (gs *Service) GroupFilter(gf gin.HandlerFunc) {
	gs.groupFilter = append(gs.groupFilter, gf)
}

func (gs *Service) Method(tag string, adapt MethodFunc) *Method {
	gm := &Method{
		tag:     tag,
		adapter: adapt,
	}
	gs.methods = append(gs.methods, gm)
	return gm
}

type Method struct {
	tag          string
	adapter      MethodFunc        // 对应方法的AdapterFunc
	handlePath   string            // 对应方法的Handler path
	handleFilter []gin.HandlerFunc // 对应方法的Handler Filter
	socketPath   string            // 对应方法的Socket path
	socketFilter []gin.HandlerFunc // 对应方法的Handler Filter
}

func (gm *Method) HandlePath(path string) {
	gm.handlePath = path
}

func (gm *Method) HandleFilter(hf gin.HandlerFunc) {
	gm.handleFilter = append(gm.handleFilter, hf)
}

func (gm *Method) SocketPath(path string) {
	gm.socketPath = path
}

func (gm *Method) SocketFilter(hf gin.HandlerFunc) {
	gm.socketFilter = append(gm.socketFilter, hf)
}
