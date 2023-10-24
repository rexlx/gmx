package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// var upgrader = websocket.Upgrader{
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

type Applcation struct {
	BasicStyle
	WsChan     chan WsMsg
	Uptime     time.Duration
	TableCache []Visitor
	Details    *RuntimeDetails
	Server     *http.ServeMux
	Mux        sync.RWMutex
	Visitors   []Visitor
	Count      map[string]int
}

type BasicStyle struct {
	BodyBG   string
	BodyText string
	H1       string
	Btn      string
	BtnText  string
}

type Visitor struct {
	Name       string
	Saying     string
	Email      string
	RemoteAddr string
	Time       time.Time
}

type RuntimeDetails struct {
	startTime time.Time
	runtTime  time.Duration
}

func NewApplication(bs BasicStyle) *Applcation {
	var app Applcation
	app.BasicStyle = bs
	app.Count = make(map[string]int)
	app.Visitors = make([]Visitor, 0)
	app.TableCache = make([]Visitor, 0)
	app.Server = http.NewServeMux()
	app.Mux = sync.RWMutex{}
	app.Details = &RuntimeDetails{}
	app.Details.start()

	app.Server.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(&app, w, r)
	})

	app.Server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(time.Now(), r.Method, r.URL.Path)
		fmt.Fprintf(w, fmt.Sprintf(splashPage, addMinimalStyling(app.BasicStyle)))
	})

	app.Server.HandleFunc("/runtime", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(time.Now(), r.Method, r.URL.Path)
		out := fmt.Sprintf("<small>uptime: %v</small>", app.Uptime)
		fmt.Fprintf(w, out)
	})

	app.Server.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(time.Now(), r.Method, r.URL.Path)
		visitors := app.GetVisitors()
		tbl := VisitorsToHTMLTable(visitors)
		fmt.Fprintf(w, tbl)
	})

	app.Server.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(time.Now(), r.Method, r.URL.Path)
		name := r.FormValue("name")
		email := r.FormValue("email")
		saying := r.FormValue("saying")
		remoteAddr := r.RemoteAddr
		t := time.Now()
		v := Visitor{
			Name:       name,
			Email:      email,
			RemoteAddr: remoteAddr,
			Time:       t,
			Saying:     saying,
		}
		go app.AddVisitor(v)

		fmt.Fprintf(w, `
			<h1>thanks for your submission %s!</h1>
			<small>(%s)</small>
			<hr>
		`, name, email)
	})

	app.Server.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		var maxUploadSize int64 = 5 * (1024 * 1024) // 5 mb
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		// r.Header.Add("Content-Type", "multipart/form-data")
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			fmt.Fprintf(w, "file too large. max size: %v, %v", maxUploadSize, err)
			return
		}
		file, fh, err := r.FormFile("file")
		if err != nil {
			fmt.Fprintf(w, "error reading file: %v", err)
			return
		}
		defer file.Close()
		fmt.Printf("uploaded file: %v\n", fh.Filename)
		dst, err := os.Create(filepath.Join("/Volumes/_rxlx/", fh.Filename))
		if err != nil {
			fmt.Fprintf(w, "error creating file: %v", err)
			return
		}
		defer dst.Close()
		_, err = io.Copy(dst, file)
		if err != nil {
			fmt.Fprintf(w, "error copying file: %v", err)
			return
		}
		fmt.Fprintf(w, "file uploaded")
	})
	return &app
}

func (a *Applcation) AddVisitor(v Visitor) {
	a.Mux.Lock()
	defer a.Mux.Unlock()
	a.Visitors = append(a.Visitors, v)
	a.Count[v.Email]++
}

func (a *Applcation) updateUptime() {
	a.Mux.Lock()
	defer a.Mux.Unlock()
	a.Uptime = time.Since(a.Details.startTime)
}

func (a *Applcation) updateTableCache(n int) {
	a.Mux.Lock()
	defer a.Mux.Unlock()
	// the table cache is the 100 most recent visitors
	if len(a.Visitors) < n {
		n = len(a.Visitors)
	}
	a.TableCache = a.Visitors[len(a.Visitors)-n:]
}

func (a *Applcation) GetVisitors() []Visitor {
	return a.TableCache
}

func (rt *RuntimeDetails) updateRuntime() {
	rt.runtTime = time.Since(rt.startTime)
}

func (rt *RuntimeDetails) start() {
	rt.startTime = time.Now()
}
