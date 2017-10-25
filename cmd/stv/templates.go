package main

import (
	"html/template"

	"github.com/serge-v/stv/channel"
)

type mainData struct {
	Error error
	List  []channel.Item
}

type genericData struct {
	Error error
}

var (
	mainTemplate    = template.Must(template.New("main").Parse(mainHTML))
	genericTemplate = template.Must(template.New("generic").Parse(genericHTML))
	playTemplate    = template.Must(template.New("play").Parse(playHTML))
)
