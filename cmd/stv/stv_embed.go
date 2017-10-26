package main

const (
	genericHTML = `<html>
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
)

const (
	mainHTML = `<html>
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
	<table border="0" style="width: 240px">
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
)

const (
	playHTML = `<html>
<head>
<title>Videos</title>
<meta name="viewport" content="width=device-width; maximum-scale=1; minimum-scale=1;" />
<style>
.button {
    background-color: lightgray;
    border: none;
    color: white;
    padding: 15px 32px;
    text-align: center;
    vertical-align: middle;
    text-decoration: none;
    display: inline-block;
    font-size: 32px;
    width: 100px;
    height: 80px;
}

.left {
    background-color: #6C70BF;
}

.right {
    background-color: #4C50AF;
}

</style>
</head>
<body>
	<table border="0" style="width: 200px">
		{{if .Error}}
		<tr><td style="color: red">ERROR: {{.Error}}</td><tr>
		{{end}}
		<tr>
			<td><a href="/" class="button left">Stop</a></td>
			<td><a href="/play?cmd=pause" class="button right">Pause</a></td>
		<tr>
		<tr>
			<td><a href="/play?cmd=volume&arg=-10" class="button left">Vol -</a></td>
			<td><a href="/play?cmd=volume&arg=10" class="button right">Vol +</a></td>
		</tr>
		<tr>
			<td><a href="/play?cmd=seek&arg=-30" class="button left">-30</a></td>
			<td><a href="/play?cmd=seek&arg=30" class="button right">+30</a></td>
		</tr>
		<tr>
			<td><a href="/play?cmd=seek&arg=-600" class="button left">-600</a></td>
			<td><a href="/play?cmd=seek&arg=600" class="button right">+600</a></td>
		</tr>
	</table>
</body>
</html>
`
)
