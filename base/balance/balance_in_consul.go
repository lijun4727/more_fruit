package balance

import (
	"errors"
	"fmt"
	"morefruit/third_party/grpc-lb/registry/consul"
	"net"
	"sync"

	capi "github.com/hashicorp/consul/api"
)

/*
	节点启动命令：
	客户端:consul agent -node=agent-three -bind=192.168.1.5 -enable-local-script-checks=true -data-dir=/tmp/consul -join=192.168.1.6
	服务端:consul agent -server -bootstrap-expect=1 -node=agent-one -bind=192.168.1.6 -data-dir=/tmp/consul
*/

var (
	QueryIpError = errors.New("can not query ip")
)

type Config struct {
	ConsulAddr    string
	ConsulSrvName string
	NodeID        string
	Port          int
	Weight        string
	Ttl           int
}

type ConsulBalance struct {
	wg       sync.WaitGroup
	register *consul.Registrar
}

func (cb *ConsulBalance) Register(config *Config) error {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return err
	}

	var srvAddr string
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				srvAddr = ipnet.IP.String()
				break
			}
		}
	}
	if len(srvAddr) == 0 {
		return QueryIpError
	}

	consulCfg := &capi.Config{
		Address: config.ConsulAddr,
	}

	cb.register, err = consul.NewRegistrar(
		&consul.Congfig{
			ConsulCfg:   consulCfg,
			ServiceName: config.ConsulSrvName,
			NData: consul.NodeData{
				ID:       config.NodeID,
				Address:  srvAddr,
				Port:     config.Port,
				Metadata: map[string]string{"weight": config.Weight},
			},
			Ttl: config.Ttl,
		})
	if err != nil {
		return err
	}

	cb.wg = sync.WaitGroup{}
	cb.wg.Add(1)
	result := make(chan error, 1)
	go func(result chan<- error) {
		cb.register.Register(result)
		cb.wg.Done()
	}(result)
	return <-result
}

func (cb *ConsulBalance) UnRegister() {
	cb.register.Unregister()
	cb.wg.Wait()
	fmt.Println("ConsulBalance UnRegister")
}
