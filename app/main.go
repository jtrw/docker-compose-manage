package main

import (
	"docker-compose-manage/m/app/config"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
)

type Commands struct {
	Stop  string
	Start string
}

type DockerCompose struct {
	Index    int
	Path     string
	Status   string
	Commands Commands
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
		dc := DockerCompose{
			Index:  i,
			Path:   row.Path,
			Status: "stopped",
			Commands: Commands{
				Start: row.Commands.Start,
				Stop:  row.Commands.Stop,
			},
		}
		composes = append(composes, dc)
	}

	for index, compose := range composes {
		os.Chdir(compose.Path)
		output, err := exec.Command("docker-compose", "top").Output()
		if err != nil {
			fmt.Println("Docker compose ps failed")
			return
		}
		if len(output) > 0 {
			composes[index].Status = "running"
		}
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

	for _, compose := range composes {
		if compose.Index == index {
			if composes[index].Status == "stopped" {
				composes[index].Start()
			} else {
				composes[index].Stop()
			}
		}
	}
}

func (d DockerCompose) String() string {
	return fmt.Sprintf("Path: %s, Status: %s", d.Path, d.Status)
}

func (d DockerCompose) Start() {
	os.Chdir(d.Path)

	commands := []string{"docker-compose", "up", "-d"}

	if d.Commands.Start != "" {
		commands = strings.Split(d.Commands.Start, " ")
	}
	fmt.Println(commands)

	output, err := exec.Command(commands[0], commands[1:]...).Output()
	if err != nil {
		fmt.Println("Docker compose up failed")
	}
	fmt.Println(string(output))
}

func (d DockerCompose) Stop() []byte {
	os.Chdir(d.Path)
	commands := []string{"docker-compose", "down"}

	if d.Commands.Stop != "" {
		commands = strings.Split(d.Commands.Stop, " ")
	}
	fmt.Println(commands)
	output, err := exec.Command(commands[0], commands[1:]...).Output()
	if err != nil {
		fmt.Println("Docker compose down failed")
	}
	fmt.Println(string(output))
	return output
}
