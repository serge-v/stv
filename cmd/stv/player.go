package main

import (
	"log"
	"os"
	"os/exec"
	"time"
)

type videoPlayer struct {
	pid  int      // process id
	cmd  string   // depending on platform can be mplayer, omxplayer or vlc
	args []string // default player startup parameters
}

func newPlayer() *videoPlayer {
	p := &videoPlayer{}

	// TODO: make selection based on platform
	user := os.Getenv("USER")
	if user == "pi" || user == "alarm" {
		p.cmd = "omxplayer"
		p.args = []string{}
	} else if user == "odroid" {
		p.cmd = "vlc"
		p.args = []string{}
	} else {
		p.cmd = "mplayer"
		p.args = []string{"-geometry", "480x240+1920+0"}
	}
	return p
}

func (p *videoPlayer) stop() error {
	log.Println("player pid:", p.pid)
	if p.pid > 0 {
		cmd := exec.Command("pkill", p.cmd)
		err := cmd.Run()
		time.Sleep(time.Second)
		log.Println("kill signal sent")
		return err
	}
	return nil
}

func (p *videoPlayer) start(href string) error {
	err := p.stop()
	if err != nil {
		return err
	}

	streamURL := href
	args := append([]string{}, p.args...)
	args = append(args, streamURL)
	cmd := exec.Command(p.cmd, args...)
	log.Printf("%+v\n", cmd.Args)

	if err = cmd.Start(); err != nil {
		log.Println(err)
	}
	p.pid = cmd.Process.Pid

	go func() {
		err = cmd.Wait()
		if err != nil {
			println(err.Error())
		}
		p.pid = 0
		log.Println("player stopped")
	}()

	return err
}
