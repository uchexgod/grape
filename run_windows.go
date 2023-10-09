//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

func Run(config *Config, namespace string) error {
	// For now, we'll use a placeholder. We need to decide whether
	// the functionality for watching files and running commands on change
	// should differ for Windows vs. other platforms.
	fmt.Println("Windows-specific run not implemented yet!")
	return nil
}

func run(ns *Namespace) *exec.Cmd {
	chunks := strings.Fields(ns.Run)
	if len(chunks) == 0 {
		return nil
	}

	cmd := exec.Command(chunks[0], chunks[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting the command: %v\n", err)
		return nil
	}
	fmt.Println(infoText(RunNotice))
	return cmd
}

func kill(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}

	pid := cmd.Process.Pid
	handle, err := syscall.OpenProcess(syscall.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening process: %v\n", err)
		return
	}
	defer syscall.CloseHandle(handle)

	if err := syscall.TerminateProcess(handle, 0); err != nil {
		fmt.Fprintf(os.Stderr, "Error terminating the process: %v\n", err)
	}
	cmd.Wait()
}
