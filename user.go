package main

import (
	"fmt"
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	go user.ListenMsg()
	return user
}

func (this *User) Online() {
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "active")
}

func (this *User) OffLine() {
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()
	this.server.BroadCast(this, "inactive")
}

func (this *User) SendMsgToCurrentUser(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) SendMsg(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := fmt.Sprintf("[%s] %s online...\n", user.Addr, user.Name)
			this.SendMsgToCurrentUser(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else {
		// 广播消息
		this.server.BroadCast(this, msg)
	}
}

func (this *User) ListenMsg() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
