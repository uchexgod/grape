//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
)

func run(ns *Namespace) *exec.Cmd {
	chunks := strings.Fields(ns.Run)
	if len(chunks) == 0 {
		log.Println("No command provided")
		return nil
	}

	cmd := exec.Command(chunks[0], chunks[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		log.Printf("Error starting the command: %v\n", err)
		return nil
	}
	fmt.Println(infoText(RunNotice))
	return cmd
}

func kill(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		pgid, err := syscall.Getpgid(cmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		}
		cmd.Wait()
	}
}

func Run(config *Config, namespace string) error {
	quit := make(chan os.Signal, 1)
	exit := make(chan struct{}, 1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	ns := config.GetNameSpace(namespace)
	cmd := run(ns)

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					fmt.Println(delText(event.Name))

					// Display the contents of the changed file
					// if content, err := ioutil.ReadFile(event.Name); err == nil {
					// 	fmt.Printf("Content of %s:\n%s\n", event.Name, content)
					// }

					kill(cmd)
					cmd = run(ns)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-quit
		fmt.Println(stopText())
		kill(cmd)
		exit <- struct{}{}
	}()

	for _, targets := range ns.Watch.Include {
		go walk(targets, watcher.Add, ns.Watch.Exclude)
	}
	<-exit
	return nil
}

func walk(watchTarget string, fn func(string) error, ignore []string) {
	pathsToWatch, err := fs.Glob(os.DirFS("."), watchTarget)
	if err != nil && err != fs.ErrNotExist {
		log.Fatal(err.Error())
	}

	for _, path := range pathsToWatch {
		if err := fn(path); err != nil {
			log.Fatal(err.Error())
		}
	}
}
