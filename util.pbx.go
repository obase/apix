package pbx

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
	"net"
	"net/http"
)

// TODO: 附加访问时长
func RecoverHandleFunc(c *gin.Context) {
	if perr := recover(); perr != nil {
		Errorf(c, "recover error: %v", perr)
	}
}

func CreateHandleFunc(mf MethodFunc, tag string) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer RecoverHandleFunc(c)

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
				Errorf(c, "%s execute service: %v", tag, err)
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
			Errorf(c, "%s reading request: %v", tag, err)
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

func CreateSocketFunc(upgrader *websocket.Upgrader, af MethodFunc, tag string) gin.HandlerFunc {
	return func(c *gin.Context) {

		defer RecoverHandleFunc(c)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			Errorf(c, "upgrade connection: %v", tag, err)
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
				Errorf(c, "%s reading message: %v", tag, err)
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
				Errorf(c, "%s execute service: %v", tag, err)
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
				Errorf(c, "%s writing message: %v", tag, err)
				return
			}
		}
	}
}

// 创建upgrader
func CreateSocketUpgrader(conf *Conf) *websocket.Upgrader {
	upgrader := new(websocket.Upgrader)
	if conf.WsReadBufferSize != 0 {
		upgrader.ReadBufferSize = conf.WsReadBufferSize
	}
	if conf.WsWriteBufferSize != 0 {
		upgrader.WriteBufferSize = conf.WsWriteBufferSize
	}
	if conf.WsNotCheckOrigin {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
	}
	return upgrader
}

var PrivateAddress = func(def string) (ret string) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return def
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ip4 := ipnet.IP.To4(); ip4 != nil {
				// 必须私有网段
				if (ip4[0] == 10) || (ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || (ip4[0] == 192 && ip4[1] == 168) {
					return ip4.String()
				}
			}
		}
	}
	return def
}("127.0.0.1")
