package balancer

import (
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/grpclog"
	"sync"
	"math/rand"
	"container/ring"
)

const RANDOM = "weightedrandom"

type randomPicker struct {
	r ring.Ring
	subConn []balancer.SubConn
	mu sync.Mutex
}

func (p *randomPicker)Pick(balancer.PickInfo)(balancer.PickResult,error){
	p.mu.Lock()
	index := rand.Intn(len(p.subConn))
	sc := p.subConn[index]
	p.mu.Unlock()
	return balancer.PickResult{SubConn:sc},nil
}

type randomPickerBuilder struct {}

func(*randomPickerBuilder)Build(info base.PickerBuildInfo)balancer.V2Picker{
	grpclog.Infof("new picker called with info:%v",info)
	if len(info.ReadySCs) == 0{
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}
	var scs []balancer.SubConn
	for subConn,addr := range info.ReadySCs{
		node := GetAddrInfo(addr.Address)
		if node.Weight <= 0{
			node.Weight = 0
		}else if node.Weight > 5{
			node.Weight = 5
		}
		for i := 0;i<node.Weight;i++{
			scs = append(scs,subConn)
		}
	}
	return &randomPicker{
		subConn:scs,
	}
}

func newRandomBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(RANDOM, &randomPickerBuilder{}, base.Config{HealthCheck: false})
}