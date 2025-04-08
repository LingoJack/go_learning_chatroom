package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip        string
	Port      int
	OnlineMap map[string]*User
	mapLock   sync.RWMutex
	Message   chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) Handler(conn net.Conn) {
	fmt.Println("connect successfully")
	user := NewUser(conn, this)
	user.Online()
	aliveChannel := make(chan bool)
	go func() {
		buffer := make([]byte, 4096)
		len, err := conn.Read(buffer)
		if len == 0 {
			user.OffLine()
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Connection Read error: ", err)
			return
		}
		msg := string(buffer[:len-1])
		user.doMsg(msg)
		aliveChannel <- true
	}()
	for {
		select {
		case <-aliveChannel:

		case <-time.After(time.Second * 10):
			// op1 这个是发送到chan了，后续才来消费，现在希望这个消息发出去之后才执行op2
			user.sendMsgToUser(user, "too long no operation... you have been offline")
			user.OffLine()
			// op2
			close(user.C)
			conn.Close()
		}
	}
}

func (this *Server) BroadCast(user *User, msg string) {
	// sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg + "\n"
	sendMsg := format("[%s] %s: %s\n", user.Addr, user.Name, msg)
	this.Message <- sendMsg
}

func (this *Server) ListenMsg() {
	for {
		msg := <-this.Message
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}

func (this *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s: %d", this.Ip, this.Port))
	if exception(err, "net lisen erro") {
		return
	}
	defer listener.Close()
	go this.ListenMsg()
	for {
		connect, err := listener.Accept()
		if exception(err, "listen accept error") {
			continue
		}
		go this.Handler(connect)
	}
}
