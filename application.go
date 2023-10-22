package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Applcation struct {
	BasicStyle
	Uptime  time.Duration
	Details *RuntimeDetails
	Server  *http.ServeMux
	Mux     sync.RWMutex
	Visitos []Visitor
	Count   map[string]int
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
	app.Visitos = make([]Visitor, 0)
	app.Server = http.NewServeMux()
	app.Mux = sync.RWMutex{}
	app.Details = &RuntimeDetails{}
	app.Details.start()

	app.Server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL.Path)
		fmt.Fprintf(w, fmt.Sprintf(splashPage, addMinimalStyling(app.BasicStyle)))
	})

	app.Server.HandleFunc("/runtime", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL.Path)
		out := fmt.Sprintf("<small>uptime: %v</small>", app.Uptime)
		fmt.Fprintf(w, out)
	})

	app.Server.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL.Path)
		visitors := app.GetVisitors()
		tbl := VisitorsToHTMLTable(visitors)
		fmt.Fprintf(w, tbl)
	})

	app.Server.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(time.Now(), r.Method, r.URL.Path)
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
	return &app
}

func (a *Applcation) AddVisitor(v Visitor) {
	a.Mux.Lock()
	a.Visitos = append(a.Visitos, v)
	a.Count[v.Email]++
	a.Mux.Unlock()
}

func (a *Applcation) updateUptime() {
	a.Mux.Lock()
	defer a.Mux.Unlock()
	a.Uptime = time.Since(a.Details.startTime)
}

func (a *Applcation) GetVisitors() []Visitor {
	a.Mux.RLock()
	defer a.Mux.RUnlock()
	return a.Visitos
}

func (a *Applcation) GetCount() {
	a.Mux.RLock()
	defer a.Mux.RUnlock()
	for k, v := range a.Count {
		fmt.Printf("%v: %v\n", k, v)
	}
}

func (rt *RuntimeDetails) String() string {
	var t time.Time
	if rt.startTime == t {
		return "not started"
	}
	rt.updateRuntime()
	return fmt.Sprintf("started at %v, running for %v", rt.startTime.Format(time.UnixDate), rt.runtTime)
}

func (rt *RuntimeDetails) updateRuntime() {
	rt.runtTime = time.Since(rt.startTime)
}

func (rt *RuntimeDetails) start() {
	rt.startTime = time.Now()
}
