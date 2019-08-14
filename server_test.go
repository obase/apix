package apix

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	server.Route(func(router gin.IRouter) {
		router.GET("/now", func(context *gin.Context) {
			fmt.Fprintf(context.Writer, "current time: %v\n", time.Now().Format("2006-01-02 15:04:05.777"))
		})
	})
	server.Serve()
}

func test(args ...interface{}) {
	fmt.Printf("%v", args == nil)
}
