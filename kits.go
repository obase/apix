package apix

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
	"github.com/obase/log"
	"net/http"
)

const (
	GRACE_ENV  = "_GRC_"
	GRACE_NONE = "0"
	GRACE_GRPC = "1"
	GRACE_HTTP = "2"
	GRACE_ALL  = "3" // grpc是3, http是4
)

// TODO: 附加访问时长
func recoverHandleFunc(c *gin.Context) {
	if perr := recover(); perr != nil {
		log.ErrorStack(c, fmt.Errorf("panic error: uri=%v, err=%v", c.Request.RequestURI, perr), false) // 打印堆栈错误
	}
}

func createHandleFunc(mf MethodFunc, tag string) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer recoverHandleFunc(c)

		var (
			rdata []byte
			wdata []byte
			rsp   interface{}
			err   error
		)
		rdata, err = c.GetRawData()
		if err == nil {
			rsp, err = mf(c, rdata)
			if err == nil {
				wdata, _ = json.Marshal(&api.Response{
					Code: api.SUCCESS,
					Data: rsp,
					Tag:  tag,
				})
			} else {
				log.Error(c, "%s execute service: %v", tag, err)
				if ersp, ok := err.(*api.Response); ok {
					wdata, _ = json.Marshal(ersp)
				} else {
					wdata, _ = json.Marshal(&api.Response{
						Code: api.EXECUTE_SERVICE_ERROR,
						Msg:  err.Error(),
						Tag:  tag,
					})
				}
			}
		} else {
			log.Error(c, "%s reading request: %v", tag, err)
			wdata, _ = json.Marshal(&api.Response{
				Code: api.READING_REQUEST_ERROR,
				Msg:  err.Error(),
				Tag:  tag,
			})
		}
		c.Writer.Header()["Content-Type"] = api.JsonContentType
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write(wdata)
	}
}

func createSocketFunc(upgrader *websocket.Upgrader, af MethodFunc, tag string) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer recoverHandleFunc(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error(c, "upgrade connection: %v", tag, err)
			return
		}
		for {
			var (
				mtype int
				rdata []byte
				wdata []byte
				rsp   interface{}
				err   error
			)
			mtype, rdata, err = conn.ReadMessage()
			if err != nil {
				log.Error(c, "%s reading message: %v", tag, err)
				return
			}
			rsp, err = af(c, rdata)
			if err == nil {
				wdata, _ = json.Marshal(&api.Response{
					Code: api.SUCCESS,
					Data: rsp,
					Tag:  tag,
				})
			} else {
				log.Error(c, "%s execute service: %v", tag, err)
				if ersp, ok := err.(*api.Response); ok {
					wdata, _ = json.Marshal(ersp)
				} else {
					wdata, _ = json.Marshal(&api.Response{
						Code: api.EXECUTE_SERVICE_ERROR,
						Msg:  err.Error(),
						Tag:  tag,
					})
				}
			}
			err = conn.WriteMessage(mtype, wdata)
			if err != nil {
				log.Error(c, "%s writing message: %v", tag, err)
				return
			}
		}
	}
}

// 创建upgrader
func createSocketUpgrader(conf *Config) *websocket.Upgrader {
	upgrader := new(websocket.Upgrader)
	if conf.WbskReadBufferSize != 0 {
		upgrader.ReadBufferSize = conf.WbskReadBufferSize
	}
	if conf.WbskWriteBufferSize != 0 {
		upgrader.WriteBufferSize = conf.WbskWriteBufferSize
	}
	if conf.WbskNotCheckOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	return upgrader
}

func Errorf(code int, format string, args ...interface{}) error {
	return &api.Response{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
}
