package apix

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/obase/api"
	"github.com/obase/api/x"
	"github.com/obase/log"
	"net"
	"net/http"
	"time"
)

const (
	GRACE_ENV  = "_GRC_"
	GRACE_NONE = "0"
	GRACE_GRPC = "1"
	GRACE_HTTP = "2"
	GRACE_ALL  = "3" // grpc是3, http是4
)

// TODO: 附加访问时长
func RecoverHandleFunc(c *gin.Context) {
	if perr := recover(); perr != nil {
		log.ErrorStack(c, fmt.Errorf("panic error: uri=%v, err=%v", c.Request.RequestURI, perr), false) // 打印堆栈错误
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
				log.Errorf(c, "%s execute service: %v", tag, err)
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
			log.Errorf(c, "%s reading request: %v", tag, err)
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
			log.Errorf(c, "upgrade connection: %v", tag, err)
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
				log.Errorf(c, "%s reading message: %v", tag, err)
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
				log.Errorf(c, "%s execute service: %v", tag, err)
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
				log.Errorf(c, "%s writing message: %v", tag, err)
				return
			}
		}
	}
}

// 创建upgrader
func CreateSocketUpgrader(conf *Config) *websocket.Upgrader {
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

func Errorf(code int, format string, args ...interface{}) error {
	return &api.Response{
		Code: code,
		Msg:  fmt.Sprintf(format, args...),
	}
}

var None = new(x.Void) // 定义空值

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
	keepAlivePeriod time.Duration
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	if ln.keepAlivePeriod > 0 {
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(ln.keepAlivePeriod)
	} else {
		// 用回http.Server默认设置
		tc.SetKeepAlive(true)
		tc.SetKeepAlivePeriod(3 * time.Minute)
	}
	return tc, nil
}
