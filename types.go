package apix

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

/*
通过type别名解决vendor内部gin api的问题!
*/
type (
	Context = gin.Context
	IRoutes = gin.IRoutes
	HandlerFunc = gin.HandlerFunc
	RouterGroup = gin.RouterGroup
	IRouter = gin.IRouter
)
/*
通过type别名解决vendor内部grpc api的问题!
*/
type (
	Server = grpc.Server
)
