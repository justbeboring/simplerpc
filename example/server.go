package main

import (
	"context"
	"flag"
	"main/pb"
	"github.com/justbeboring/simplerpc/registry"
	"github.com/justbeboring/simplerpc"
)

type hello struct {
}

var addr = "locahost:50051"

func (*hello) Echo(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
	//pr,ok := peer.FromContext(ctx)
	//if !ok{
	//	log.Println("get client ipaddr error")
	//}else{
	//	if pr.Addr == net.Addr(nil) {
	//		log.Println("client ipaddr is nil")
	//	}else{
	//		log.Println("request from:" + pr.Addr.String())
	//	}
	//}
	//req.Data = "response from:" + addr
	req.Data = "from " + addr
	return req, nil
}

func main(){
	flag.StringVar(&addr, "addr", addr, "addr to lis")
	flag.Parse()
	service := simplerpc.NewService("test",addr)
	service.AddRegistry(registry.ETCD,"localhost:2379","")
	//service.AddRegistry(registry.CONSUL,"11.36.208.249:8500","")
	//service.AddRegistry(registry.ZK,"11.36.208.249:2181","")
	//service.AddRegistry(registry.REDIS,"11.36.208.249:6379","")
	//service.SetCreds("tls/server.crt","tls/server.key")

    service.Init()
	pb.RegisterHelloServiceServer(service.GrpcServer, &hello{})
	service.Run()
}