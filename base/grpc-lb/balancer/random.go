package balancer

import (
	"context"
	"github.com/liyue201/grpc-lb/common"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"math/rand"
	"sync"
)

const Random = "random"

// newRandomBuilder creates a new random balancer builder.
func newRandomBuilder() balancer.Builder {
	return base.NewBalancerBuilderWithConfig(Random, &randomPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newRandomBuilder())
}

type randomPickerBuilder struct{}

func (*randomPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	grpclog.Infof("randomPicker: newPicker called with readySCs: %v", readySCs)
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	var scs []balancer.SubConn

	for addr, sc := range readySCs {
		weight := common.GetWeight(addr)
		for i := 0; i < weight; i++ {
			scs = append(scs, sc)
		}
	}

	return &randomPicker{
		subConns: scs,
	}
}

type randomPicker struct {
	subConns []balancer.SubConn
	mu       sync.Mutex
}

func (p *randomPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	p.mu.Lock()
	sc := p.subConns[rand.Intn(len(p.subConns))]
	p.mu.Unlock()
	return sc, nil, nil
}
