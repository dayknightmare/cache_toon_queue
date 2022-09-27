package sockets

import (
	"log"
	"time"

	"github.com/Vupy/cache-toon-queue/src/structs"
	"github.com/gorilla/websocket"
)

const (
	pongWait       = 15 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

func ReadWSMessage(sub structs.Subscription, callback structs.Callback) {
	conn := sub.Conn

	defer func() {
		structs.HubRoom.Unregister <- sub
		conn.WS.Close()
		callback()
	}()

	conn.WS.SetReadLimit(maxMessageSize)
	conn.WS.SetReadDeadline(time.Now().Add(pongWait))
	conn.WS.SetPongHandler(func(string) error {
		return conn.WS.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, jsonMsg, err := conn.WS.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("unexpected close error: %v", err)
			}
			break
		}

		m := structs.Message{Data: jsonMsg, Room: sub.Room}

		structs.HubRoom.Broadcast <- m
	}
}

func write(c *structs.Connection, mt int, payload []byte) error {
	defer c.Mu.Unlock()
	return c.WS.WriteMessage(mt, payload)
}

func WriteWSMessage(sub *structs.Subscription) {
	conn := sub.Conn
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		conn.WS.Close()
	}()

	for {
		select {
		case msg, ok := <-conn.Send:
			conn.Mu.Lock()
			conn.WS.SetReadDeadline(time.Now().Add(pongWait))

			if !ok {
				write(conn, websocket.CloseMessage, []byte{})
				return
			}

			if err := write(conn, websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			conn.Mu.Lock()
			if err := conn.WS.WriteMessage(websocket.PingMessage, []byte("Ping")); err != nil {
				conn.Mu.Unlock()
				return
			}
			conn.Mu.Unlock()
		}
	}
}
