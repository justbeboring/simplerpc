# simplerpc
a simple rpc-frame over grpc with registry(etcd/consul/zookeeper/redis) and load-balance(weighted-round-roubin/weighted-random).

server：

	package main

	import (
		"context"
		"flag"
		"google.golang.org/grpc/peer"
		"log"
		"main/pb"
		"github.com/justbeboring/simplerpc"
		"net"
	)

	type hello struct {}

	var addr = "127.0.0.1:50051"

	func (*hello) Echo(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
		pr,ok := peer.FromContext(ctx)
		if !ok{
			log.Println("get client ipaddr error")
		}else{
			if pr.Addr == net.Addr(nil) {
				log.Println("client ipaddr is nil")
			}else{
				log.Println("request from:" + pr.Addr.String())
			}
		}
		req.Data = "response from:" + addr
		return req, nil
	}

	func main(){
		flag.StringVar(&addr, "addr", addr, "addr to lis")
		flag.Parse()
		service := simplerpc.NewService("test",addr)
		service.AddRegistry("etcd","127.0.0.1:2379","")
		//service.AddRegistry("consul","127.0.0.1:8500","")
		//service.AddRegistry("zookeeper","127.0.0.1:2181","")
		//service.AddRegistry("redis","127.0.0.1:6379","")
		//service.SetCreds("tls/server.crt","tls/server.key")
		service.Init()
		pb.RegisterHelloServiceServer(service.GrpcServer, &hello{})
		service.Run()
	}

client：

	package main

	import (
		"context"
		"github.com/justbeboring/simplerpc"
		"google.golang.org/grpc"
		"google.golang.org/grpc/resolver"
		"log"
		"main/pb"
		"time"
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
