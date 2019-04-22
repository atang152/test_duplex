package main

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestACallsB(t *testing.T) {

	var t1 = struct {
		testInput      Person
		response       Person
		expectedResult string
	}{
		testInput:      Person{"Anto"},
		expectedResult: "Anto",
	}

	connA, connB := net.Pipe()
	defer connA.Close()
	defer connB.Close()

	c1 := make(chan *RPCDuplex)

	api := new(API)

	// Run a NewRPCDuplex as a server (serving on connA) in go routine
	go func() {
		aDuplex := NewRPCDuplex(connA)
		aDuplex.Register(api)
		aDuplex.Serve()
	}()

	// Run a seperate NewRPCDuplex as a client (serving on connB) in go routine
	go func() {
		bDuplex := NewRPCDuplex(connB)
		c1 <- bDuplex
	}()

	client := <-c1
	err := client.Call("API.SayHello", t1.testInput, &t1.response)
	assert.Nil(t, err)
	assert.Equal(t, t1.response.Name, t1.expectedResult, "The two should be the same.")

}
