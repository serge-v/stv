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

func testSerialization() {
	resp := Response{
		Count: 1,
		List:  []channel.Item{channel.Item{Name: "test", Href: "testhref"}},
	}

	local, _ := filepath.Glob(cacheDir + "/*.mp4")
	for _, path := range local {
		name := filepath.Base(path)
		href := filepath.Base(path)
		name = strings.TrimSuffix(name, ".mp4")
		it := channel.Item{
			Name: name,
			Href: "http://localhost:8085/stv/file/" + href,
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
	ioutil.WriteFile(fname, buf, 0666)
	fmt.Printf("%+v\n", respcli)
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
var test = flag.Bool("test", false, "run test")
var cacheDir = os.Getenv("HOME") + "/.local/dl"

func main() {
	flag.Parse()
	if *test {
		testSerialization()
		return
	}
	if *server {
		runFileServer()
		return
	}

	list, err := smithsonian.GetVideos()
	if err != nil {
		panic(err)
	}
	resp := Response{
		Count: len(list),
		List:  list,
	}
	for idx, item := range resp.List {
		fmt.Printf("%d. %s\n", idx+1, item.Name)
	}
}
