package structs

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Callback func()

type Connection struct {
	WS   *websocket.Conn
	Mu   sync.Mutex
	Send chan []byte
}

type Message struct {
	Data []byte
	Room string
}

type Subscription struct {
	Conn      *Connection
	Room      string
	TypeQueue string
	Id        string
}

type hub struct {
	rooms      map[string]map[*Connection]bool
	Broadcast  chan Message
	Register   chan Subscription
	Unregister chan Subscription
	ID         string
}

var HubRoom = hub{
	Broadcast:  make(chan Message),
	Register:   make(chan Subscription),
	Unregister: make(chan Subscription),
	rooms:      make(map[string]map[*Connection]bool),
}

func (h *hub) Run(id string) {
	h.ID = id
	for {
		select {
		case s := <-h.Register:
			connections := h.rooms[s.Room]

			if connections == nil {
				connections = make(map[*Connection]bool)
				h.rooms[s.Room] = connections
			}

			h.rooms[s.Room][s.Conn] = true

			RedisQueue.Client.RPush(
				"wss:"+s.Room+":"+s.TypeQueue,
				s.Id,
			)

			RedisQueue.Client.SAdd(
				"connections:"+h.ID,
				s.Id,
			)

		case s := <-h.Unregister:
			connections := h.rooms[s.Room]

			if connections != nil {
				if _, ok := connections[s.Conn]; ok {
					delete(connections, s.Conn)
					close(s.Conn.Send)

					if len(connections) == 0 {
						delete(h.rooms, s.Room)
					}

					RedisQueue.Client.LRem(
						"wss:"+s.Room+":"+s.TypeQueue,
						1,
						s.Id,
					)

					RedisQueue.Client.SRem(
						"connections:"+h.ID,
						s.Id,
					)
				}
			}

		case m := <-h.Broadcast:
			connections := h.rooms[m.Room]

			for c := range connections {
				select {
				case c.Send <- m.Data:
				default:
					close(c.Send)
					delete(connections, c)

					if len(connections) == 0 {
						delete(h.rooms, m.Room)
					}
				}
			}
		}
	}
}
