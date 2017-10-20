package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"time"
)

type videoPlayer struct {
	pid      int            // process id
	cmd      string         // depending on platform can be mplayer, omxplayer or vlc
	args     []string       // default player startup parameters
	fifoName string         // mplayer control pipe
	stdin    io.WriteCloser // stdin pipe for controlling omxplayer
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
	if p.cmd == "omxplayer" {
		p.stdin, err = cmd.StdinPipe()
		if err != nil {
			log.Println(err)
			return err
		}
	}
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

func (p *videoPlayer) seek(d int) {
	switch p.cmd {
	case "omxplayer":
		switch {
		case d > 300:
			p.sendStdinCommand("^[[A")
		case d < -300:
			p.sendStdinCommand("^[[B")
		case d > 0 && d < 300:
			p.sendStdinCommand("^[[C")
		case d < 0 && d > -300:
			p.sendStdinCommand("^[[D")
		}
	case "mplayer":
		p.sendPipeCommand(fmt.Sprintf("seek %d", d))
	}

}

func (p *videoPlayer) volume(d int) {
	switch p.cmd {
	case "omxplayer":
		switch {
		case d > 0:
			p.sendStdinCommand("+")
		case d < 0:
			p.sendStdinCommand("-")
		}
	case "mplayer":
		p.sendPipeCommand(fmt.Sprintf("volume %d", d))
	}
}

func (p *videoPlayer) pause() {
	switch p.cmd {
	case "omxplayer":
		p.sendStdinCommand("p")
	case "mplayer":
		p.sendPipeCommand("pause")
	}
}

func (p *videoPlayer) sendPipeCommand(s string) {
	f, err := os.OpenFile(p.fifoName, os.O_WRONLY, 0)
	if err != nil {
		panic(err)
	}
	log.Println("pipe command:", s)
	if _, err := fmt.Fprintln(f, s); err != nil {
		log.Println("pipe command error:", err.Error())
	}
	f.Close()
}

func (p *videoPlayer) sendStdinCommand(s string) {
	log.Println("stdin command:", s)
	if _, err := fmt.Fprint(p.stdin, s); err != nil {
		log.Println("stdin command error:", err.Error())
	}
}
