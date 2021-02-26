package main

import (
	"context"
	"github.com/justbeboring/simplerpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"log"
	"main/pb"
)

func main() {
	r := simplerpc.NewResolver("test")

	r.AddRegistry("etcd","127.0.0.1:2379","")
	//r.AddRegistry(registry.CONSUL,"11.36.208.249:8500","")
	//r.AddRegistry(registry.ZK,"11.36.208.249:2181","")
	//r.AddRegistry(registry.REDIS, "11.36.208.249:6379", "")
	//r.SetCreds("tls/server.crt","server.grpc.io")

	resolver.Register(r)
	var err error
	r.Init()
	if err != nil {
		panic(err)
	}

	client := pb.NewHelloServiceClient(r.Conn)
	
	for i :=0;i< 20;i++{
		resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
		if err != nil {
			log.Println(err)
		} else {
			log.Println(resp.Data)
		}
	}

}
