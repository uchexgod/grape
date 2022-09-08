package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type cmdToExec struct {
	cmd []string
}

func (c *cmdToExec) exec(ctx context.Context) {
	cwd, _ := os.Getwd()

	fmt.Printf("working dir %s \n", cwd)

	cmd := exec.CommandContext(ctx, c.cmd[0], c.cmd[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

func run(config *Config, namespace string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	ns := config.getNameSpace(namespace)

	cmd := cmdToExec{
		cmd: strings.Split(ns.Run, " "),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Start listening for events.
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op == fsnotify.Write || event.Op == fsnotify.Create {
					cmd.exec(ctx)
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = transverse(ns, watcher.Add)
	if err != nil {
		log.Fatal(errors.New("failed to add path"))
	}

	<-make(chan struct{})
	return nil
}

func transverse(ns *Namespace, fn func(string) error) error {

	paths_to_watch := make(chan []string, 1)
	errCh := make(chan error, 1)

	go func(w *FWatcher) {
		for _, path := range w.Include {
			matches, err := filepath.Glob(path)
			if err != nil {
				log.Fatal(err)
			}
			paths_to_watch <- matches
		}
		close(paths_to_watch)
	}(&ns.Watch)

	for paths := range paths_to_watch {
		for _, path := range paths {
			if err := fn(path); err != nil {
				errCh <- err
			}
		}
	}

	return <-errCh
}
