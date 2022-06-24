package main
import (
	"fmt"
	//"time"
	"net"
	"flag"
)
import "github.com/wutianze/nats.go"

var host_port = flag.String("socket listen address",":8000","listen on which port")
var nats_address = flag.String("nats","nats://39.101.140.145:4222","address of nats server")
// name: server, cli0, cli1, cli2
var name = flag.String("name","","who am I")
func main(){
flag.Parse()
	tcpServer, err0 := net.ResolveTCPAddr("tcp4",*host_port)
if err0 != nil{
fmt.Println(err0)
return
}
listener, err1:=net.ListenTCP("tcp",tcpServer)
if err1 != nil{
fmt.Println(err1)
return
}

nc, err2:=nats.IConnect(*nats_address)
defer nc.IClose()
if err2 != nil{
fmt.Println(err2)
return
}


for {
	socket_conn,err2 := listener.Accept()
	defer socket_conn.Close()
	if err2 != nil{
fmt.Println(err2)
continue
	}
	switch{
		case *name=="server":
_, err3:=nc.ISubscribe("gt_log", func(m *nats.Msg) {
	socket_conn.Write(m.Data)
		fmt.Printf("Received a message: %s\n", string(m.Data))
	})
if err3 != nil{
fmt.Println(err3)
}
case *name=="client0":
	fmt.Println("cli0");
	}

}


//nc.IPublish("gt_log",[]byte("what"))
//time.Sleep(time.Duration(5)*time.Second)
}
