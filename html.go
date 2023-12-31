package main

import (
	"fmt"
	"time"
)

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
			<th>Time</th>
			<th>RemoteAddr</th>
			<th>Saying</th>
		</tr>
	</thead>
	<tbody>
		%s
	</tbody>`
	row := `<tr>
		<td>%s</td>
		<td>%s</td>
		<td>%s</td>
	</tr>`
	rows := make([]byte, 0)
	for _, v := range vals {
		if v.Saying == "" {
			v.Saying = "..."
		}
		rows = append(rows, []byte(
			fmt.Sprintf(
				row,
				v.Time.Format(time.UnixDate),
				v.RemoteAddr,
				v.Saying))...)
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
	  table{border-collapse:collapse;}
	  th,td{padding:0.5rem;}
	  button{padding:0.5rem 1rem;background-color:%v;color:%v;border:none;border-radius:0.25rem;cursor:pointer;}
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
  <div id="runtime" hx-trigger="every 2s" hx-get="/runtime">runtime stats</div>
  <div class="target" id="target"></div>
  <div id="content">
  <div id="guests"></div>
	<form hx-post="/submit" hx-target="#target" hx-swap="innerHTML">
		<label for="name">Name:</label>
		<input type="text" name="name" id="name">

		<label for="email">Email:</label>
		<input type="email" name="email" id="email">

		<label for="saying">Saying:</label>
		<input type="text" name="saying" id="saying">

		<button type="submit">Submit</button>
		<input type="file" name="file" id="file">
		<button hx-post="/upload" hx-encoding="multipart/form-data" id="uploadJob">upload</button>
		<button onclick="htmx.trigger('#uploadJob', 'htmx.abort')">cancel</button>
	</form>
	</div>
	<div hx-ws="connect:ws://localhost:8080/ws">
		<form hx-ws="send">
			<input type="text" name="saying" id="saying">
			<button type="submit">say</button>
		</form>
	</div>
	%v
	<script>

    </script>
</body>
</html>`
