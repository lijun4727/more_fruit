package balance

import (
	"errors"
	"morefruit/base/grpc-lb/registry/consul"
	"net"

	capi "github.com/hashicorp/consul/api"
)

var (
	QueryIpError = errors.New("can not query ip")
)

type Config struct {
	ConsulAddr    string
	ConsulSrvName string
	NodeID        string
	Port          int
	Weight        string
}

func RegisterServerIntoConsul(config *Config) (*consul.Registrar, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
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
		return nil, QueryIpError
	}

	consulCfg := &capi.Config{
		Address: config.ConsulAddr,
	}

	register, err := consul.NewRegistrar(
		&consul.Congfig{
			ConsulCfg:   consulCfg,
			ServiceName: config.ConsulSrvName,
			NData: consul.NodeData{
				ID:       config.NodeID,
				Address:  srvAddr,
				Port:     config.Port,
				Metadata: map[string]string{"weight": config.Weight},
			},
			Ttl: 5,
		})
	if err != nil {
		return nil, err
	}
	err = register.Register()
	return register, err
}
