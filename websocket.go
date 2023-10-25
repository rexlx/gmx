package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WsMsg struct {
	Time       time.Time `json:"time"`
	Name       string    `json:"name"`
	Saying     string    `json:"saying"`
	Email      string    `json:"email"`
	RemoteAddr string    `json:"remoteAddr"`
}

type WsHandler struct {
	mx       *sync.RWMutex
	conn     *websocket.Conn
	messages chan []byte
}

func (wh *WsHandler) Write(app *Applcation, msg WsMsg) error {
	wh.mx.Lock()
	defer wh.mx.Unlock()
	fmt.Println("WsHandler: writing message", msg)
	m := `<div id="guests" hx-swap-oob='true'> %s </div>`
	html := fmt.Sprintf(m, VisitorsToHTMLTable(app.Visitors))
	return wh.conn.WriteMessage(websocket.TextMessage, []byte(html))
}

// TODO: youre only writing the saying!
func serveWs(app *Applcation, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("serveWs", err)
		return
	}

	fmt.Println("serveWS: websocket connection", r.RemoteAddr)

	m := make(chan []byte, 100)
	wh := WsHandler{
		mx:       app.Mux,
		conn:     conn,
		messages: m,
	}

	go func() {
		for {
			msgType, msg, err := wh.conn.ReadMessage()
			if err != nil {
				fmt.Println("serveWs", err)
				break
			}
			fmt.Println("serveWs: got type", msgType)
			// wh.mx.Lock()
			wh.messages <- msg
			// wh.mx.Unlock()
		}
	}()

	for msg := range wh.messages {
		fmt.Println("serveWs: got message", string(msg))

		var newMsg WsMsg

		err := json.Unmarshal(msg, &newMsg)
		if err != nil {
			fmt.Println("serveWs", err)
			break
		}
		newMsg.RemoteAddr = r.RemoteAddr
		newMsg.Time = time.Now()
		app.AddVisitor(Visitor{
			Saying:     newMsg.Saying,
			Time:       newMsg.Time,
			RemoteAddr: newMsg.RemoteAddr,
		})

		wh.Write(app, newMsg)
	}
}
