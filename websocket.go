package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WsMsg struct {
	Name       string
	Saying     string
	Email      string
	RemoteAddr string
}

func serveWs(app *Applcation, w http.ResponseWriter, r *http.Request) {
	fmt.Println("websocket connection", r.RemoteAddr)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("serveWs", err)
		return
	}
	go wsWriter(app, conn)
	go wsReader(app, conn)
}

func wsWriter(app *Applcation, conn *websocket.Conn) {
	for {
		msg := <-app.WsChan
		if err := conn.WriteJSON(msg); err != nil {
			fmt.Println("wsWriter", err)
			return
		}
	}
}

func wsReader(app *Applcation, conn *websocket.Conn) {
	for {
		var msg WsMsg
		if err := conn.ReadJSON(&msg); err != nil {
			fmt.Println("wsReader", err)
			return
		}
		app.Mux.Lock()
		app.Visitors = append(app.Visitors, Visitor{
			Name:       msg.Name,
			Saying:     msg.Saying,
			Email:      msg.Email,
			RemoteAddr: msg.RemoteAddr,
			Time:       time.Now(),
		})
		app.Count[msg.Name]++
		app.Mux.Unlock()
		fmt.Printf("wsReader\t%+v\n", msg)
	}
}
