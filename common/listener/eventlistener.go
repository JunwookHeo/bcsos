package listener

type EventListener struct {
	name      string
	listeners []chan string
}

func (el *EventListener) AddListener(ch chan string) {
	if el.listeners == nil {
		el.listeners = make([] chan string, 0)
	}

	el.listeners = append(el.listeners, ch)
}

func (el *EventListener) RemoveListener(ch chan string) {
	for i := range el.listeners{
		if el.listeners[i] == ch {
			el.listeners = append(el.listeners[:i], el.listeners[i+1:]...)
		}
	}
}

func (el *EventListener) Notify(res string) {
	for _, handler := range el.listeners {
		go func(handler chan string){
			handler <- res
		}(handler)
	}
}