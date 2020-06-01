package balance

import (
	"errors"
	"fmt"
	"morefruit/base/grpc-lb/registry/consul"
	"net"
	"sync"

	capi "github.com/hashicorp/consul/api"
)

/*
	节点启动命令：
		“consul agent -node=agent-three -bind=192.168.1.5 -enable-local-script-checks=true
		-data-dir=/tmp/consul -join=192.168.1.6“
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

func (cb *ConsulBalance) RegisterServerIntoConsul(config *Config) error {
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
	errChan := make(chan error, 1)
	go func(errChan chan<- error) {
		cb.register.Register(errChan)
		cb.wg.Done()
	}(errChan)
	return <-errChan
}

func (cb *ConsulBalance) UnRegister() {
	cb.register.Unregister()
	cb.wg.Wait()
	fmt.Println("ConsulBalance UnRegister")
}
