package apix

import (
	"context"
	"github.com/gin-gonic/gin"
)

/*方法处理原型*/
type MethodFunc func(ctx context.Context, rdata []byte) (interface{}, error)

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
