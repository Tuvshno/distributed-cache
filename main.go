package main

import (
	"distributed-cache/cache"
	"flag"
)

func main() {
	listenaddr := flag.String("listenaddr", ":3000", "listen address of server")
	leaderaddr := flag.String("leaderaddr", "", "the listen address of the leader")
	flag.Parse()

	opts := ServerOpts{
		ListenAddr: *listenaddr,
		IsLeader:   len(*leaderaddr) == 0,
		LeaderAddr: *leaderaddr,
	}

	server := NewServer(opts, cache.New())
	server.Start()
}
