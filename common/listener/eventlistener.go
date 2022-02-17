package listener

import (
	"sync"
)

type EventListener struct {
	name      string
	listeners []chan string
}

var (
	el   *EventListener
	once sync.Once
)

func (el *EventListener) AddListener(ch chan string) {
	if el.listeners == nil {
		el.listeners = make([]chan string, 0)
	}

	el.listeners = append(el.listeners, ch)
}

func (el *EventListener) RemoveListener(ch chan string) {
	for i := range el.listeners {
		if el.listeners[i] == ch {
			el.listeners = append(el.listeners[:i], el.listeners[i+1:]...)
		}
	}
}

func (el *EventListener) Notify(res string) {
	for _, handler := range el.listeners {
		go func(handler chan string) {
			handler <- res
		}(handler)
	}
}

func EventListenerInst() *EventListener {
	once.Do(func() {
		el = &EventListener{
			name:      "EventListenr",
			listeners: nil,
		}
	})

	return el
}
