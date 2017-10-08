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
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		return
	}

	stopPlayer()

	var d mainData
	d.List, d.Error = getVideos()
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

func playHandler(w http.ResponseWriter, r *http.Request) {
	var d genericData

	href := r.URL.Query().Get("href")
	if href != "" {
		log.Println("starting player")
		d.Error = startPlayer(href)
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

var (
	rexLI    = regexp.MustCompile("(?sU)data-premium.*</li>")
	rexHref  = regexp.MustCompile("(?sU)href=\"([^\"]+)\"")
	rexTitle = regexp.MustCompile("(?sU)<h2 class=\"promo-show-name\">(.*)</h2>")
	rexBcid  = regexp.MustCompile("data-bcid=\"([^\"]+)\"")
)

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

	var streamURL string
	var start time.Time
	var pos int
	var bcid string

	if strings.HasPrefix(href, "vid/") {
		streamURL = href
		bcid = href
	} else {

		resp, err1 := http.Get("http://www.smithsonianchannel.com" + href)
		if err1 != nil {
			return err1
		}
		buf, err1 := ioutil.ReadAll(resp.Body)
		if err1 != nil {
			return err1
		}
		m := rexBcid.FindStringSubmatch(string(buf))
		bcid = m[1]
		start = time.Now()
		pos = state.Elapsed[bcid]
		streamURL = fmt.Sprintf("http://c.brightcove.com/services/mobile/streaming/index/master.m3u8?videoId=%s&pubId=1466806621001", bcid)
	}

	cmd := exec.Command(player, streamURL)
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
		elapsed := time.Since(start)
		if elapsed > 20 {
			state.Elapsed[bcid] = pos + int(elapsed.Seconds()) - 10
		}
		log.Println("player stopped. elapsed: ", elapsed)
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

func getVideos() ([]item, error) {
	resp, err := http.Get("http://www.smithsonianchannel.com/full-episodes")
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	s := string(buf)

	chunks := rexLI.FindAllStringIndex(s, 20)
	list := make([]item, 0, len(chunks))

	for _, c := range chunks {
		ss := s[c[0]:c[1]]
		href := rexHref.FindStringSubmatch(ss)
		title := rexTitle.FindStringSubmatch(ss)
		it := item{
			Name: title[1],
			Href: href[1],
		}
		list = append(list, it)
	}

	local, _ := filepath.Glob("vid/*")
	for _, name := range local {
		it := item{
			Name: filepath.Base(name),
			Href: name,
		}
		list = append(list, it)
	}

	return list, nil
}

func printList(n int) {
	list, err := getVideos()
	if err != nil {
		log.Fatal(err)
	}

	if n < 0 || n > len(list) {
		log.Fatal("wrong movie index")
	}

	if n > 0 {
		startPlayer(list[n-1].Href)
		return
	}

	for idx, item := range list {
		fmt.Println(idx+1, item.Name)
	}
}

var debug = flag.Bool("d", false, "debug in console")
var playNum = flag.Int("p", 0, "play episode `NUM`")
var cacheDir = os.Getenv("HOME") + "/.cache/stv"
var configDir = os.Getenv("HOME") + "/.config/stv"
var player = "mplayer"
var playerArgs = []string{}
var token string
var tokenFname string
var confapi string
var addr = ":6061"

func init() {
	user := os.Getenv("USER")
	if user == "pi" || user == "alarm" {
		player = "omxplayer"
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
	loadState()

	dbusc, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	}

	log.Println("player:", player, "token:", token)
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		log.Println("addr: ", addr.String())
	}

	flag.Parse()
	if *debug {
		printList(*playNum)
		return
	}

	http.HandleFunc("/shutdown", shutdownHandler)
	http.HandleFunc("/restart", restartHandler)
	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/", mainHandler)

	log.Println("serving: http://localhost" + addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
