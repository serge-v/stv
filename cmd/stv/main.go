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
	"strings"

	"github.com/godbus/dbus"
	"github.com/serge-v/stv/channel"
)

var (
	debug      = flag.Bool("d", false, "debug in console")
	cacheDir   = os.Getenv("HOME") + "/.cache/stv"
	configDir  = os.Getenv("HOME") + "/.config/stv"
	token      string
	tokenFname string
	baseURL    string
	confURL    string
	listURL    string
	addr       = ":6061"
	player     *videoPlayer
	dbusc      *dbus.Conn
	transport  *http.Transport
)

func init() {
	tokenFname = configDir + "/token.txt"
	player = newPlayer()
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		return
	}

	player.stop()

	var d mainData
	d.List, d.Error = getLocalVideos()

	var rd mainData
	if err := unmarshalURL(listURL, &rd); err != nil {
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
	if err := unmarshalURL(listURL, &resp); err != nil {
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
		d.Error = player.start(link)
	}

	seek := r.URL.Query().Get("seek")
	if seek != "" {
		//		d, _ := time.ParseDuration(seek)
		//		seekPlayer(d)
	}

	vol := r.URL.Query().Get("vol")
	if vol != "" {
		//		n, _ := strconv.Atoi(vol)
		//		changeVolume(n)
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
	client := &http.Client{Transport: transport}
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
	client := &http.Client{Transport: transport}
	resp, err := client.Get(confURL)
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
	client := &http.Client{Transport: transport}
	resp, err := client.Post(confURL, "application/json", &b)
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

func createProxy() *httputil.ReverseProxy {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Transport = transport
	d := proxy.Director
	proxy.Director = func(r *http.Request) {
		r.Host = u.Host
		if *debug {
			log.Println("url1:", r.URL.String())
		}
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/stv")
		if *debug {
			log.Println("url2:", r.URL.String())
			buf, _ := httputil.DumpRequest(r, false)
			log.Println("proxy:", string(buf))
		}
		d(r)
	}
	return proxy
}

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

	if *debug {
		baseURL = "https://conf.svtest.com:9001/"
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	} else {
		baseURL = "https://conf.voilokov.com/"
		transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: false}}
	}
	confURL = baseURL + token + "/stv/config"
	listURL = baseURL + token + "/stv/list.json"
	loadState()

	//	dbusc, err = dbus.SessionBus()
	//	if err != nil {
	//		panic(err)
	//	}

	log.Println("player:", player.cmd, "token:", token)
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		log.Println("addr: ", addr.String())
	}

	http.HandleFunc("/shutdown", shutdownHandler)
	http.HandleFunc("/restart", restartHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/", mainHandler)

	http.Handle("/stv/", createProxy())

	log.Println("serving: http://localhost" + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
