package apix

import (
	"context"
	"github.com/obase/log"
	"github.com/obase/pbx/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"net"
	"strconv"
)

type consulResolverWatcher struct {
	service   string
	address   map[string]bool
	lastIndex uint64
}

func (w *consulResolverWatcher) Next() ([]*naming.Update, error) {
	for {
		services, metainfo, err := consul.DiscoveryService(w.lastIndex, w.service)
		if err != nil {
			log.Errorf(context.Background(), "Retrieving consul service entries error: %v", err)
		}
		w.lastIndex = metainfo.LastIndex

		address := make(map[string]bool)
		for _, service := range services {
			address[net.JoinHostPort(service.Service.Address, strconv.Itoa(service.Service.Port))] = true
		}

		var updates []*naming.Update
		for addr := range w.address {
			if _, ok := address[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
			}
		}

		for addr := range address {
			if _, ok := w.address[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
			}
		}

		if len(updates) != 0 {
			w.address = address
			return updates, nil
		}
	}
}

func (w *consulResolverWatcher) Close() {
	// nothing to do
}

func (r *consulResolverWatcher) Resolve(target string) (naming.Watcher, error) {
	return r, nil
}

func DialConn(service string) (*grpc.ClientConn, error) {
	return grpc.Dial("", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithBalancer(
		grpc.RoundRobin(&consulResolverWatcher{
			service: service,
		})))
}
