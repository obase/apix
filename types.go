package apix

import "github.com/gin-gonic/gin"

type (
	Context = gin.Context
	IRoutes = gin.IRoutes
	HandlerFunc = gin.HandlerFunc
	RouterGroup = gin.RouterGroup
	IRouter interface {
		IRoutes
		Group(string, ...HandlerFunc) *RouterGroup
	}
)
