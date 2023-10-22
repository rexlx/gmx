package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	bs := BasicStyle{
		BodyBG:   "#f5f5f5",
		BodyText: "#333",
		H1:       "#444",
		Btn:      "#becdc3",
		BtnText:  "#000",
	}
	app := NewApplication(bs)

	go func() {
		for range time.Tick(time.Second * 2) {
			app.updateUptime()
		}
	}()

	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", app.Server)
}
