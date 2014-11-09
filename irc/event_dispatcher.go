package irc

import (
	"sort"
)

type Event struct {
	Priority int
	Callback func(args interface{}) bool
}

type Events []*Event

func (p Events) Len() int {
	return len(p)
}

func (p Events) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p Events) Less(i, j int) bool {
	return p[i].Priority < p[j].Priority
}


type EventDispatcher struct {
	Subscribers map[string]Events
}

func NewEventDispatcher() *EventDispatcher{
	return &EventDispatcher{
		Subscribers: make(map[string]Events),
	}
}

func (self *EventDispatcher) Subscribe(event string, priority int, callback func(args interface{}) bool) {
	self.Subscribers[event] = append(self.Subscribers[event], &Event{
			Priority: priority,
			Callback: callback,
	})

	sort.Sort(sort.Reverse(self.Subscribers[event]))
}

func (self *EventDispatcher) Dispatch(event string, args interface{}) {
	if _, ok := self.Subscribers[event]; ok {
		for _, ev := range self.Subscribers[event] {
			ret := ev.Callback(args)

			if !ret {
				break
			}
		}
	}
}
