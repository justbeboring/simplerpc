package balancer

import (
	"sync"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

const ROUND = "weightedround"

//Peer 单个配置节点
type Peer struct {
	Sockaddr        string    //id地址
	SubConn balancer.SubConn  //连接
	CurrentWeight   int //当前权重
	EffectiveWeight int //有效权重
	Weight          int //配置权重
}

//Peers 服务器地址池
type Peers struct {
	number      int     //服务器数量
	TotalWeight int     //总权重
	peer        []*Peer //服务器节点数组
}

type roundPicker struct {
	Peers *Peers //服务器池数据
	mu sync.Mutex
}

func (p *roundPicker)Pick(balancer.PickInfo)(balancer.PickResult,error){
	p.mu.Lock()
	peer := p.GetPeer()
	p.mu.Unlock()
	return balancer.PickResult{SubConn:peer.SubConn},nil
}

type roundPickerBuilder struct {}

func(*roundPickerBuilder)Build(info base.PickerBuildInfo)balancer.V2Picker{
	grpclog.Infof("new picker called with info:%v",info)
	if len(info.ReadySCs) == 0{
		return base.NewErrPickerV2(balancer.ErrNoSubConnAvailable)
	}
	peers := &Peers{}
	for subConn,addr := range info.ReadySCs{
		node := GetAddrInfo(addr.Address)
		peer := Peer{Sockaddr:addr.Address.Addr,
			SubConn:subConn,
		EffectiveWeight:node.Weight,
		Weight:node.Weight,
		}
		peers.peer = append(peers.peer,&peer)
		peers.number += 1
	}
	return &roundPicker{
		Peers:peers,
	}
}

func (pd *roundPicker)GetPeer() *Peer {
	var best *Peer
	total := 0

	for i := 0; i < pd.Peers.number; i++ {
		peer := pd.Peers.peer[i]
		peer.CurrentWeight += peer.EffectiveWeight //将当前权重与有效权重相加

		total += peer.EffectiveWeight //累加总权重

		if best == nil || peer.CurrentWeight > best.CurrentWeight {
			best = peer
		}

	}
	if best == nil {
		return nil
	}
	best.CurrentWeight -= total //将当前权重改为当前权重-总权重
	return best
}

func newRoundBuilder() balancer.Builder {
	return base.NewBalancerBuilderV2(ROUND, &randomPickerBuilder{}, base.Config{HealthCheck: false})
}