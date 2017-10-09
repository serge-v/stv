package smithsonian

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
)

type Item struct {
	ID   int
	Name string
	Href string
}

var (
	rexLI    = regexp.MustCompile("(?sU)data-premium.*</li>")
	rexHref  = regexp.MustCompile("(?sU)href=\"([^\"]+)\"")
	rexTitle = regexp.MustCompile("(?sU)<h2 class=\"promo-show-name\">(.*)</h2>")
	rexBcid  = regexp.MustCompile("data-bcid=\"([^\"]+)\"")
)

func getStreamURL(href string) (int, string, error) {
	resp, err := http.Get("http://www.smithsonianchannel.com" + href)
	if err != nil {
		return 0, "", err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return 0, "", err
	}
	m := rexBcid.FindStringSubmatch(string(buf))
	bcid := m[1]
	id, err := strconv.Atoi(bcid)
	if err != nil {
		return 0, "", err
	}
	url := fmt.Sprintf("http://c.brightcove.com/services/mobile/streaming/index/master.m3u8?videoId=%s&pubId=1466806621001", bcid)
	return id, url, nil
}

func cached(id int) bool {
	fname := fmt.Sprintf("smithsonian-%d.mp4", id)
	_, err := os.Stat(fname)
	if err == nil {
		return true
	}
	return false
}

func download(id int, url string) error {
	fname := fmt.Sprintf("smithsonian-%d.mp4", id)
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	log.Println("id:", id, "written:", written)

	return nil
}

func GetVideos() ([]Item, error) {
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
	list := make([]Item, 0, len(chunks))

	for _, c := range chunks {
		ss := s[c[0]:c[1]]
		href := rexHref.FindStringSubmatch(ss)[1]
		title := rexTitle.FindStringSubmatch(ss)[1]
		id, stream, err := getStreamURL(href)
		if err != nil {
			return nil, err
		}
		if !cached(id) {
			if err := download(id, stream); err != nil {
				return nil, err
			}
		}

		it := Item{
			ID:   id,
			Name: title,
			Href: href,
		}
		list = append(list, it)
	}

	return list, nil
}
