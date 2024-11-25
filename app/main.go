package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	dockerPaths := []string{
		"/path/to/docker-compose-1",
	}

	for _, path := range dockerPaths {
		// Run docker-compose ps
		os.Chdir(path)
		output, err := exec.Command("docker-compose", "ps").Output()
		if err != nil {
			fmt.Println("Docker compose ps failed")
			return
		}

		fmt.Println(string(output))

	}
}
