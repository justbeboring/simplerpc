package balancer

import (
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/balancer"
)

type attributeKey struct{}

type AddrInfo struct {
	Weight int
}

func SetAddrInfo(addr resolver.Address,addrInfo AddrInfo)resolver.Address{
	addr.Attributes = attributes.New()
	addr.Attributes = addr.Attributes.WithValues(attributeKey{},addrInfo)
	return addr
}

func GetAddrInfo(addr resolver.Address)AddrInfo{
	v := addr.Attributes.Value(attributeKey{})
	ai,_ := v.(AddrInfo)
	return ai
}

func init(){
	balancer.Register(newRandomBuilder())
	balancer.Register(newRoundBuilder())
}