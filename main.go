package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Applcation struct {
	BasicStyle
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
	Email      string
	RemoteAddr string
	Time       time.Time
}

func main() {
	bs := BasicStyle{
		BodyBG:   "#f5f5f5",
		BodyText: "#333",
		H1:       "#444",
		Btn:      "#333",
		BtnText:  "#f7f7f7",
	}
	app := NewApplication(bs)

	fmt.Println("Listening on port 8080...")
	http.ListenAndServe(":8080", app.Server)
}

func NewApplication(bs BasicStyle) *Applcation {
	var app Applcation
	app.BasicStyle = bs
	app.Count = make(map[string]int)
	app.Visitos = make([]Visitor, 0)
	app.Server = http.NewServeMux()
	app.Mux = sync.RWMutex{}

	app.Server.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, fmt.Sprintf(splashPage, addMinimalStyling(bs)))
	})

	app.Server.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		visitors := app.GetVisitors()
		tbl := VisitorsToHTMLTable(visitors)
		fmt.Fprintf(w, tbl)
	})

	app.Server.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("name")
		email := r.FormValue("email")
		remoteAddr := r.RemoteAddr
		t := time.Now()
		v := Visitor{
			Name:       name,
			Email:      email,
			RemoteAddr: remoteAddr,
			Time:       t,
		}
		go app.AddVisitor(v)

		fmt.Fprintf(w, `
			<h1>thanks for your submission!</h1>

			<p>name: %s</p>
			<p>email: %s</p>
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

func VisitorsToCSV(vals []Visitor) []byte {
	line := "%s,%s,%s,%s\n"
	csvHeader := []byte("Name,Email,RemoteAddr,Time\n")
	csv := make([]byte, 0)
	csv = append(csv, csvHeader...)
	for _, v := range vals {
		csv = append(csv, []byte(fmt.Sprintf(line, v.Name, v.Email, v.RemoteAddr, v.Time))...)
	}
	return csv
}

func VisitorsToHTMLTable(vals []Visitor) string {
	table := `<table>
	<thead>
		<tr>
			<th>Name</th>
			<th>Email</th>
			<th>RemoteAddr</th>
			<th>Time</th>
		</tr>
	</thead>
	<tbody>
		%s
	</tbody>`
	row := `<tr>
		<td>%s</td>
		<td>%s</td>
		<td>%s</td>
		<td>%s</td>
	</tr>`
	rows := make([]byte, 0)
	for _, v := range vals {
		rows = append(rows, []byte(fmt.Sprintf(row, v.Name, v.Email, v.RemoteAddr, v.Time.Format(time.UnixDate)))...)
	}
	table = fmt.Sprintf(table, rows)
	return table
}

func addMinimalStyling(bs BasicStyle) string {
	styleString := `
	<style>
	  body{font-family:Arial,Helvetica,sans-serif;font-size:16px;line-height:1.5;margin:0;padding:0;background-color:%v;color:%v;}
	  h1{font-size:2rem;margin-bottom:1rem;color:%v;}
	  form{display:flex;flex-direction:column;max-width:400px;margin:0 auto;}
	  label{margin-bottom:0.5rem;}input{padding:0.5rem;margin-bottom:1rem;border-radius:0.25rem;border:1px solid #ccc;}
	  button{padding:0.5rem 1rem;background-color:%v;color:%v;border:none;border-radius:0.25rem;cursor:pointer;}
	  button:hover{background-color:#444;}
	  .target{margin-top:2rem;overflow-y:scroll;}
	</style>`
	return fmt.Sprintf(styleString, bs.BodyBG, bs.BodyText, bs.H1, bs.Btn, bs.BtnText)
}

var splashPage = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>w e l c o m e</title>
  <script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
</head>
<body>
  <h1>thanks for visiting!</h1>
  <div class="target" id="target"></div>
  <button hx-get="/visitors" hx-target="#guests" hx-swap="innerHTML">see guests</button>
  <div id="guests"></div>
  <div id="content">
  <form hx-post="/submit" hx-target="#target" hx-swap="innerHTML">
  <label for="name">Name:</label>
  <input type="text" name="name" id="name">

  <label for="email">Email:</label>
  <input type="email" name="email" id="email">

  <button type="submit">Submit</button>
	</div>
	</form>
	%v
	<script>
  </script>
</body>
</html>`
