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
	Index  int
	Path   string
	Status string
}

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
}

var revision string = "development"

func main() {
	log.Printf("[INFO] Docker compose manager: %s\n", revision)

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

	for i, row := range cnf.Projects {
		composes = append(composes, DockerCompose{Index: i, Path: row.Path, Status: "stopped"})
	}

	for index, compose := range composes {
		os.Chdir(compose.Path)
		_, err := exec.Command("docker-compose", "ps").Output()
		if err != nil {
			fmt.Println("Docker compose ps failed")
			return
		}
		composes[index].Status = "running"
	}

	for _, compose := range composes {
		fmt.Printf("%d: %s Path: %s \n", compose.Index, compose.Status, compose.Path)
	}

	var index int

	fmt.Printf("Enter index docker ...\n")
	_, err = fmt.Scanln(&index)

	if err != nil {
		panic(err)
	}

	output := composes[index].Stop()
	fmt.Println(string(output))
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

func (d DockerCompose) Stop() []byte {
	output, err := exec.Command("docker-compose", "down").Output()
	if err != nil {
		fmt.Println("Docker compose down failed")
	}
	return output
}
