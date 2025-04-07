package main

import (
	"fmt"
	"io"
	"net"
	"sync"
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
		user.SendMsg(msg)
	}()
	select {}
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ": " + msg
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
	if err != nil {
		fmt.Println("net listen error: ", err)
		return
	}
	defer listener.Close()
	go this.ListenMsg()
	for {
		connect, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept error: ", err)
			continue
		}
		go this.Handler(connect)
	}
}
