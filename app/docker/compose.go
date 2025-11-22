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

// parseCommand splits a command string into arguments, respecting quoted strings
func parseCommand(cmd string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	escapeNext := false

	for _, r := range cmd {
		if escapeNext {
			current.WriteRune(r)
			escapeNext = false
			continue
		}

		switch r {
		case '\\':
			escapeNext = true
		case '"':
			inQuotes = !inQuotes
		case ' ', '\t':
			if inQuotes {
				current.WriteRune(r)
			} else if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
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
		status, err := compose.GetActualStatus()
		if err != nil {
			log.Printf("[WARN] Failed to get status for %s: %v", compose.Config.Name, err)
			status = "unknown"
		}
		composes[index].Status = status
	}

	return composes, nil
}

func (d DockerCompose) String() string {
	return fmt.Sprintf("Path: %s, Status: %s", d.Path, d.Status)
}

func (d DockerCompose) executeCommand(action string, defaultCmd []string, customCmd string) ([]byte, error) {
	log.Printf("%s %s", action, d.Config.Name)
	commands := defaultCmd

	if customCmd != "" {
		commands = parseCommand(customCmd)
	}

	if len(commands) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	cmd := exec.Command(commands[0], commands[1:]...)
	cmd.Dir = d.Path
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("[ERROR] Failed to %s %s: %v, output: %s", action, d.Config.Name, err, string(output))
		return output, err
	}

	return output, nil
}

func (d DockerCompose) Start() ([]byte, error) {
	return d.executeCommand("Starting", []string{"docker-compose", "up", "-d"}, d.Config.Commands.Start)
}

func (d DockerCompose) Stop() ([]byte, error) {
	return d.executeCommand("Stopping", []string{"docker-compose", "down"}, d.Config.Commands.Stop)
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

func (d *DockerCompose) SetStatus(status string) {
	d.Status = status
}

func (d DockerCompose) IsStatusRunning() bool {
	return d.Status == "running"
}
