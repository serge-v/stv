package main

import (
	"html/template"

	"github.com/serge-v/stv/channel"
)

type mainData struct {
	Error error
	List  []channel.Item
}

var mainPageHTML = `
<html>
<head>
<title>Videos</title>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style>
.button {
    background-color: #4CAF50; /* Green */
    border: none;
    color: white;
    padding: 15px 32px;
    text-align: center;
    vertical-align: middle;
    text-decoration: none;
    display: inline-block;
    font-size: 32px;
    width: 240px;
    height: 100px;
}
</style>
</head>
<body>
	{{if .Error}}
	<table border="0" style="width: 260px">
		<tr><td style="color: red">ERROR: {{.Error}}</td><tr>
		<tr><td><a class="button" href="/shutdown">Shutdown</a></td><tr>
		<tr><td><a class="button" href="/restart">Restart</a></td><tr>
		<tr><td><a class="button" href="/play">Test play</a></td><tr>
	</table>
	{{end}}
	<table border="0">
		{{range .List}}
		<tr><td><a class="button" href="/play?id={{.ID}}">{{.Title}}</a></td><tr>
		{{end}}
	</table>
</body>
</html>
`

type genericData struct {
	Error error
}

var genericPageHTML = `
<html>
<head>
<title>Videos</title>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style>
.button {
    background-color: #4C50AF;
    border: none;
    color: white;
    padding: 15px 32px;
    text-align: center;
    vertical-align: middle;
    text-decoration: none;
    display: inline-block;
    font-size: 32px;
    width: 240px;
    height: 100px;
}
</style>
</head>
<body>
	<table border="0" style="width: 260px">
		{{if .Error}}
		<tr><td style="color: red">ERROR: {{.Error}}</td><tr>
		{{end}}
		<tr><td><a href="/" class="button">Stop</a></td><tr>
	</table>
</body>
</html>
`

var playPageHTML = `
<html>
<head>
<title>Videos</title>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style>
.button {
    background-color: #4C50AF;
    border: none;
    color: white;
    padding: 15px 32px;
    text-align: center;
    vertical-align: middle;
    text-decoration: none;
    display: inline-block;
    font-size: 32px;
    width: 240px;
    height: 100px;
}
</style>
</head>
<body>
	<table border="0" style="width: 260px">
		{{if .Error}}
		<tr><td style="color: red">ERROR: {{.Error}}</td><tr>
		{{end}}
		<tr><td><a href="/" class="button">Stop</a></td><tr>
		<tr><td><a href="/play?seek=-60s" class="button">Seek -60</a></td></tr>
		<tr><td><a href="/play?seek=60s" class="button">Seek +60</a></td></tr>
		<tr><td><a href="/play?vol=+10" class="button">Vol -10</a></td></tr>
		<tr><td><a href="/play?vol=-10" class="button">Vol +10</a></td></tr>
	</table>
</body>
</html>
`

var mt = template.Must(template.New("main").Parse(mainPageHTML))
var gt = template.Must(template.New("generic").Parse(genericPageHTML))
var playTemplate = template.Must(template.New("play").Parse(playPageHTML))
