package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type videoPlayer struct {
	pid      int      // process id
	cmd      string   // depending on platform can be mplayer, omxplayer or vlc
	args     []string // default player startup parameters
	fifoName string   // mplayer control pipe
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
		p.fifoName = "/tmp/mp_fifo"
		p.cmd = "mplayer"
		p.args = []string{"-geometry", "480x240+1920+0", "-input", "file=" + p.fifoName}
	}

	var err error

	if len(p.fifoName) > 0 {
		cmd := exec.Command("mkfifo", p.fifoName)
		log.Println("create fifo")
		if err = cmd.Run(); err != nil {
			log.Println("mkfifo:", err.Error())
		}
		log.Println("fifo created")
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

func (p *videoPlayer) command(s string) {
	f, err := os.OpenFile(p.fifoName, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	log.Println("command:", s)
	if _, err := fmt.Fprintln(f, s); err != nil {
		log.Println("command error:", err.Error())
	}
	f.Close()
}
