package registry

type RegistryType int32

const (
	ETCD RegistryType = iota
	REDIS
	ZK
	CONSUL
	schema = "simple_rpc"
)

type Registry interface {
	Init(string,string) error  //init a registry
	Register(string,string) error  //register a service
	Unregister(string,string) error  //register a service
	GetService(string) (map[string]int, error)  //get service by name
	ListServices() (map[string]map[string]int,error)  //list all service on this registry
	Watch(string,chan *Event) (error)  //watch change on registry,get message from registry
	String() string  //get registry'name as "type[addr]"
}

type Event struct {
	Type int `0:keep;1:delete`
	Addr string
	Weight int
}


