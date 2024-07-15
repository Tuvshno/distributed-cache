package main

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestMainFunction(t *testing.T) {
	conn, err := net.Dial("tcp", ":3000")
	if err != nil {
		t.Fatalf("failed connection: %s", err)
	}
	defer conn.Close()

	got, err := conn.Write([]byte("SET Foo Bar 4000000000"))
	if err != nil {
		t.Fatalf("failed write: %s", err)
	}

	fmt.Println("Got : ", got)

	time.Sleep(time.Second * 1)

	got, err = conn.Write([]byte("GET Foo"))
	if err != nil {
		t.Fatalf("failed write: %s", err)
	}

	fmt.Println("Got : ", got)

}
