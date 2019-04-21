package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
)

// RPCDuplex represents a RPC Duplex implementation where both ends of the connection
// has a rpc.Server and a rpc.Client.
type RPCDuplex struct {
	net.Conn
	*rpc.Client
	*rpc.Server
}

// RPCMethod is a receiver which we will use Register to publishes the receiver's methods in the DefaultServer.
type RPCMethod struct{}

// Person is a struct that A will use to expose it's RPC method
type Person struct {
	Name string
}

// SayHello is a RPC method
// RPC methods must look schematically like: func (t *T) MethodName(argType T1, replyType *T2) error
func (RPCMethod) SayHello(person Person, reply *Person) error {
	*reply = person
	return nil
}

// NewRPCDuplex takes in a single net.Conn and returns a RPC Duplex construct
func NewRPCDuplex(conn net.Conn) *RPCDuplex {
	return &RPCDuplex{conn, rpc.NewClient(conn), rpc.NewServer()}
}

// Register registers an object in the server, making it visible as a service with the name of the type of the object.
func (d *RPCDuplex) Register(obj *RPCMethod) {
	d.Server.Register(obj)
}

// Serve serves the rpc.Server via net.Conn.
func (d *RPCDuplex) Serve() {
	d.Server.ServeConn(d.Conn)
}

func main() {

	object := new(RPCMethod)

	connA, connB := net.Pipe()
	defer connA.Close()
	defer connB.Close()

	go func() {
		svr := NewRPCDuplex(connA)
		svr.Register(object)
		svr.Serve()
	}()

	// Client
	var reply Person
	testInput := Person{"Anto"}

	clientA := NewRPCDuplex(connB)
	err := clientA.Call("RPCMethod.SayHello", testInput, &reply)

	if err != nil {
		log.Fatal("error", err)
	}

	fmt.Println(reply.Name)

}

// All the other members needed should be made available from the embedded structures.

// ==================
// BACKGROUND
// ==================
// Currently, communication between manager-node and skywire-node is done via a rpc.
// Client to rpc.Server relationship. However, in the future, skywire-node will need
// to initiate communication with manager-node (e.g. for notifications, logging and
// Skywire App to Manager communication).

// ==================
// DESCRIPTION
// ==================
// The netutil module is to be a shared library to aid communication between Skywire
// services. Communication can be either via Transport (or higher level interfaces),
// or noise.Conn.

// The first structure to be implemented is netutil.RPCDuplex. This structure implements
// an RPC Duplex connection via a single net.Conn implementation. In other words, both
// ends of the connection has a rpc.Server and a rpc.Client.

// Task 1: Implement RPCDuplex (as specified above).
// Task 2: Write tests using net.Pipe() and having two RPCDuplex instances communicate with one another.
