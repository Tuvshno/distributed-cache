package main

import (
	"fmt"
	"time"
)

type Command string

const (
	CMDSet    Command = "SET"
	CMDGet    Command = "GET"
	CMDDelete Command = "DELETE"
)

type Message struct {
	Cmd   Command
	Key   []byte
	Value []byte
	TTL   time.Duration
}

func (m *Message) ToBytes() []byte {
	var cmd string

	switch m.Cmd {
	case CMDSet:
		cmd = fmt.Sprintf("%s %s %s %d", m.Cmd, m.Key, m.Value, m.TTL)
	case CMDGet:
		cmd = fmt.Sprintf("%s %s", m.Cmd, m.Key)
	case CMDDelete:
		cmd = fmt.Sprintf("%s %s", m.Cmd, m.Key)
	default:
		panic("unknown command")
	}

	return []byte(cmd)

}
