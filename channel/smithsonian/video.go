package smithsonian

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

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

func cached(fname string) bool {
	_, err := os.Stat(fname)
	if err == nil {
		return true
	}
	return false
}

func download(fname, srcurl string) error {
	f, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer f.Close()

	resp, err := http.Get(srcurl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return err
	}
	log.Println("fname:", fname, "written:", written)

	return nil
}

func getVideoURL(fname string) (string, error) {
	f, err := os.Open(fname)
	if err != nil {
		return "", err
	}

	defer f.Close()
	scanner := bufio.NewScanner(f)
	found := false
	vidurl := ""
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF") && strings.HasSuffix(line, ",RESOLUTION=640x360") {
			found = true
			break
		}
	}
	if found {
		if scanner.Scan() {
			vidurl = scanner.Text()
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	if vidurl == "" {
		return "", fmt.Errorf("cannot get videourl from %s", fname)
	}
	return vidurl, nil
}

func saveSegments(dstfname, fname string) (int64, error) {
	out, err := os.Create(dstfname)
	if err != nil {
		return 0, err
	}

	f, err := os.Open(fname)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	var totalWritten int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "http") {
			continue
		}
		resp, err := http.Get(line)
		if err != nil {
			return 0, err
		}
		written, err := io.Copy(out, resp.Body)
		if err != nil {
			return 0, err
		}
		totalWritten += written
		if totalWritten%(1024*1024) == 0 {
			log.Println(totalWritten/1024/1024, "Mib")
		}
		resp.Body.Close()
	}
	out.Close()
	return totalWritten, nil
}

func CacheVideo(cacheDir string) error {
	url := "http://www.smithsonianchannel.com/full-episodes"
	log.Println("loading:", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	s := string(buf)

	chunks := rexLI.FindAllStringIndex(s, 20)
	saved := 0

	for _, c := range chunks {
		ss := s[c[0]:c[1]]
		href := rexHref.FindStringSubmatch(ss)[1]
		title := rexTitle.FindStringSubmatch(ss)[1]
		id, stream, err := getStreamURL(href)
		fname := fmt.Sprintf("%s/sms-a-%d.txt", cacheDir, id)
		log.Println("id:", id, "title:", title)
		if err != nil {
			return err
		}
		if !cached(fname) {
			if err := download(fname, stream); err != nil {
				return err
			}
		}

		videoURL, err := getVideoURL(fname)
		if err != nil {
			return err
		}
		log.Println("id:", id, "videoURL:", videoURL)
		videoFname := fmt.Sprintf("%s/sms-b-%d.txt", cacheDir, id)
		if !cached(videoFname) {
			if err := download(videoFname, videoURL); err != nil {
				return err
			}
		}

		dstfname := fmt.Sprintf("%s/sms-c-%d.mp4", cacheDir, id)
		if saved == 0 && !cached(dstfname) {
			if _, err = saveSegments(dstfname, videoFname); err != nil {
				return err
			}
			saved++
		}
	}

	return nil
}
