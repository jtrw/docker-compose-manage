package main

import (
	"docker-compose-manage/m/app/config"
	"log"
	"os"

	jBubble "docker-compose-manage/m/app/jBubble"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jessevdk/go-flags"
)

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
	Dbg    bool   `long:"dbg" env:"DEBUG" description:"show debug info"`
}

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		log.Printf("[FATAL] %v", err)
		os.Exit(1)
	}

	setupLog(opts.Dbg)

	cnf, err := config.LoadConfig(opts.Config)
	if err != nil {
		log.Printf("[FATAL] %v", err)
		os.Exit(1)
	}

	m := jBubble.GetModel(cnf)

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		log.Printf("[FATAL] Error running program: %v\n", err)
	}
}

func setupLog(dbg bool) {
	if !dbg {
		return
	}
	if _, err := tea.LogToFile("debug.log", "debug"); err != nil {
		panic(err)
	}
	log.Printf("[DEBUG] debug mode ON")
}

func setDebugFile() {
	if _, err := tea.LogToFile("debug.log", "debug"); err != nil {
		panic(err)
	}
}
