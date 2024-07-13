package main

import (
	"distributed-cache/cache"
	"log"
	"net"
	"time"
)

func main() {
	opts := ServerOpts{
		ListenAddr: ":3000",
		IsLeader:   true,
	}

	go func() {
		time.Sleep(time.Second * 2)
		conn, err := net.Dial("tcp", ":3000")
		if err != nil {
			log.Fatal(err)
		}

		conn.Write([]byte("SET Foo Bar 3400"))
	}()

	server := NewServer(opts, cache.New())
	server.Start()
}
