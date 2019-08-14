package apix

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
)

func graceListenGrpc(host string, port int) (net.Listener, error) {
	return net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
}

func graceListenHttp(host string, port int) (net.Listener, error) {
	return net.Listen("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
}

func graceShutdownOrRestart(grpcServer *grpc.Server, grpcListener net.Listener, httpServer *http.Server, httpListener net.Listener) {
	sch := make(chan os.Signal, 1)
	defer signal.Stop(sch)

	signal.Notify(sch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	for {
		sig := <-sch

		switch sig {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM:
			ws := new(sync.WaitGroup)
			if httpServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					httpServer.Shutdown(context.Background())
				}(ws)
			}
			if grpcServer != nil {
				ws.Add(1)
				go func(ws *sync.WaitGroup) {
					defer ws.Done()
					grpcServer.GracefulStop()
				}(ws)
			}
			ws.Wait()
			return
		}
	}
}
