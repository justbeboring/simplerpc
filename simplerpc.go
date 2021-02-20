package simplerpc

import (
	"github.com/justbeboring/simplerpc/balancer"
	regtry "github.com/justbeboring/simplerpc/registry"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/resolver"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

const schema = "simple_rpc"

type Service struct {
	Name       string
	Version    string
	Addr       string
	Registries []regtry.Registry                //registery list
	GrpcServer *grpc.Server                     //grpcserver
	Listener   net.Listener                     //socket listener
	creds      credentials.TransportCredentials //ssl creds
}

type Resolver struct {
	Name       string
	Registries []regtry.Registry //registery list
	Nodes      []*Node           //server addr list
	cc         resolver.ClientConn
	msgChan    chan *regtry.Event               //channel to recieve registry's message
	creds      credentials.TransportCredentials //ssl creds
	Conn       *grpc.ClientConn                 //grpc client
}

type Node struct {
	Addr   string
	Weight int
	Des    string
}

func NewService(name,addr string)Service{
	return Service{Name:name,Addr:addr}
}

////register in all registry
func(service *Service)RegisterAll(){
	var err error
		for _,registry := range service.Registries{
			err = registry.Register(service.Name,service.Addr)
			if err != nil{
				log.Println(err)
			}
	}
}

////add registry in service's registries
func(service *Service)AddRegistry(registryType string, addr string, des string)error{
	registry,err := NewRegistry(registryType, addr, des)
	if err != nil{
		registerFailed(registryType,addr)
		log.Println(err)
		return err
	}
	service.Registries = append(service.Registries,registry)
	return nil
}

////to log register failed message
func registerFailed(registryType string,addr string){
		log.Println(fmt.Sprintf("$s[%s] register failed!",registryType,addr))
}

//// unregister in all registry
func(service *Service)UnregisterAll()error{
	var err error
	for _,registry := range service.Registries{
		err = registry.Unregister(service.Name,service.Addr)
	}
	return err
}

////set ssh creds
func(service *Service)SetCreds(certFile, keyFile string)error{
	var err error
	service.creds,err = credentials.NewServerTLSFromFile(certFile, keyFile)
	return  err
}

////init service
func(service *Service)Init(){
	service.RegisterAll()
	if service.creds == nil {
		service.GrpcServer = grpc.NewServer()
	}else{
		service.GrpcServer = grpc.NewServer(grpc.Creds(service.creds))
	}
	var err error
	service.Listener, err = net.Listen("tcp", service.Addr)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
}

////start service
func(service *Service)Run()error{
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		service.UnregisterAll()
		service.GrpcServer.GracefulStop()
		service.Listener.Close()
		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}
	}()

	if err := service.GrpcServer.Serve(service.Listener); err != nil {
		log.Fatalf("failed to serve: %s", err)
		return err
	}
	return nil
}

////create a registry(etcd/consul/redis/zookeeper)
func NewRegistry(registryType string, addr string, des string) (regtry.Registry, error) {
	var registry regtry.Registry
	var err error
	switch registryType {
	case "etcd":
		registry = new(regtry.Etcd)
	case "consul":
		registry = new(regtry.Consul)
	case "zookeeper":
		registry = new(regtry.Zk)
	case "redis":
		registry = new(regtry.Redis)
	default:
		return nil,err
	}
	err = registry.Init(addr, des)
	return registry,err
}

func NewResolver(name string)*Resolver{
	return &Resolver{Name:name}
}

func(r *Resolver)AddRegistry(registryType string, addr string, des string)error{
	registry,err := NewRegistry(registryType, addr, des)
	if err != nil{
		log.Println(err)
		return err
	}
	r.Registries = append(r.Registries,registry)
	return nil
}

func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
    r.cc = cc

	go r.watch("/" + target.Scheme + "/" + target.Endpoint)

	return r, nil
}

func (r *Resolver) Scheme() string {
	return schema
}

func (r *Resolver) ResolveNow(rn resolver.ResolveNowOptions) {
	log.Println("ResolveNow") // TODO check
}

// Close closes the resolver.
func (r *Resolver) Close() {
	log.Println("Close")
}

func (r *Resolver) watch(keyPrefix string) {
	var addrList []resolver.Address

	r.msgChan = make(chan *regtry.Event,100)
	for _, registry := range r.Registries {
		registry.Watch(keyPrefix, r.msgChan)
	}

	for {
		msg := <-r.msgChan
		if msg.Type == 0 {
			if !exist(addrList, msg.Addr) {
				addrList = append(addrList, balancer.SetAddrInfo(resolver.Address{Addr:msg.Addr},balancer.AddrInfo{msg.Weight}))
				r.cc.NewAddress(addrList)
			}
		} else {
			remove(addrList, msg.Addr)
			r.cc.NewAddress(addrList)
		}
	}
}

func (r *Resolver)SetCreds(certFile, keyFile string)error{
	var err error
	r.creds,err = credentials.NewClientTLSFromFile(certFile, keyFile)
	return err
}

func (r *Resolver)Init()error{
	var err error
	if r.creds == nil {
		//r.Conn, err = grpc.Dial(r.Scheme()+"://author/test", grpc.WithBalancerName("round_robin"), grpc.WithInsecure())
		r.Conn, err = grpc.Dial(r.Scheme()+"://author/test", grpc.WithBalancerName("weightedround"), grpc.WithInsecure())
	}else{
		//r.Conn, err = grpc.Dial(r.Scheme()+"://author/test", grpc.WithBalancerName("round_robin"), grpc.WithTransportCredentials(r.creds))
		r.Conn, err = grpc.Dial(r.Scheme()+"://author/test", grpc.WithBalancerName("weightedround"), grpc.WithTransportCredentials(r.creds))
	}
	return err
}

func exist(l []resolver.Address, addr string) bool {
	for i := range l {
		if l[i].Addr == addr {
			return true
		}
	}
	return false
}

func remove(s []resolver.Address, addr string) ([]resolver.Address, bool) {
	for i := range s {
		if s[i].Addr == addr {
			s[i] = s[len(s)-1]
			return s[:len(s)-1], true
		}
	}
	return nil, false
}



