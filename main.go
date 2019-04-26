package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"time"
)

// Question:
// Data -> PrefixConn.Write() (adds prefix 0) -> OriginalConn.Write()
// Where is this data coming from? How do I retrieve and look at this data to be used with PrefixConn.Write so I can add prefix 0.
// e.g in traditional interfaces:
// I would declare var s interface then assign a struct to the interface
// I would be able to declare var netConn and then assign prefixConn{prefix, writeConn, readBuff} into this netConn in main...

// API is a receiver which we will use Register to publishes the receiver's methods in the DefaultServer.
type API struct{}

// Person is a struct that A will use to expose it's RPC method
type Person struct {
	Name string
}

// SayHello is a RPC method
// RPC methods must look schematically like: func (t *T) MethodName(argType T1, replyType *T2) error
func (API) SayHello(person Person, reply *Person) error {
	*reply = person
	return nil
}

// PrefixedConn will inherit original net.Conn connection from the interface. Duplicate original net.Conn
type PrefixedConn struct {
	prefix    byte
	writeConn io.Writer    // Original connection. net.Conn has a write method therefore it implements the writer interface.
	readBuf   bytes.Buffer // Read data from original connection. It is RPCDuplex's responsibility to push to here.
}

// func WriteString(w Writer, s string) (n int, err error)
// io.WriteString(os.Stdout, "Hello World")

// Any type that has this method will implement the Writer interface.
// type Writer interface {
// 	Write(p []byte) (n int, err error)
// }

// https://golang.org/src/bytes/buffer.go
// https://www.kancloud.cn/digest/batu-go/153538

// RPCDuplex is
type RPCDuplex struct {
	clientConn *PrefixedConn
	serverConn *PrefixedConn
}

// NewRPCDuplex is
func NewRPCDuplex(conn net.Conn, initiator bool) *RPCDuplex {
	var d RPCDuplex
	var a bytes.Buffer

	// PrefixedConn implements net.Conn and assigned it to d.clientConn and d.serverConn
	if initiator {
		d.clientConn = &PrefixedConn{prefix: 0, writeConn: conn, readBuf: a}
		d.serverConn = &PrefixedConn{prefix: 1, writeConn: conn, readBuf: a}
	} else {
		d.clientConn = &PrefixedConn{prefix: 1, writeConn: conn, readBuf: a}
		d.serverConn = &PrefixedConn{prefix: 0, writeConn: conn, readBuf: a}
	}

	// conn is rootConn

	return &d
}

// All the members in net.Conn interfaces --> https://golang.org/src/net/net.go
// type Conn interface {
// 	Read(b []byte) (n int, err error)
// 	Write(b []byte) (n int, err error)
// 	Close() error
// 	LocalAddr()
// 	RemoteAddr() Addr
// 	SetDeadline(t time.Time) error
// 	SetReadDeadline(t time.Time) error
// 	SetWriteDeadline(t time.Time) error
// }

// https://golang.org/pkg/bytes/
// https://www.kancloud.cn/digest/batu-go/153538
// https://juejin.im/post/5bf909cb51882521c8114523
func (pc *PrefixedConn) Read(b []byte) (n int, err error) {

	// Essentially need a buffer or something to do that. The root connection - you read from that. Need event loop
	// that reads the data from root connection and look at the first byte of the data. If it's 0, you push it to the buffer to where it's suppose to go.
	// you read to where it is suppose to go.
	// One: Need to inherit the original connection 2: some sort of buffer... have the structure (rpcDuplex) to push
	// into these channels of these branchconn/prefixconn

	// A -> B (0-prefixed) : Talk to B's RPC server.
	// A -> B (1-prefixed) : Talk to B's RPC client.
	// A <- B (0-prefixed) : Talk to A's RPC client.
	// A <- B (1-prefixed) : Talk to A's RPC server.

	// https://www.cnblogs.com/golove/p/3276678.html

	// Reads one byte of data from original conn to determine it's prefix
	buf := bytes.NewBuffer(b)
	prefixFromRoot, err := buf.ReadByte()

	if err != nil {
		log.Fatal("Error reading byte", err)
	}

	// Determine which channel to send to.
	if prefixFromRoot == 0 {
		pc.readBuf.ReadString

	} else if prefixFromRoot == 1 {
		pc.prefix = 0
	}

	if n == 0 {
		return n, io.EOF
	}

	return n, err
}

// Write is
func (pc *PrefixedConn) Write(b []byte) (n int, err error) {

	n, err = pc.writeConn.Write(append([]byte{pc.prefix}, b...))

	if n > 0 {
		n--
	}

	return n, err
}

// Close closes the connection.
func (pc *PrefixedConn) Close() error {
	return nil
}

// type Addr interface {
// 	Network() string // name of the network (for example, "tcp", "udp")
// 	String() string  // string form of address (for example, "192.0.2.1:25", "[2001:db8::1]:80")
// }

// LocalAddr returns the local network address.
func (pc *PrefixedConn) LocalAddr() net.Addr {
	var addr net.Addr
	return addr
}

// RemoteAddr returns the remote network address.
func (pc *PrefixedConn) RemoteAddr() net.Addr {
	var addr net.Addr
	return addr
}

// SetDeadline sets the read
func (pc *PrefixedConn) SetDeadline(t time.Time) error {
	return nil
}

// SetReadDeadline sets the deadline
func (pc *PrefixedConn) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls
func (pc *PrefixedConn) SetWriteDeadline(t time.Time) error {
	return nil
}

// prefixConn.Write() - I prefix it. I get that new data and write it back to original data. Prefix conn will inherit original connection.

// 1) PrefixConn implement net.Conn interface
// 2) Actual RPC server/ rpc client or rootConn

func main() {

	connA, connB := net.Pipe()
	defer connA.Close()
	defer connB.Close()

	api := new(API)

	svr := rpc.NewServer()

	svr.Register(api)

	go svr.ServeConn(connA)

	// CLIENT
	var reply Person
	a := Person{"Anto"}

	client := rpc.NewClient(connA)

	// Call RPC method through server
	err := client.Call("API.SayHello", a, &reply)

	if err != nil {
		log.Fatal("error", err)
	}

	fmt.Println(reply.Name)

	// https: //nathanleclaire.com/blog/2014/07/19/demystifying-golangs-io-dot-reader-and-io-dot-writer-interfaces/

	// ClientConn/ServerConn are exposed to rpc.Client and rpc.Server respectively.
	// When rpc.<x> calls Write() on their associated BranchConn, the BranchConn should prefix 0 or 1 to
	// the data before forwarding it onto the RootConn.

	// aDuplex is the initiator --> ROOTConn
	// aDuplex := NewRPCDuplex(connA, true)

	// aDuplex.clientConn.Write([]byte("haha"))

	// b, err := ioutil.ReadAll(connB)
	// if err != nil {
	// 	log.Fatal("error", err)
	// }

	// fmt.Println(string(b))

	// TASK IS TO DO FOLLOWING: Get A to talk to B's RPC server and A to talk to B's RPC client
	// A -> B (0-prefixed) : A Talk to B's RPC server.
	// A -> B (1-prefixed) : A Talk to B's RPC client.

	// d.clientConn = &PrefixedConn{prefix: 0}
	// d.serverConn = &PrefixedConn{prefix: 1}

	// Direct A's messages to B's RPC server

	// if aDuplex.clientConn.prefix == 0 {

	// 	go func() {
	// 		connA.Write([]byte("haha"))
	// 		connA.Close()
	// 	}()

	// 	b, err := ioutil.ReadAll(connB)
	// 	if err != nil {
	// 		log.Fatal("error", err)
	// 	}

	// 	fmt.Println(string(b))

	// }

	// ==============================================
	// NOOOOTES
	// ==============================================

	// A ---> B (A is initiator, B is responder)
	// Prefix read/writes with either 0 (talk to RPC server) or 1 (talk to RPC client).

	// A -> B (0-prefixed) : Talk to B's RPC server.
	// A -> B (1-prefixed) : Talk to B's RPC client.
	// A <- B (0-prefixed) : Talk to A's RPC client.
	// A <- B (1-prefixed) : Talk to A's RPC server.

	// 0 prefix (net.Conn), 1 prefix (net.Conn)
}

// What net.Pipe() returns
// func Pipe() (Conn, Conn) {
//     r1, w1 := io.Pipe()
//     r2, w2 := io.Pipe()

//     return &pipe{r1, w2}, &pipe{r2, w1}
// }

// type pipe struct {
//     *io.PipeReader
//     *io.PipeWriter
// }

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

// I have a very simple understanding of it. I understand the following... e.g

// type Shape interface {
// 	Area() float64
// 	Perimeter() float64
// }

// func (r Rect) Area() float64 {
// 	return r.width * r.height
// }

// type Rect struct {
// 	width  float64
// 	height float64
// }

// func main() {
// 	var s Shape
// 	s = Rect{5.0, 4.0}
// 	fmt.Println("area of rectange s", s.Area())
// }

// WRITE
// I would be able to declare var netConn and then assign prefixConn{prefix, writeConn, readBuff}
// into this netConn in main...

// A ---> B (A is initiator, B is responder)
// Prefix read/writes with either 0 (talk to RPC server) or 1 (talk to RPC client).

// A -> B (0-prefixed) : Talk to B's RPC server.
// A -> B (1-prefixed) : Talk to B's RPC client.
// A <- B (0-prefixed) : Talk to A's RPC client.
// A <- B (1-prefixed) : Talk to A's RPC server.

// 0 prefix (net.Conn), 1 prefix (net.Conn)

// First look at the data, prefix it then get that new data and write it further to the original net.Conn.

// return pc.writeConn.Write(append([]byte{pc.prefix}, b...))
