package main

import (
	"docker-compose-manage/m/app/config"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/jessevdk/go-flags"
)

type DockerCompose struct {
	Path   string
	Status string
}

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
}

var revision string

func main() {
	log.Printf("[INFO] Micro tracker praser: %s\n", revision)

	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		os.Exit(1)
	}

	cnf, err := config.LoadConfig(opts.Config)
	if err != nil {
		log.Printf("[FATAL] %v", err)
		os.Exit(1)
	}

	composes := []DockerCompose{}

	for _, row := range cnf.Projects {
		composes = append(composes, DockerCompose{Path: row.Path, Status: "stopped"})
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
