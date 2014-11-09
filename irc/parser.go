package irc

import (
	"bufio"
	"fmt"
	"bytes"
)

func NewParser(reader *bufio.Reader) *IrcProtocolParser {
	return &IrcProtocolParser{
		reader: reader,
	}
}


type IrcProtocolParser struct {
	reader *bufio.Reader
}

func (self *IrcProtocolParser) subParse(line []byte, expected_arg_count int, advanced *int) [][]byte {
	result := make([][]byte, 0)
	offset := 0
	for i := 0; i < len(line); i++ {
		if line[i] == 0x20 {
			result = append(result, line[offset:offset + i])

			i++
			offset = i
			if len(result) +1 >= expected_arg_count {
				result = append(result, line[offset:])
				offset = len(line)
				break
			}
		}
	}

	if offset < len(line) {
		result = append(result, line)
		offset = len(line)
	}

	*advanced = offset;
	return result
}

func (self *IrcProtocolParser) Parse() (*Message, error) {
	line, _, err := self.reader.ReadLine()
	if err != nil {
		return nil, err
	}
	// Textualだとなぜかゴミが入るんだよね
	line = bytes.TrimSuffix(line, []byte{0x01})
	fmt.Printf("< %s\n", line)

	msg := &Message{}
	state := 0
	offset := 0
	for i := 0; i < len(line); i++ {
		code := line[i]

		if state == 0 {
			if !((0x41 <= code && code <= 0x5a) || (0x61 <= code && code <= 0x7a)) {
				state = 1
				err := msg.SetCommand(line[0:i])
				if err != nil {
					return nil, err
				}
				offset = i+1
			}
		} else {
			// spaceは一個以上あるときもあるのでtrimする必要がある
			advanced := 0
			var parts [][]byte

			switch msg.GetCommandType() {
			// KICK, INVITE
			case COMMAND_USER, COMMAND_CAP, COMMAND_MODE, COMMAND_WHO:
				parts = self.subParse(line[offset:], 4, &advanced)
			case COMMAND_PRIVMSG:
				parts = self.subParse(line[offset:], 2, &advanced)
				if len(parts) > 0 {
					p := bytes.ToLower(parts[1])
					if bytes.Contains(p, []byte("action")) {
						_parts := bytes.Split([]byte(" "), parts[1])
						parts = append(parts, _parts[1:]...)
					}
				}
			case COMMAND_PING, COMMAND_PONG, COMMAND_ME, COMMAND_TOPIC ,COMMAND_PART, COMMAND_QUIT, COMMAND_ISON,
				COMMAND_PASS, COMMAND_SERVER, COMMAND_OPER, COMMAND_NAMES, COMMAND_LIST, COMMAND_INVITE, COMMAND_KICK,
				COMMAND_VERSION, COMMAND_STATS, COMMAND_LINKS, COMMAND_TIME, COMMAND_CONNECT, COMMAND_TRACE, COMMAND_ADMIN,
				COMMAND_INFO, COMMAND_NOTICE, COMMAND_WHOWAS, COMMAND_KILL, COMMAND_ERROR, COMMAND_OPTIONALS, COMMAND_AWAY,
				COMMAND_NICK, COMMAND_JOIN:
				// &か#始まり、, Ctrl+Gは許可されない
				parts = self.subParse(line[offset:], 2, &advanced)
			default:
				parts = self.subParse(line[offset:], 2, &advanced)
			}

			msg.SetParameters(parts);
			i = i + advanced
		}
	}

	return msg, nil
}

type ResponseType int

const (
	RPL_NONE ResponseType = 300
	RPL_USERHOST ResponseType = 302
	RPL_ISON ResponseType = 303
	RPL_AWAY ResponseType = 301
	RPL_UNAWAY ResponseType = 305
	RPL_NOWAWAY ResponseType = 306
	RPL_WHOISUSER ResponseType = 311
	RPL_WHOISSERVER ResponseType = 312
	RPL_WHOISOPERATOR ResponseType = 313
	RPL_WHOISIDLE ResponseType = 317
	RPL_ENDOFWHOIS ResponseType = 318
	RPL_WHOISCHANNELS ResponseType = 319
	RPL_WHOWASUSER ResponseType = 314
	RPL_ENDOFWHOWAS ResponseType = 369
	RPL_LISTSTART ResponseType = 321
	RPL_LIST ResponseType = 322
	RPL_LISTEND ResponseType = 323
	RPL_CHANNELMODEIS ResponseType = 324
	RPL_NOTOPIC ResponseType = 331
	RPL_TOPIC ResponseType = 332
	RPL_INVITING ResponseType = 341
	RPL_SUMMONING ResponseType = 342
	RPL_VERSION ResponseType = 351
	RPL_WHOREPLY ResponseType = 352
	RPL_ENDOFWHO ResponseType = 315
	RPL_NAMREPLY ResponseType = 353
	RPL_ENDOFNAMES ResponseType = 366
	RPL_LINKS ResponseType = 364
	RPL_ENDOFLINKS ResponseType = 365
	RPL_BANLIST ResponseType = 367
	RPL_ENDOFBANLIST ResponseType = 368
	RPL_INFO ResponseType = 371
	RPL_ENDOFINFO ResponseType = 374
	RPL_MOTDSTART ResponseType = 375
	RPL_MOTD ResponseType = 372
	RPL_ENDOFMOTD ResponseType = 376
	RPL_YOUREOPER ResponseType = 381
	RPL_REHASHING ResponseType = 382
	RPL_TIME ResponseType = 391
	RPL_USERSSTART ResponseType = 392
	RPL_USERS ResponseType = 393
	RPL_ENDOFUSERS ResponseType = 394
	RPL_NOUSERS ResponseType = 395
	RPL_TRACELINK ResponseType = 200
	RPL_TRACECONNECTING ResponseType = 201
	RPL_TRACEHANDSHAKE ResponseType = 202
	RPL_TRACEUNKNOWN ResponseType = 203
	RPL_TRACEOPERATOR ResponseType = 204
	RPL_TRACEUSER ResponseType = 205
	RPL_TRACESERVER ResponseType = 206
	RPL_TRACENEWTYPE ResponseType = 208
	RPL_TRACELOG ResponseType = 261
	RPL_STATSLINKINFO ResponseType = 211
	RPL_STATSCOMMANDS ResponseType = 212
	RPL_STATSCLINE ResponseType = 213
	RPL_STATSNLINE ResponseType = 214
	RPL_STATSILINE ResponseType = 215
	RPL_STATSKLINE ResponseType = 216
	RPL_STATSYLINE ResponseType = 218
	RPL_ENDOFSTATS ResponseType = 219
	RPL_STATSLLINE ResponseType = 241
	RPL_STATSUPTIME ResponseType = 242
	RPL_STATSOLINE ResponseType = 243
	RPL_STATSHLINE ResponseType = 244
	RPL_UMODEIS ResponseType = 221
	RPL_LUSERCLIENT ResponseType = 251
	RPL_LUSEROP ResponseType = 252
	RPL_LUSERUNKNOWN ResponseType = 253
	RPL_LUSERCHANNELS ResponseType = 254
	RPL_LUSERME ResponseType = 255
	RPL_ADMINME ResponseType = 256
	RPL_ADMINLOC1 ResponseType = 257
	RPL_ADMINLOC2 ResponseType = 258
	RPL_ADMINEMAIL ResponseType = 259
)

type MessageType int

const (
	COMMAND_UNKNOWN MessageType = iota
	COMMAND_NICK
	COMMAND_USER
	COMMAND_JOIN
	COMMAND_CAP
	COMMAND_MODE
	COMMAND_WHO
	COMMAND_PRIVMSG
	COMMAND_PING
	COMMAND_PONG
	COMMAND_ME
	COMMAND_TOPIC
	COMMAND_PART
	COMMAND_QUIT
	COMMAND_ISON
	COMMAND_PASS
	COMMAND_SERVER
	COMMAND_OPER
	COMMAND_NAMES
	COMMAND_LIST
	COMMAND_INVITE
	COMMAND_KICK
	COMMAND_VERSION
	COMMAND_STATS
	COMMAND_LINKS
	COMMAND_TIME
	COMMAND_CONNECT
	COMMAND_TRACE
	COMMAND_ADMIN
	COMMAND_INFO
	COMMAND_NOTICE
	COMMAND_WHOWAS
	COMMAND_KILL
	COMMAND_ERROR
	COMMAND_OPTIONALS
	COMMAND_AWAY
	COMMAND_REHASH
	COMMAND_RESTART
	COMMAND_SUMMON
	COMMAND_USERS
	COMMAND_OPERWALL
	COMMAND_USERHOST
)

type Message struct {
	command []byte
	commandType MessageType
	parameters [][]byte
}

func (self *Message) SetCommand(command []byte) error {
	self.command = command
	cmd := string(bytes.ToLower(command))

	switch cmd {
	case "nick":
		self.commandType = COMMAND_NICK
	case "user":
		self.commandType = COMMAND_USER
	case "join":
		self.commandType = COMMAND_JOIN
	case "cap":
		self.commandType = COMMAND_CAP
	case "mode":
		self.commandType = COMMAND_MODE
	case "who":
		self.commandType = COMMAND_WHO
	case "privmsg":
		self.commandType = COMMAND_PRIVMSG
	case "ping":
		self.commandType = COMMAND_PING
	case "pong":
		self.commandType = COMMAND_PONG
	case "me":
		self.commandType = COMMAND_ME
	case "topic":
		self.commandType = COMMAND_TOPIC
	case "part":
		self.commandType = COMMAND_PART
	case "quit":
		self.commandType = COMMAND_QUIT
	case "ison":
		self.commandType = COMMAND_ISON
	case "pass":
		self.commandType = COMMAND_PASS
	case "server":
		self.commandType = COMMAND_SERVER
	case "oper":
		self.commandType = COMMAND_OPER
	case "names":
		self.commandType = COMMAND_NAMES
	case "list":
		self.commandType = COMMAND_LIST
	case "invite":
		self.commandType = COMMAND_INVITE
	case "kick":
		self.commandType = COMMAND_KICK
	case "version":
		self.commandType = COMMAND_VERSION
	case "stats":
		self.commandType = COMMAND_STATS
	case "links":
		self.commandType = COMMAND_LINKS
	case "time":
		self.commandType = COMMAND_TIME
	case "connect":
		self.commandType = COMMAND_CONNECT
	case "trace":
		self.commandType = COMMAND_TRACE
	case "admin":
		self.commandType = COMMAND_ADMIN
	case "info":
		self.commandType = COMMAND_INFO
	case "notice":
		self.commandType = COMMAND_NOTICE
	case "whowas":
		self.commandType = COMMAND_WHOWAS
	case "kill":
		self.commandType = COMMAND_KILL
	case "error":
		self.commandType = COMMAND_ERROR
	case "optionals":
		self.commandType = COMMAND_OPTIONALS
	case "away":
		self.commandType = COMMAND_AWAY
	default:
		self.commandType = COMMAND_UNKNOWN
	}

	return nil
}

func (self *Message) GetCommandType() MessageType {
	return self.commandType
}

func (self *Message) SetParameters(params [][]byte) {
	self.parameters = params
}

func (self *Message) GetParameters() [][]byte {
	return self.parameters
}

func (self *Message) GetParameter(offset int) []byte {
	if offset < len(self.parameters) {
		return self.parameters[offset]
	}

	return nil
}

