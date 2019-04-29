package main

import (
	"bytes"
	"io"
	"log"
	"net"
	"time"
)

// RPCDuplex holds the basic structure of two prefixed connections
type RPCDuplex struct {
	clientConn *PrefixedConn
	serverConn *PrefixedConn
}

// NewRPCDuplex initiates a new RPCDuplex struct and reads in the
func NewRPCDuplex(conn net.Conn, initiator bool) *RPCDuplex {
	var d RPCDuplex
	var buf bytes.Buffer

	// PrefixedConn implements net.Conn and assigned it to d.clientConn and d.serverConn
	if initiator {
		d.clientConn = &PrefixedConn{prefix: 0, writeConn: conn, readBuf: buf}
		d.serverConn = &PrefixedConn{prefix: 1, writeConn: conn, readBuf: buf}
	} else {
		d.clientConn = &PrefixedConn{prefix: 1, writeConn: conn, readBuf: buf}
		d.serverConn = &PrefixedConn{prefix: 0, writeConn: conn, readBuf: buf}
	}

	return &d
}

// PrefixedConn will inherit net.Conn from the interface.
type PrefixedConn struct {
	prefix    byte
	writeConn io.Writer    // Original connection. net.Conn has a write method therefore it implements the writer interface.
	readBuf   bytes.Buffer // Read data from original connection. It is RPCDuplex's responsibility to push to here.
}

// Read reads in prefixed data from root connection and reads it into the appropriate branch connection
func (pc *PrefixedConn) Read(b []byte) (n int, err error) {

	n, err = pc.readBuf.Write(b)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("Write %d bytes to prfxConn, content is: %s\n", n, pc.readBuf.String())
	log.Printf("Read %d bytes from inside, content is: %s\n", n, string(b))

	return n, err

}

// Write prefixes data to the connection and then writes this prefixed data to the root connection.
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

func main() {

	var buf = make([]byte, 1024)

	svr, client := net.Pipe()
	defer svr.Close()
	defer client.Close()

	aDuplex := NewRPCDuplex(client, true)
	bDuplex := NewRPCDuplex(svr, true)

	b := []byte("Helloworld")

	go func() {

		// n, err := client.Write([]byte(b))
		n, err := aDuplex.clientConn.Write([]byte(b))
		if err != nil {
			log.Fatalln("error writing to server", err)
		}

		log.Printf("Write %d bytes to server, content is: %s\n", n, string(b))

	}()

	time.Sleep(time.Second * 1)

	// Reads n bytes from client
	n, err := svr.Read(buf[:])
	if err != nil {
		log.Fatalln("error reading from conn", err)
	}

	n, err = bDuplex.serverConn.Read(buf[:n])
	if err != nil {
		log.Fatalln("error reading from conn", err)
	}
	log.Printf("Read %d bytes from bDuplex, content is: %s\n", n, string(buf[:n]))

}
