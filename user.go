package main

import (
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
	this.server.BroadCast(this, "inactive")
	this.server.mapLock.Unlock()
}

func (this *User) SendMsgToCurrentUser(msg string) {
	// this.conn.Write([]byte(msg + "\n"))
	this.sendMsgToUser(this, msg)
}

func (u *User) ModifyNickname(newName string) {

}

func (this *User) doMsg(msg string) {
	if msg == "who" {
		// 查询当前在线用户都有哪些
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := format("[%s] %s online...", user.Addr, user.Name)
			this.SendMsgToCurrentUser(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		newName := msg[7:]
		_, ok := this.server.OnlineMap[newName]
		if ok {
			this.SendMsgToCurrentUser("This name has been used")
			return
		}
		this.server.mapLock.Lock()
		delete(this.server.OnlineMap, this.Name)
		this.Name = newName
		this.server.OnlineMap[newName] = this
		this.server.mapLock.Unlock()
		this.SendMsgToCurrentUser(format("username updated success: %s", this.Name))
	} else {
		// 广播消息
		this.server.BroadCast(this, msg)
	}
}

func (this *User) sendMsgToUser(user *User, msg string) {
	user.conn.Write([]byte(msg + "\n"))
}

func (this *User) ListenMsg() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
