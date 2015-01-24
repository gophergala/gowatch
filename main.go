package main

import (
	"github.com/codegangsta/cli"
	"os"

	"fmt"
	"github.com/howeyc/fsnotify"
	"os/exec"
	"strings"
)

type Reloader struct {
	// The path of the project directory. Defaults to"./" current working directory
	ProjectDir string
	//
	RunCmd string
	//
	ReloadCmd string
	//
	Pid int
}

func NewReloader() *Reloader {
	return &Reloader{
		ProjectDir: "./",
		RunCmd:     "echo alo",
		ReloadCmd:  "",
		Pid:        0,
	}
}

func (self *Reloader) Bump() {
	fmt.Println("Bumping", self)
	if self.Pid != 0 {
		process, err := os.FindProcess(self.Pid)
		if err != nil {
			fmt.Println("Process already died")
		} else {
			fmt.Println("Killing", process.Pid)
			process.Kill()
		}
	}

	command := exec.Command(strings.Fields(self.RunCmd)[0], strings.Fields(self.RunCmd)[1:]...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	err := command.Start()
	if err != nil {
		fmt.Println(err)
	}
	self.Pid = command.Process.Pid
}

func (self *Reloader) Run() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(err)
	}

	done := make(chan bool)
	bump := make(chan bool, 2)

	// Process events
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				fmt.Println("event:", ev.Name)
				bump <- true
			case err := <-watcher.Error:
				fmt.Println("error:", err)
			case <-bump:
				go self.Bump()
			}
		}
	}()
	bump <- true

	err = watcher.Watch(self.ProjectDir)
	if err != nil {
		fmt.Println(err)
	}

	<-done

	/* ... do stuff ... */
	watcher.Close()
}

func main() {
	app := cli.NewApp()
	app.Name = "gowatch"
	app.EnableBashCompletion = true
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:      "path",
			ShortName: "p",
			Usage:     "project path",
			Action: func(c *cli.Context) {
				println("completed task: ", c.Args().First())
			},
		},
	}
	app.Run(os.Args)

	reloader := NewReloader()
	reloader.Run()
}
