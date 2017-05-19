package main

import (
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

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
    font-size: 22px;
    width: 240px;
    height: 100px;
}
</style>
</head>
<body>
	<table border="0">
		{{range .}}
		<tr><td><a class="button" href="/play?href={{.Href}}">{{.Name}}</a></td><tr>
		{{end}}
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
    font-size: 22px;
    width: 240px;
    height: 100px;
}
</style>
</head>
<body>
	<table>
		<tr><td><a href="/" class="button">Stop</a></td><tr>
	</table>
</body>
</html>
`

var mt = template.Must(template.New("main").Parse(mainPageHTML))
var pt = template.Must(template.New("play").Parse(playPageHTML))

type item struct {
	Name string
	Href string
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		return
	}

	stopPlayer()
	list, err := getVideos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := mt.Execute(w, list); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func playHandler(w http.ResponseWriter, r *http.Request) {
	href := r.URL.Query().Get("href")
	fmt.Println("starting player")
	playVideo(href)
	if err := pt.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func readInt(name string) int {
	fname := fmt.Sprintf("cfg-%s.txt", name)
	f, err := os.Open(fname)
	if err != nil {
		return 0
	}
	var n int
	fmt.Fscan(f, &n)
	f.Close()
	n -= 10
	if n < 0 {
		n = 0
	}
	return n
}

func saveInt(name string, n int) {
	fname := fmt.Sprintf("cfg-%s.txt", name)
	f, err := os.Create(fname)
	if err != nil {
		return
	}
	fmt.Fprintf(f, "%d", n)
	f.Close()
}

var (
	rexLI    = regexp.MustCompile("(?sU)data-premium.*</li>")
	rexHref  = regexp.MustCompile("(?sU)href=\"([^\"]+)\"")
	rexTitle = regexp.MustCompile("(?sU)<h2 class=\"promo-show-name\">(.*)</h2>")
	rexBcid  = regexp.MustCompile("data-bcid=\"([^\"]+)\"")
)

func stopPlayer() {
	pid := readInt("player-pid")
	println("player pid:", pid)
	if pid > 0 {
		cmd := exec.Command("pkill", "omxplayer")
		cmd.Run()
		time.Sleep(time.Second)
		println("mplayer killed")
	}
}

func playVideo(href string) {
	stopPlayer()

	resp, err := http.Get("http://www.smithsonianchannel.com" + href)
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	m := rexBcid.FindStringSubmatch(string(buf))
	bcid := m[1]

	n := readInt(bcid)
	start := time.Now()
	streamURL := fmt.Sprintf("http://c.brightcove.com/services/mobile/streaming/index/master.m3u8?videoId=%s&pubId=1466806621001", bcid)

	//	cmd := exec.Command("mplayer", "-ss", strconv.Itoa(n), streamURL)
	cmd := exec.Command("omxplayer", "--live", streamURL)

	fmt.Printf("%+v\n", cmd.Args)

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	saveInt("player-pid", cmd.Process.Pid)

	go func() {
		err := cmd.Wait()
		if err != nil {
			println(err.Error())
		}
		println("player stopped")
		saveInt("player-pid", 0)
		elapsed := time.Since(start)
		if n+int(elapsed.Seconds()) > 20 {
			saveInt(bcid, int(elapsed.Seconds())+n)
		}
	}()
}

func getVideos() ([]item, error) {

	resp, err := http.Get("http://www.smithsonianchannel.com/full-episodes")
	if err != nil {
		panic(err)
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
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
	return list, nil
}

func printList(n int) {
	list, err := getVideos()
	if err != nil {
		panic(err)
	}

	if n < 0 || n > len(list) {
		panic("wrong movie index")
	}

	if n > 0 {
		playVideo(list[n-1].Href)
		return
	}

	for idx, item := range list {
		fmt.Println(idx+1, item.Name)
	}
}

var debug = flag.Bool("d", false, "debug in console")
var playNum = flag.Int("p", 0, "play episode `NUM`")

func main() {
	flag.Parse()
	if *debug {
		printList(*playNum)
		return
	}

	http.HandleFunc("/play", playHandler)
	http.HandleFunc("/", mainHandler)
	if err := http.ListenAndServe(":6061", nil); err != nil {
		panic(err)
	}
}
