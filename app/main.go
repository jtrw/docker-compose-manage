package main

import (
	"fmt"
	"os"
	"os/exec"
)

type DockerCompose struct {
	Path   string
	Status string
}

func (d DockerCompose) String() string {
	return fmt.Sprintf("Path: %s, Status: %s", d.Path, d.Status)
}

func (d DockerCompose) Start() {
	_, err := exec.Command("docker-compose", "up", "-d").Output()
	if err != nil {
		fmt.Println("Docker compose up failed")
	}
}

func (d DockerCompose) Stop() {
	_, err := exec.Command("docker-compose", "down").Output()
	if err != nil {
		fmt.Println("Docker compose down failed")
	}
}

func main() {
	dockerPaths := []string{
		"/path/to/docker-compose-1",
	}

	composes := []DockerCompose{}

	for _, path := range dockerPaths {
		composes = append(composes, DockerCompose{Path: path, Status: "stopped"})
	}

	for index, compose := range composes {
		// Run docker-compose ps
		os.Chdir(compose.Path)
		_, err := exec.Command("docker-compose", "ps").Output()
		if err != nil {
			fmt.Println("Docker compose ps failed")
			return
		}
		composes[index].Status = "running"

		//fmt.Println(string(output))
	}

	fmt.Println(composes)
}
