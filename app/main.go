package main

import (
	"docker-compose-manage/m/app/config"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
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
}

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
}

var revision string = "development"

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

const (
	focusColor = "#2EF8BB"
	breakColor = "#FF5F87"
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	form        *huh.Form
	list        list.Model
	choice      string
	choiceIndex int
	quitting    bool
	composes    []DockerCompose
	progress    progress.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				m.choiceIndex = m.list.Index()
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {

	if m.choice != "" {
		var status string = "stopped"
		for _, compose := range m.composes {
			if compose.Index == m.choiceIndex {
				if compose.Status == "stopped" {
					fmt.Printf("Starting %s\n", compose.Config.Name)
					compose.Start()
					status = "started"
				} else {
					fmt.Printf("Stopping %s\n", compose.Config.Name)
					compose.Stop()
				}
			}
		}

		return quitTextStyle.Render(fmt.Sprintf("Containers was %s ...", status))
	}
	if m.quitting {
		return quitTextStyle.Render("quitting ...")
	}
	return "\n" + m.list.View()
}

func NewModel() model {
	theme := huh.ThemeCharm()
	theme.Focused.Base.Border(lipgloss.HiddenBorder())
	theme.Focused.Title.Foreground(lipgloss.Color(focusColor))
	theme.Focused.SelectSelector.Foreground(lipgloss.Color(focusColor))
	theme.Focused.SelectedOption.Foreground(lipgloss.Color("15"))
	theme.Focused.Option.Foreground(lipgloss.Color("7"))

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[time.Duration]().
				Title("Focus Time").
				Key("focus").
				Options(
					huh.NewOption("25 minutes", 25*time.Minute),
					huh.NewOption("30 minutes", 30*time.Minute),
					huh.NewOption("45 minutes", 45*time.Minute),
					huh.NewOption("1 hour", time.Hour),
				),
		),
		huh.NewGroup(
			huh.NewSelect[time.Duration]().
				Title("Break Time").
				Key("break").
				Options(
					huh.NewOption("5 minutes", 5*time.Minute),
					huh.NewOption("10 minutes", 10*time.Minute),
					huh.NewOption("15 minutes", 15*time.Minute),
					huh.NewOption("20 minutes", 20*time.Minute),
				),
		),
	).WithShowHelp(false).WithTheme(theme)

	progress := progress.New()
	progress.FullColor = focusColor
	progress.SetSpringOptions(1, 1)

	return model{
		form:     form,
		progress: progress,
	}
}

func main() {
	//log.Printf("[INFO] Docker compose manager: %s\n", revision)

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

	composes, _ := loadComposes(cnf)

	items := []list.Item{}

	for _, compose := range composes {
		msg := fmt.Sprintf("%d: %s - %s", compose.Index, compose.Config.Name, compose.Status)
		items = append(items, item(msg))
	}

	const defaultWidth = 20

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Choise a compose to start/stop:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := NewModel()
	m.list = l
	m.composes = composes
	//m := model{list: l, composes: composes}

	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Println("could not start program:", err)
		os.Exit(1)
	}
}

func loadComposes(cnf config.Config) ([]DockerCompose, error) {
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
		os.Chdir(compose.Path)
		output, err := exec.Command("docker-compose", "top").Output()
		if err != nil {
			composes[index].Status = "error"
			continue
		}
		if len(output) > 0 {
			composes[index].Status = "running"
		}
	}

	return composes, nil
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
