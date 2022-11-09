package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
)

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runWithOutput(name string, args ...string) (string, error) {
	output := bytes.NewBuffer(nil)

	cmd := exec.Command(name, args...)
	cmd.Stdout = output
	cmd.Stderr = os.Stderr

	return output.String(), cmd.Run()
}

func checkInterface(name string) error {
	output := bytes.NewBuffer(nil)

	cmd := exec.Command("wg", "show", name)
	cmd.Stdout = output
	cmd.Stderr = output

	err := cmd.Run()
	if err != nil {
		log.Println("wireguard interface check failed. output:", output.String())
	}

	return err
}
