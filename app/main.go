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

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jessevdk/go-flags"
)

const listHeight = 20

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

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
}

type item struct {
	title   string
	status  string
	compose DockerCompose
}

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

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.status }
func (i item) FilterValue() string { return i.title }

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
	list        list.Model
	spinner     spinner.Model
	showSpinner bool
	activeItem  item
	choiceIndex int
	items       []item
}

type processMsg struct{}

func main() {
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

	items := []item{}

	for _, compose := range composes {
		title := fmt.Sprintf("%s (%s)", compose.Config.Name, compose.Status)
		items = append(items, item{title: title, status: compose.Status, compose: compose})
	}

	listItems := make([]list.Item, len(items))
	for i, itm := range items {
		listItems[i] = itm
	}

	m := getModel(listItems, items)

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}

func getModel(listItems []list.Item, items []item) model {
	const defaultWidth = 100

	l := list.New(listItems, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Choise a compose to start/stop:"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := model{
		list:    list.New(listItems, list.NewDefaultDelegate(), defaultWidth, listHeight),
		spinner: spinner.New(),
		items:   items,
	}

	m.list.Title = "Items List"

	return m
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
		status, _ := compose.getActualStatus()
		composes[index].Status = status
	}

	return composes, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter":
			selectedItem, ok := m.list.SelectedItem().(item)
			if ok && !m.showSpinner {
				m.activeItem = selectedItem
				m.showSpinner = true
				m.spinner = spinner.New()
				m.choiceIndex = m.list.Index()
				return m, tea.Batch(m.spinner.Tick, processItem())
			}
		}
	case processMsg:
		// Update the status of the selected item
		for i, itm := range m.items {
			if itm.title == m.activeItem.title {
				m.items[i] = item{title: itm.title, status: "Completed"}
			}
		}

		listItems := make([]list.Item, len(m.items))
		for i, itm := range m.items {
			listItems[i] = itm
		}
		m.list.SetItems(listItems)

		m.showSpinner = false
		m.activeItem = item{}
		return m, nil
	}

	// If showing spinner, update spinner only
	if m.showSpinner {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Default list update
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.showSpinner {
		var status string = "stopped"
		for _, item := range m.items {
			if item.compose.Index == m.choiceIndex {
				if item.compose.Status == "stopped" {
					_, err := item.compose.Start()
					if err != nil {
						log.Printf("[ERROR] %v", err)
					}
					item.compose.Status = "started"
					status = "started"
				} else {
					_, err := item.compose.Stop()
					if err != nil {
						log.Printf("[ERROR] %v", err)
					}
					item.compose.Status = "stopped"
				}
			}
		}

		return fmt.Sprintf("Processing %s to status %s ... \n\n%s", m.activeItem.title, status, m.spinner.View())
	}
	return m.list.View()
}

func processItem() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return processMsg{}
	})
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

func (d *DockerCompose) getActualStatus() (string, error) {
	os.Chdir(d.Path)
	output, err := exec.Command("docker-compose", "top").Output()
	if err != nil {
		return "", err
	}
	if len(output) > 0 {
		d.Status = "running"
		return "running", nil
	}
	d.Status = "stopped"
	return "stopped", nil
}
