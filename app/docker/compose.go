package docker

import (
	"docker-compose-manage/m/app/config"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Commands struct {
	Stop  string
	Start string
}

type DockerCompose struct {
	Index    int
	Path     string
	Status   string
	Config   config.Project
	Commands Commands
	title    string
}

func (d DockerCompose) String() string {
	return fmt.Sprintf("Path: %s, Status: %s", d.Path, d.Status)
}

func (d DockerCompose) Start() ([]byte, error) {
	os.Chdir(d.Path)

	commands := []string{"docker-compose", "up", "-d"}

	if d.Config.Commands.Start != "" {
		commands = strings.Split(d.Config.Commands.Start, " ")
	}

	output, err := exec.Command(commands[0], commands[1:]...).Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (d DockerCompose) Stop() ([]byte, error) {
	os.Chdir(d.Path)
	commands := []string{"docker-compose", "down"}

	if d.Config.Commands.Stop != "" {
		commands = strings.Split(d.Config.Commands.Stop, " ")
	}

	output, err := exec.Command(commands[0], commands[1:]...).Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (d DockerCompose) GetActualStatus() (string, error) {
	os.Chdir(d.Path)
	output, err := exec.Command("docker-compose", "top").Output()
	if err != nil {
		return "", err
	}
	if len(output) > 0 {
		return "running", nil
	}

	return "stopped", nil
}
