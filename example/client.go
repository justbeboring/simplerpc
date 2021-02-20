package main

import (
	"log"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc"
	"context"
	"main/pb"
	"time"
	"github.com/justbeboring/simplerpc/registry"
	"github.com/justbeboring/simplerpc"
)

func main() {
	r := simplerpc.NewResolver("test")

	r.AddRegistry(registry.ETCD,"localhost:2379","")
	//r.AddRegistry(registry.CONSUL,"11.36.208.249:8500","")
	//r.AddRegistry(registry.ZK,"11.36.208.249:2181","")
	//r.AddRegistry(registry.REDIS, "11.36.208.249:6379", "")
	//r.SetCreds("tls/server.crt","server.grpc.io")

	resolver.Register(r)
	var err error
	//r.Conn, err = grpc.Dial(r.Scheme()+"://author/test", grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
	r.Init()
	if err != nil {
		panic(err)
	}

	client := pb.NewHelloServiceClient(r.Conn)

	count := 0
	time.Sleep(1*time.Second)
	for i :=0;i< 20;i++{
		resp, err := client.Echo(context.Background(), &pb.Payload{Data: "hello"}, grpc.FailFast(true))
		if err != nil {
			log.Println(err)
		} else {
			log.Println(resp.Data)
		}
		count++
	}

}
