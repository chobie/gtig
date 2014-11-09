package irc

import (
	"time"
	"net"
	"fmt"
)

type User struct {
	Nick []byte
	Name []byte
	Host string
	Server []byte
	RealName []byte
	Last *time.Time
	conn net.Conn
}

func (self *User) GetFQ() string {
	return fmt.Sprintf("%s!~%s@%s", self.Nick, self.Nick, self.Server)
}

func (self *User) Write(data []byte) {
	if self.conn != nil {
		fmt.Printf("> %s", data)
		self.conn.Write(data)
	}
}

func NewUser(name []byte, conn net.Conn) *User {
	u := &User{
		Name: make([]byte, len(name)),
		conn: conn,
	}
	copy(u.Name, name)

	return u
}
