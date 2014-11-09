package irc

type NewRoomEvent struct {
	Room []byte
	User *User
	Callback func(*Room)
}

type JoinEvent struct {
	Rooms [][]byte
	User *User
}

type InviteEvent struct {
	Room []byte
	To []byte
	From []byte
}

type PartEvent struct {
	Room []byte
	User *User
	Message []byte
}

type WhoEvent struct {
	User *User
	Query []byte
}

type PasswordEvent struct {
	User *User
	Password []byte
}

type PingEvent struct {
	User *User
}

type PongEvent struct {
	User *User
}

type TimeEvent struct {
	Server []byte
	User *User
}

type NewMessageEvent struct {
	Room []byte
	User *User
	Message []byte
	Params [][]byte
	External bool
}

type NoticeEvent struct {
	Room []byte
	User *User
	Message []byte
}

type TopicEvent struct {
	Room []byte
	User *User
	Message []byte
}

type QuitEvent struct {
	User *User
}
