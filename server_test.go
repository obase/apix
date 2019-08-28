package apix

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/obase/httpx/ginx"
	"testing"
	"time"
)

var demo = func(args []string) gin.HandlerFunc {
	return func(context *gin.Context) {
		fmt.Println("plugin arguments: ", args)
	}
}

func TestNewServer(t *testing.T) {
	server := NewServer()

	server.Plugin("demo", demo)

	server.Use(func(context *gin.Context) {
		fmt.Println("this is use 1...")
	}, func(context *gin.Context) {
		fmt.Println("this is use 2...")
	})

	server.Routes(func(server *ginx.Server) {
		g := server.Group("/g", func(context *gin.Context) {
			fmt.Println("this is in group...")
		})
		g.GET("/now", func(context *gin.Context) {
			fmt.Fprintf(context.Writer, "current time: %v\n", time.Now().Format("2006-01-02 15:04:05.777"))
		})
	})
	server.Serve()
}
