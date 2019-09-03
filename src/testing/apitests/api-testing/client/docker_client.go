package client

import "os/exec"
import "strings"
import "errors"
import "bufio"
import "fmt"

// DockerClient : Run docker commands
type DockerClient struct{}

// Status : Check if docker daemon is there
func (dc *DockerClient) Status() error {
	cmdName := "docker"
	args := []string{"info"}

	return dc.runCommand(cmdName, args)
}

// Pull : Pull image
func (dc *DockerClient) Pull(image string) error {
	if len(strings.TrimSpace(image)) == 0 {
		return errors.New("Empty image")
	}

	cmdName := "docker"
	args := []string{"pull", image}

	return dc.runCommandWithOutput(cmdName, args)
}

// Tag :Tag image
func (dc *DockerClient) Tag(source, target string) error {
	if len(strings.TrimSpace(source)) == 0 ||
		len(strings.TrimSpace(target)) == 0 {
		return errors.New("Empty images")
	}

	cmdName := "docker"
	args := []string{"tag", source, target}

	return dc.runCommandWithOutput(cmdName, args)
}

// Push : push image
func (dc *DockerClient) Push(image string) error {
	if len(strings.TrimSpace(image)) == 0 {
		return errors.New("Empty image")
	}

	cmdName := "docker"
	args := []string{"push", image}

	return dc.runCommandWithOutput(cmdName, args)
}

// Login : Login docker
func (dc *DockerClient) Login(userName, password string, uri string) error {
	if len(strings.TrimSpace(userName)) == 0 ||
		len(strings.TrimSpace(password)) == 0 {
		return errors.New("Invlaid credential")
	}

	cmdName := "docker"
	args := []string{"login", "-u", userName, "-p", password, uri}

	return dc.runCommandWithOutput(cmdName, args)
}

func (dc *DockerClient) runCommand(cmdName string, args []string) error {
	return exec.Command(cmdName, args...).Run()
}

func (dc *DockerClient) runCommandWithOutput(cmdName string, args []string) error {
	cmd := exec.Command(cmdName, args...)
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			fmt.Printf("%s out | %s\n", cmdName, scanner.Text())
		}
	}()

	if err = cmd.Start(); err != nil {
		return err
	}

	if err = cmd.Wait(); err != nil {
		return err
	}

	return nil
}
