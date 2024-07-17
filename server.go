package main

import (
	"context"
	"distributed-cache/cache"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"time"
)

type ServerOpts struct {
	ListenAddr string
	IsLeader   bool
	LeaderAddr string
}

type Server struct {
	ServerOpts

	followers map[net.Conn]struct{}
	clients   map[net.Conn]struct{}
	cache     cache.Cacher
}

func NewServer(opts ServerOpts, c cache.Cacher) *Server {
	return &Server{
		ServerOpts: opts,
		cache:      c,
		followers:  make(map[net.Conn]struct{}),
		clients:    make(map[net.Conn]struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return fmt.Errorf("listen error: %s", err)
	}

	log.Printf("server starting on port %s \n", s.ListenAddr)

	if !s.IsLeader {
		go func() {
			conn, err := net.Dial("tcp", s.LeaderAddr)
			fmt.Println("Connected to leader : ", s.LeaderAddr)
			if err != nil {
				log.Fatal(err)
			}
			conn.Write([]byte("FOLLOWER\n"))
			s.handleConn(conn)
		}()
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept error  : %s \n", err)
			continue
		}

		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.removeConn(conn)
	}()

	if s.IsLeader {
		bufMsg := make([]byte, 2048)

		n, err := conn.Read(bufMsg)
		if err != nil {
			log.Printf("Handshake read error: %v", err)
			return
		}
		handshakeMsg := string(bufMsg[:n])

		if strings.TrimSpace(handshakeMsg) == "FOLLOWER" {
			s.addFollower(conn)
			log.Println("Follower connected:", conn.RemoteAddr())
		} else {
			s.addClient(conn)
			log.Println("Client connected:", conn.RemoteAddr())
		}
	}

	fmt.Println("Connection made : ", conn.RemoteAddr())

	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Conn read error: %s \n", err)
			break
		}

		msg := buf[:n]
		log.Println(string(msg))

		go s.handleCommand(conn, msg)
	}
}

func (s *Server) handleCommand(conn net.Conn, rawCmd []byte) {
	msg, err := parseCommand(rawCmd)
	if err != nil {
		log.Println("Failed to parse command : ", err)
		conn.Write([]byte(err.Error()))
		return
	}

	log.Printf("Recieved Command %s", msg.Cmd)
	switch msg.Cmd {
	case CMDSet:
		err = s.handleSetCommand(conn, msg)
	case CMDGet:
		err = s.handleGetCommand(conn, msg)
	case CMDDelete:
		err = s.handleDeleteCommand(conn, msg)
	}

	if err != nil {
		log.Println("Failed to handle command : ", err)
		conn.Write([]byte(err.Error()))
	}

}

func (s *Server) handleSetCommand(conn net.Conn, msg *Message) error {
	if err := s.cache.Set(msg.Key, msg.Value, msg.TTL); err != nil {
		conn.Write([]byte(fmt.Sprintf("Error setting key: %v", err)))
		return err
	}

	go s.sendToFollowers(context.TODO(), msg)

	return nil
}

func (s *Server) handleGetCommand(conn net.Conn, msg *Message) error {
	val, err := s.cache.Get(msg.Key)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error getting key: %v", err)))
		return err
	}
	log.Println("Got Value: ", string(val))
	_, err = conn.Write(val)
	if err != nil {
		log.Printf("Error writing value to client: %v", err)
		return err
	}
	return nil
}

func (s *Server) handleDeleteCommand(conn net.Conn, msg *Message) error {
	err := s.cache.Delete(msg.Key)
	if err != nil {
		conn.Write([]byte(fmt.Sprintf("Error deleting key: %v", err)))
		return err
	}

	_, err = conn.Write([]byte(fmt.Sprintf("Deleted %s", msg.Key)))
	go s.sendToFollowers(context.TODO(), msg)

	return nil
}

func (s *Server) sendToFollowers(ctx context.Context, msg *Message) error {
	for conn := range s.followers {
		rawMsg := msg.ToBytes()
		fmt.Println(" fowarding :", string(rawMsg))
		_, err := conn.Write(rawMsg)
		if err != nil {
			log.Println("write to followers error ", err)
			continue
		}
	}

	return nil
}

func parseCommand(raw []byte) (*Message, error) {
	var (
		rawStr = strings.TrimSpace(string(raw))
		parts  = strings.Split(rawStr, " ")
	)

	if len(parts) < 2 {
		log.Print("invalid command")
		return nil, errors.New("invalid command")
	}

	msg := &Message{
		Cmd: Command(parts[0]),
		Key: []byte(parts[1]),
	}

	if msg.Cmd == CMDSet {
		if len(parts) != 4 {
			log.Print("invalid command")
			return nil, errors.New("invalid SET command")
		}
		msg.Value = []byte(parts[2])

		ttl, err := strconv.Atoi(strings.TrimSpace(parts[3]))

		if err != nil {
			log.Print("invalid SET ttl command")
			return nil, fmt.Errorf("invalid SET ttl command: %v", err)
		}
		msg.TTL = time.Duration(ttl)
	}

	return msg, nil
}

func (s *Server) addFollower(conn net.Conn) {
	s.followers[conn] = struct{}{}
}

func (s *Server) addClient(conn net.Conn) {
	s.clients[conn] = struct{}{}
}

func (s *Server) removeConn(conn net.Conn) {
	delete(s.followers, conn)
	delete(s.clients, conn)
}
