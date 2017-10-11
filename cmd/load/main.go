package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/serge-v/stv/channel"
	"github.com/serge-v/stv/channel/smithsonian"
)

type Response struct {
	Count int
	List  []channel.Item
}

func createList() {
	resp := Response{}

	local, _ := filepath.Glob(cacheDir + "/*.mp4")
	for _, path := range local {
		fi, err := os.Stat(path)
		if err != nil {
			panic(err)
		}
		name := filepath.Base(path)
		href := filepath.Base(path)
		titleFname := strings.TrimSuffix(path, ".mp4") + ".info"
		buf, err := ioutil.ReadFile(titleFname)
		title := ""
		if err != nil {
			title = name
		} else {
			title = string(buf)
		}
		it := channel.Item{
			Name: title,
			Href: "http://wet.voilokov.com:8085/stv/file/" + href,
			Size: fi.Size(),
		}
		resp.List = append(resp.List, it)
	}

	buf, err := json.Marshal(&resp)
	if err != nil {
		panic(err)
	}

	var respcli Response
	if err := json.Unmarshal(buf, &respcli); err != nil {
		panic(err)
	}

	fname := os.Getenv("HOME") + "/.local/dl/list.json"
	err = ioutil.WriteFile(fname, buf, 0666)
	if err != nil {
		panic(err)
	}
}

func configHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "{}")
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fname := cacheDir + "/list.json"
	http.ServeFile(w, r, fname)
}

func runFileServer() {
	http.Handle("/stv/file/", http.StripPrefix("/stv/file/", http.FileServer(http.Dir(cacheDir))))
	http.HandleFunc("/stv/list", listHandler)
	http.HandleFunc("/stv/config", configHandler)
	http.ListenAndServe(":8085", nil)
}

var server = flag.Bool("server", false, "start server")
var list = flag.Bool("list", false, "create list")
var load = flag.Bool("load", false, "load one item")
var cacheDir = os.Getenv("HOME") + "/.local/dl"

func main() {
	flag.Parse()

	if *server {
		runFileServer()
		return
	}

	if *load {
		err := smithsonian.CacheVideos()
		if err != nil {
			panic(err)
		}
	}

	createList()
}
