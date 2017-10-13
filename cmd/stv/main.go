package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/serge-v/stv/channel"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		return
	}

	stopPlayer()

	var d mainData
	d.List, d.Error = getLocalVideos()

	var rd mainData
	if err := unmarshalURL(listEndpoint, &rd); err != nil {
		rd.List = []channel.Item{channel.Item{Title: err.Error()}}
		log.Println("mainHandler", err.Error())
	}

	d.List = append(d.List, rd.List...)

	if err := mt.Execute(w, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func restartHandler(w http.ResponseWriter, r *http.Request) {
	var d genericData
	d.Error = exec.Command("sudo", "shutdown", "-r", "5").Run()
	if err := gt.Execute(w, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	var d genericData
	d.Error = exec.Command("sudo", "shutdown", "-h", "5").Run()
	if err := gt.Execute(w, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getVideoURL(id string) string {
	var resp mainData
	if err := unmarshalURL(listEndpoint, &resp); err != nil {
		log.Println("getVideoURL", err.Error())
		return ""
	}
	for _, item := range resp.List {
		if id == item.ID {
			return "http://localhost:6061/stv/" + token + "/stv/file/" + id + ".mp4"
		}
	}
	return ""
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	var d genericData

	id := r.URL.Query().Get("id")
	if id != "" {
		link := getVideoURL(id)
		log.Println("starting player:", link)
		d.Error = startPlayer(link)
	}

	seek := r.URL.Query().Get("seek")
	if seek != "" {
		d, _ := time.ParseDuration(seek)
		seekPlayer(d)
	}

	vol := r.URL.Query().Get("vol")
	if vol != "" {
		n, _ := strconv.Atoi(vol)
		changeVolume(n)
	}

	if err := playTemplate.Execute(w, d); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type programState struct {
	Elapsed map[string]int `json:"elapsed"`
}

var state = programState{
	Elapsed: make(map[string]int),
}

func unmarshalURL(srcurl string, v interface{}) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(srcurl)
	//	resp, err := http.Get(srcurl)

	if err != nil {
		log.Println("unmarshalURL:", srcurl, "error:", err.Error())
		return err
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("unmarshalURL:", srcurl, "status:", resp.StatusCode)
		return fmt.Errorf("list: %s", resp.Status)
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err := dec.Decode(v); err != nil {
		log.Println("unmarshalURL", err.Error())
		return err
	}
	return nil
}

func loadState() {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(confapi)
	if err != nil {
		log.Println("loadState", err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		log.Println("loadState status:", resp.StatusCode)
		return
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	if err := dec.Decode(&state); err != nil {
		log.Println("loadState", err.Error())
		return
	}
}

func saveState() {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	if err := enc.Encode(&state); err != nil {
		log.Println(err)
		return
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Post(confapi, "application/json", &b)
	if err != nil {
		log.Println("saveState", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Println("saveState status", resp.Status)
		return
	}
}

var pid int

func stopPlayer() error {
	log.Println("player pid:", pid)
	if pid > 0 {
		cmd := exec.Command("pkill", player)
		err := cmd.Run()
		time.Sleep(time.Second)
		log.Println("kill signal sent")
		return err
	}
	return nil
}

func startPlayer(href string) error {
	err := stopPlayer()
	if err != nil {
		return err
	}

	streamURL := href
	args := append([]string{}, playerArgs...)
	args = append(args, streamURL)
	cmd := exec.Command(player, args...)
	log.Printf("%+v\n", cmd.Args)

	if err = cmd.Start(); err != nil {
		log.Println(err)
	}
	pid = cmd.Process.Pid

	go func() {
		err = cmd.Wait()
		if err != nil {
			println(err.Error())
		}
		pid = 0
		log.Println("player stopped")
		saveState()
	}()

	return err
}

func seekPlayer(d time.Duration) {
	// TODO: call dbus
}

func changeVolume(percent int) {
	// TODO: call dbus
}

func pausePlayer() {
	// TODO: call dbus
}

func getLocalVideos() ([]channel.Item, error) {
	local, _ := filepath.Glob("vid/*")
	list := []channel.Item{}
	for idx, path := range local {
		it := channel.Item{
			ID:    fmt.Sprintf("%d", idx),
			Title: filepath.Base(path),
			Link:  path,
		}
		list = append(list, it)
	}

	return list, nil
}

var debug = flag.Bool("d", false, "debug in console")
var cacheDir = os.Getenv("HOME") + "/.cache/stv"
var configDir = os.Getenv("HOME") + "/.config/stv"
var player = "mplayer"
var playerArgs = []string{"-geometry", "480x240+1920+0"}
var token string
var tokenFname string
var confapi string
var listEndpoint string
var addr = ":6061"

func init() {
	user := os.Getenv("USER")
	if user == "pi" || user == "alarm" {
		player = "omxplayer"
		playerArgs = []string{}
	} else if user == "odroid" {
		player = "vlc"
		playerArgs = []string{}
	}
	tokenFname = configDir + "/token.txt"
}

func createToken() string {
	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		log.Fatal(err)
	}
	s := hex.EncodeToString(b)
	if err := ioutil.WriteFile(tokenFname, []byte(s), 0600); err != nil {
		log.Fatal(err)
	}
	return s
}

var dbusc *dbus.Conn

func main() {
	flag.Parse()
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		log.Fatal(err)
	}

	buf, err := ioutil.ReadFile(tokenFname)
	token = string(buf)
	if os.IsNotExist(err) || len(token) == 0 {
		token = createToken()
	} else if err != nil {
		log.Fatal(err)
	}
	confapi = "https://conf.voilokov.com/" + token + "/stv/config"
	listEndpoint = "https://conf.voilokov.com/" + token + "/stv/list.json"
	proxyURL := "https://conf.voilokov.com/"
	if *debug {
		confapi = "https://conf.svtest.com:9001/" + token + "/stv/config"
		listEndpoint = "https://conf.svtest.com:9001/" + token + "/stv/list.json"
		proxyURL = "https://conf.svtest.com:9001/"
	}
	loadState()

	//	dbusc, err = dbus.SessionBus()
	//	if err != nil {
	//		panic(err)
	//	}

	log.Println("player:", player, "token:", token)
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		log.Println("addr: ", addr.String())
	}

	http.HandleFunc("/shutdown", shutdownHandler)
	http.HandleFunc("/restart", restartHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/", mainHandler)
	u, err := url.Parse(proxyURL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	d := proxy.Director
	proxy.Director = func(r *http.Request) {
		r.Host = u.Host
		log.Println("url1:", r.URL.String())
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/stv")
		log.Println("url2:", r.URL.String())
		buf, _ := httputil.DumpRequest(r, false)
		log.Println("proxy:", string(buf))
		d(r)
	}

	http.Handle("/stv/", proxy)

	log.Println("serving: http://localhost" + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
