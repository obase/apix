package apix

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/httpx/ginx"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := NewServer()
	server.Routes(func(server *ginx.Server) {
		server.GET("/now", func(context *gin.Context) {
			fmt.Fprintf(context.Writer, "current time: %v\n", time.Now().Format("2006-01-02 15:04:05.777"))
		})
	})
	server.Serve()
}

func test(args ...interface{}) {
	fmt.Printf("%v", args == nil)
}
