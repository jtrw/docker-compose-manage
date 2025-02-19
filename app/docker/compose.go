package docker

import (
	"docker-compose-manage/m/app/config"
	"fmt"
	"log"
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

func LoadComposes(cnf config.Config) ([]DockerCompose, error) {
	composes := []DockerCompose{}
	index := 0
	for _, row := range cnf.Projects {
		dc := DockerCompose{
			Index:  index,
			Path:   row.Path,
			Status: "stopped",
			Config: row,
		}
		composes = append(composes, dc)
		index++
	}

	for index, compose := range composes {
		status, _ := compose.GetActualStatus()
		composes[index].Status = status
	}

	return composes, nil
}

func (d DockerCompose) String() string {
	return fmt.Sprintf("Path: %s, Status: %s", d.Path, d.Status)
}

func (d DockerCompose) Start() ([]byte, error) {
	log.Printf("Starting log %s", d.Config.Name)
	commands := []string{"docker-compose", "up", "-d"}

	if d.Config.Commands.Start != "" {
		commands = strings.Split(d.Config.Commands.Start, " ")
	}
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Dir = d.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (d DockerCompose) Stop() ([]byte, error) {
	log.Printf("Stopping log %s", d.Config.Name)
	commands := []string{"docker-compose", "down"}

	if d.Config.Commands.Stop != "" {
		commands = strings.Split(d.Config.Commands.Stop, " ")
	}
	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Dir = d.Path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	return output, nil
}

func (d DockerCompose) GetActualStatus() (string, error) {
	cmd := exec.Command("docker-compose", "top")
	cmd.Dir = d.Path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	if len(output) > 0 {
		return "running", nil
	}

	return "stopped", nil
}

func (d DockerCompose) StartAsync(ch chan string) {
	_, err := d.Start()
	if err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
		return
	}
	ch <- "running"
}

func (d DockerCompose) StopAsync(ch chan string) {
	_, err := d.Stop()
	if err != nil {
		ch <- fmt.Sprintf("Error: %v", err)
		return
	}
	ch <- "stopped"
}

func (d DockerCompose) IsStatusStopped() bool {
	return d.Status == "stopped"
}

func (d DockerCompose) SetStatus(status string) {
	d.Status = status
}

func (d DockerCompose) IsStatusRunning() bool {
	return d.Status == "running"
}
