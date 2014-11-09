package irc

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"net"
	"sync"
	"regexp"
	"time"
)

type World struct {
	sync.RWMutex
	HostName        string
	Rooms           map[string]*Room
	Users           map[string]*User
	EventDispatcher *EventDispatcher
}

func (self *World) NewRoom(name []byte) (*Room, error) {
	self.Lock()
	defer self.Unlock()
	if _, ok := self.Rooms[string(name)]; !ok {
		room := NewRoom(name)
		self.Rooms[string(name)] = room

		return room, nil
	}

	return nil, fmt.Errorf("not found")
}

func (self *World) NewRoomWithCallback(name []byte, callback func(*Room)) (*Room, error) {
	self.Lock()
	defer self.Unlock()
	if _, ok := self.Rooms[string(name)]; !ok {
		room := NewRoomWithCallback(name, callback)
		self.Rooms[string(name)] = room
		return room, nil
	}

	return nil, fmt.Errorf("not found")
}

func NewWorld() *World {
	world := &World{
		HostName:        "example.net",
		Rooms:           make(map[string]*Room),
		Users:           make(map[string]*User),
		EventDispatcher: NewEventDispatcher(),
	}

	// ping
	world.EventDispatcher.Subscribe("kernel.ping", 500, func(args interface{}) bool {
		if ev, ok := args.(*PingEvent); ok {
			ev.User.Write([]byte("PONG example.net\n"))
		}
		return true
	})

	// part
	world.EventDispatcher.Subscribe("kernel.part", 500, func(args interface{}) bool {
		if ev, ok := args.(*PartEvent); ok {
			if room, err := world.GetRoom(ev.Room); err == nil {
				for _, user := range room.Users {
					user.Write([]byte(fmt.Sprintf(":%s PART %s :%s\n", ev.User.Name, room.Name, ev.User.Name)))
				}
				room.PartUser(ev.User)

				if len(room.Users) == 0 {
					world.RemoveRoom(ev.Room)
				}
			}
		}
		return true
	})

	// privmsg (auto join for twitter)
	world.EventDispatcher.Subscribe("kernel.privmsg", 500, func(args interface{}) bool {
		if ev, ok := args.(*NewMessageEvent); ok {
			if room, err := world.GetRoom(ev.Room); err == nil {
				if _, ok := room.Users[string(ev.User.Name)]; !ok {
					world.EventDispatcher.Dispatch("kernel.join", &JoinEvent{
						Rooms: [][]byte{ev.Room},
						User:  ev.User,
					})
				}
			}
		}
		return true
	})

	// privmsg (debug)
	world.EventDispatcher.Subscribe("kernel.privmsg", 900, func(args interface{}) bool {
		if ev, ok := args.(*NewMessageEvent); ok {
			fmt.Printf("Params: %s\n", ev.Params)
		}
		return true
	})

	// privmsg (unescape)
	world.EventDispatcher.Subscribe("kernel.privmsg", 600, func(args interface{}) bool {
		if ev, ok := args.(*NewMessageEvent); ok {
			message := html.UnescapeString(string(ev.Message))
			ev.Message = make([]byte, len(message))
			copy(ev.Message, message)
		}
		return true
	})

	// privmsg (irc default)
	world.EventDispatcher.Subscribe("kernel.privmsg", 100, func(args interface{}) bool {
		if ev, ok := args.(*NewMessageEvent); ok {
			if room, err := world.GetRoom(ev.Room); err == nil {
				room.RLock()
				for _, target := range room.Users {
					if target == ev.User {
						continue
					}

					messages := regexp.MustCompile("\r?\n").Split(string(ev.Message), -1)
					for _, message := range messages {
						target.Write([]byte(fmt.Sprintf(":%s PRIVMSG %s :%s\n", ev.User.Name, room.Name, message)))
					}
				}
				room.RUnlock()
			} else {
				if room, err := world.NewRoom(ev.Room); err == nil {
					room.AddUser(ev.User)
				}
			}
		}
		return true
	})


	// notice
	world.EventDispatcher.Subscribe("kernel.notice", 500, func(args interface{}) bool {
		if ev, ok := args.(*NoticeEvent); ok {
			if room, err := world.GetRoom(ev.Room); err == nil {
				room.RLock()
				for _, target := range room.Users {
					messages := regexp.MustCompile("\r?\n").Split(string(ev.Message), -1)
					for _, message := range messages {
						target.Write([]byte(fmt.Sprintf(":%s NOTICE %s :%s\n", ev.User.Name, room.Name, message)))
					}
				}
				room.RUnlock()
			} else {
				if room, err := world.NewRoom(ev.Room); err == nil {
					room.AddUser(ev.User)
				}
			}
		}
		return true
	})

	// topic
	world.EventDispatcher.Subscribe("kernel.topic", 500, func(args interface{}) bool {
		if ev, ok := args.(*TopicEvent); ok {
			if room, err := world.GetRoom(ev.Room); err == nil {
				room.RLock()
				for _, target := range room.Users {
					target.Write([]byte(fmt.Sprintf(":%s TOPIC %s :%s\n", ev.User.Name, ev.Room, ev.Message)))
				}
				room.RUnlock()
			}
		}
		return true
	})

	// join
	world.EventDispatcher.Subscribe("kernel.join", 500, func(args interface{}) bool {
		if ev, ok := args.(*JoinEvent); ok {
			for _, room_name := range ev.Rooms {
				var room *Room
				var err error

				if room, err = world.GetRoom(room_name); err == nil {
					room.AddUser(ev.User)
				} else {
					world.EventDispatcher.Dispatch("kernel.room.create", &NewRoomEvent{
						Room: room_name,
						User: ev.User,
					})
				}

				if room, err = world.GetRoom(room_name); err != nil {
					panic(err)
				}

				var users [][]byte
				for _, user := range room.Users {
					users = append(users, user.Name)
				}

				for _, u := range room.Users {
					u.Write([]byte(fmt.Sprintf(":%s JOIN :%s\n", ev.User.Name, room_name)))

					u.Write([]byte(fmt.Sprintf(":example.net 353 %s = %s :%s\n", ev.User.Name, room_name, bytes.Join(users, []byte(" ")))))
					u.Write([]byte(fmt.Sprintf(":example.net 366 %s %s :End of /NAMES list\n", ev.User.Name, room_name)))
				}
			}
		}
		return true
	})

	world.EventDispatcher.Subscribe("kernel.room.create", 500, func(args interface{}) bool {
		if ev, ok := args.(*NewRoomEvent); ok {
			var room *Room
			if ev.Callback != nil {
				room, _ = world.NewRoomWithCallback(ev.Room, ev.Callback)
			} else {
				room, _ = world.NewRoom(ev.Room)
			}

			if room != nil && ev.User != nil {
				room.AddUser(ev.User)
			}
		}
		return true
	})

	world.EventDispatcher.Subscribe("kernel.room.invite", 500, func(args interface{}) bool {
		if ev, ok := args.(*InviteEvent); ok {
			if v, ok := world.Users[string(ev.To)]; ok {
				v.Write([]byte(fmt.Sprintf(":%s INVITE %s %s\n", ev.From, ev.To, ev.Room)))
			}
		}
		return true
	})

	world.EventDispatcher.Subscribe("kernel.time", 500, func(args interface{}) bool {
		if ev, ok := args.(*TimeEvent); ok {
			if string(ev.Server) == "irc.example.net" {
				now := time.Now()
				ev.User.Write([]byte(fmt.Sprintf(":irc.example.net 391 %s :%s\n", ev.User.Nick, now.Format(time.ANSIC))))
			}
		}
		return true
	})

	world.EventDispatcher.Subscribe("kernel.who", 500, func(args interface{}) bool {
		if ev, ok := args.(*WhoEvent); ok {
			if len(ev.Query) < 1 {
				return false
			}

			if ev.Query[0] == 0x23 { // #はじまり
				// list room members
				if room, err := world.GetRoom(ev.Query); err == nil {
					for _, u := range room.Users {
						ev.User.Write([]byte(fmt.Sprintf(":irc.example.net 352 %s %s %s\n",
							u.Name, room.Name, u.Name,
						)))
					}
					ev.User.Write([]byte(fmt.Sprintf(":irc.example.net 315 %s :End of /Who list\n",
						room.Name,
					)))
				} else {
					ev.User.Write([]byte(fmt.Sprintf(":irc.example.net 315 %s :End of /Who list\n",
						room.Name,
					)))
				}

			}
		}
		return true
	})



	return world
}

func (world *World) SendPrivMessage(from, room, message string) {
	var user *User
	var ok bool

	if _, ok = world.Users[from]; !ok {
		world.Users[from] = &User{
			Name: []byte(from),
		}
		user = world.Users[from]
	} else {
		user = world.Users[from]
	}

	if r, ok := world.Rooms[room]; ok {
		if _, exist := r.Users[from]; !exist {
			world.EventDispatcher.Dispatch("kernel.join", &JoinEvent{
				Rooms: [][]byte{[]byte(room)},
				User:  user,
			})
		}
	}

	world.EventDispatcher.Dispatch("kernel.privmsg", &NewMessageEvent{
		Room:     []byte(room),
		User:     user,
		Message:  []byte(message),
		External: true,
	})
}

func (self *World) GetRoom(name []byte) (*Room, error) {
	self.RLock()
	defer self.RUnlock()
	if room, ok := self.Rooms[string(name)]; ok {
		return room, nil
	}

	return nil, fmt.Errorf("not found")
}

func (self *World) RemoveRoom(name []byte) (*Room, error) {
	self.Lock()
	defer self.Unlock()
	if v, ok := self.Rooms[string(name)]; ok {
		v.CloseRoom()
	}
	delete(self.Rooms, string(name))
	return nil, nil
}


func (world *World) HandleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	parser := NewParser(reader)

	var user *User

	conn.Write([]byte(fmt.Sprintf(":example.net NOTICE Auth :*** Looking up your hostname\n")))
	conn.Write([]byte(fmt.Sprintf(":example.net NOTICE Auth :*** Found your hostname\n")))

	for {
		msg, err := parser.Parse()
		if err != nil {
			return
		}

		// ここらへんはevent dispatchだけにしておく
		switch msg.GetCommandType() {
		case COMMAND_CAP:
			p := string(msg.GetParameter(0))
			if p == "LS" {
				conn.Write([]byte(":example.net CAP * LS :multi-prefix\n"))
			} else if p == "RES" {
				conn.Write([]byte(":example.net CAP * ACK :multi-prefix\n"))
			} else {
				//conn.Write([]byte(":example.net CAP * ACK :multi-prefix\n"))
			}
		case COMMAND_NICK:
			user = NewUser(msg.GetParameter(0), conn)
			world.Users[string(user.Name)] = user
		case COMMAND_USER:
			conn.Write([]byte(fmt.Sprintf(":example.net 001 %s :Welcome to the Internet relay network!\n", user.Name)))
		case COMMAND_PING:
			world.EventDispatcher.Dispatch("kernel.ping", &PingEvent{
				User: user,
			})
		case COMMAND_QUIT:
			world.EventDispatcher.Dispatch("kernel.quit", &QuitEvent{
				User: user,
			})
		case COMMAND_PRIVMSG:
			room := msg.GetParameter(0)
			message := msg.GetParameter(1)[1:]
			ev := &NewMessageEvent{
				User:    user,
				Room:    make([]byte, len(room)),
				Message: make([]byte, len(message)),
				Params:  msg.GetParameters(),
			}
			copy(ev.Room, room)
			copy(ev.Message, message)

			world.EventDispatcher.Dispatch("kernel.privmsg", ev)
		case COMMAND_TOPIC:
			world.EventDispatcher.Dispatch("kernel.topic", &TopicEvent{
				Room:    msg.GetParameter(0),
				User:    user,
				Message: msg.GetParameter(1)[1:],
			})
		case COMMAND_PART:
			//部屋から抜いてspread
			world.EventDispatcher.Dispatch("kernel.part", &PartEvent{
				Room:    msg.GetParameter(0),
				User:    user,
				Message: msg.GetParameter(1)[1:],
			})
		case COMMAND_MODE:
			user.Write([]byte(fmt.Sprintf(":example.net 324 %s %s\n", user.Name, "+")))
		case COMMAND_JOIN:
			if user == nil {
				conn.Write([]byte(":example.net 451 * :Connection not registered\n"))
				break
			}

			params := msg.GetParameters()
			rooms := bytes.Split(params[0], []byte(","))
			fmt.Printf("ROOM: %s, %s\n", rooms, params[0])
			if len(rooms) == 0 {
				rooms = [][]byte{params[0]}
			}

			ev := JoinEvent{
				Rooms: rooms,
				User:  user,
			}
			world.EventDispatcher.Dispatch("kernel.join", &ev)
		case COMMAND_ISON:
		case COMMAND_WHO:
			world.EventDispatcher.Dispatch("kernel.who", &WhoEvent{
					User: user,
					Query: msg.GetParameter(1),
			})
		case COMMAND_PONG:
			world.EventDispatcher.Dispatch("kernel.pong", &PongEvent{
				User: user,
			})
		case COMMAND_PASS:
			world.EventDispatcher.Dispatch("kernel.password", &PasswordEvent{
				User: user,
				Password: msg.GetParameter(0),
			})
		case COMMAND_SERVER:
		case COMMAND_OPER:
		case COMMAND_NAMES:
		case COMMAND_LIST:
		case COMMAND_INVITE:
			world.EventDispatcher.Dispatch("kernel.invite", &InviteEvent{
					Room: msg.GetParameter(0),
					To: msg.GetParameter(1),
					From: user.Name,
			})
		case COMMAND_KICK:
		case COMMAND_VERSION:
		case COMMAND_STATS:
		case COMMAND_LINKS:
		case COMMAND_TIME:
			world.EventDispatcher.Dispatch("kernel.time", &TimeEvent{
				Server: msg.GetParameter(0),
				User: user,
			})
		case COMMAND_CONNECT:
		case COMMAND_TRACE:
		case COMMAND_ADMIN:
		case COMMAND_INFO:
		case COMMAND_NOTICE:
			world.EventDispatcher.Dispatch("kernel.notice", &NoticeEvent{
					Room: msg.GetParameter(0),
					User: user,
					Message: msg.GetParameter(1)[1:],
			})
		case COMMAND_WHOWAS:
		case COMMAND_KILL:
		case COMMAND_ERROR:
		case COMMAND_OPTIONALS:
		case COMMAND_AWAY:
		}
	}
}
