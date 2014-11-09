package main

import (
	"fmt"
	"bytes"
	"net"
	"regexp"
	"time"
	"strings"
	"encoding/json"

	"io/ioutil"
	"gopkg.in/yaml.v2"

	"github.com/chobie/gtig/irc"
	"github.com/chobie/gtig/twitter"
)


type OAuth struct {
	Token string
	Secret string
}

type Config struct {
	OAuth *OAuth `yaml:"oauth"`
}

func main() {
	data, err := ioutil.ReadFile("config.yml")

	if err != nil {
		panic(err)
	}
	config := &Config{
	}
	err = yaml.Unmarshal([]byte(data), config)
	if err != nil {
		panic(err)
	}

	world := irc.NewWorld()
	client := twitter.NewClient(config.OAuth.Token, config.OAuth.Secret)

	// privmsg (ctcp action)
	world.EventDispatcher.Subscribe("kernel.privmsg", 500, func(args interface{}) bool {
		if ev, ok := args.(*irc.NewMessageEvent); ok {
			if len(ev.Params) > 1 {
				if regexp.MustCompile("ACTION").Match(ev.Params[1]) {

					args := bytes.Split(ev.Params[1], []byte(" "))
					fmt.Printf("ACTION!: %s\n", args[1:])

					command := string(args[1])
					fmt.Printf("COMMAND: %s\n", command)

					switch command {
					case "favorite", "fav":
						id := args[2]
						client.Favorite(string(id))
					case "retweet", "rt":
						id := args[2]
						client.Retweet(string(id))
					case "search":
						word := args[2]

						fmt.Printf("search word; %s\n", word)
						world.EventDispatcher.Dispatch("irc.kernel.room.create", &irc.NewRoomEvent{
							Room: []byte(fmt.Sprintf("#search-%s", word)),
							//User: user,
							Callback: func(room *irc.Room) {
								l, err := client.Search(string(word))
								if err != nil {
									fmt.Printf("Error: %s\n", err)
									return
								}

								for i := len(l.Statuses)-1; i >= 0; i--  {
									v := l.Statuses[i]
									if v.Id > room.Last {
										fmt.Printf("%d: %s > %s\n", v.Id, v.User.ScreenName, v.Text)

										u := &irc.User{
											Name: []byte(v.User.ScreenName),
										}
										world.EventDispatcher.Dispatch("irc.kernel.privmsg", &irc.NewMessageEvent{
												User:    u,
												Room:    room.Name,
												Message: []byte(v.Text),
										})
										room.Last = v.Id
									}
								}
							},
						})

						world.EventDispatcher.Dispatch("irc.kernel.room.invite", &irc.InviteEvent{
							From: []byte("system"),
							To: []byte("chobi_e"),
							Room: []byte(fmt.Sprintf("#search-%s", word)),
						})
					case "re":
					case "show":
					case "follow":
					case "unfollow":
					case "block":
					case "unblock":
					case "list":
						//add, create, remove
					case "config":
						// dump, save
					}

					return false
				}
			}
		}
		return true
	})

	// privmsg (send to twitter)
	world.EventDispatcher.Subscribe("kernel.privmsg", 500, func(args interface{}) bool {
		if ev, ok := args.(*irc.NewMessageEvent); ok {
			if string(ev.User.Name) == "chobi_e" && !ev.External && string(ev.Room) == "#twitter" {
				client.Tweet(string(ev.Message))
			}
		}
		return true
	})


	ln, err := net.Listen("tcp", ":6668")
	if err != nil {
		panic(err)
	}

	go twitter_stream(world, config)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}

		go world.HandleConnection(conn)
	}
}

func twitter_stream(world *irc.World, config *Config) {
	client := twitter.NewClient(config.OAuth.Token, config.OAuth.Secret)

	for {
		stream := client.UserStream()
		for {
			ev, err := stream.Next()
			if err != nil {
				fmt.Printf("CLOSE EVENT: %s\n", err)
				stream.Close()
				time.Sleep(time.Second)
				break
			}

			switch ev.EventType() {
			case twitter.TWEET_EVENT:
				var text string
				if ev.RetweetedStatus != nil {
					text = fmt.Sprintf(" (%s) RT %s", ev.RetweetedStatus.User.ScreenName, ev.RetweetedStatus.Text)
				} else {
					text = ev.Text
				}

				for _, v := range ev.Entities.Urls {
					text = strings.Replace(text, v.Url, v.ExpandedUrl, -1)
				}
				for _, v := range ev.Entities.Media {
					text = strings.Replace(text, v.Url, v.MediaUrl, -1)
				}
				text = fmt.Sprintf("%s [%d]", text, ev.Id)

				if ev.RetweetedStatus != nil {
					world.SendPrivMessage(ev.RetweetedStatus.User.ScreenName, "#retweeted", text[3:])
				}

				world.SendPrivMessage(ev.User.ScreenName, "#twitter", text)
			case twitter.FAVORITE_EVENT:
				text := ev.TargetObject.(map[string]interface{})["text"].(string)
				screen_name := ev.TargetObject.(map[string]interface{})["user"].(map[string]interface{})["screen_name"].(string)
				fmt.Printf("# FAV %s > %s\n", screen_name, text)

				if _, ok := ev.TargetObject.(map[string]interface{})["entities"].((map[string]interface{}))["urls"]; ok {
					for _, v := range ev.TargetObject.(map[string]interface{})["entities"].((map[string]interface{}))["urls"].([]interface{}) {
						text = strings.Replace(text, v.(map[string]interface{})["url"].(string), v.(map[string]interface{})["expanded_url"].(string), -1)
					}
				}

				if _, ok := ev.TargetObject.(map[string]interface{})["entities"].((map[string]interface{}))["media"]; ok {
					for _, v := range ev.TargetObject.(map[string]interface{})["entities"].((map[string]interface{}))["media"].([]interface{}) {
						text = strings.Replace(text, v.(map[string]interface{})["url"].(string), v.(map[string]interface{})["media_url"].(string), -1)
					}
				}

				world.SendPrivMessage(screen_name, "#favorited", text)
			case twitter.KEEPALIVE_EVENT:
			default:
				b, _ := json.MarshalIndent(ev, "", "  ")
				fmt.Printf("%s\n", b)
			}
		}
	}
}
