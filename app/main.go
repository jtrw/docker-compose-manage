package main

import (
	"docker-compose-manage/m/app/config"
	compose "docker-compose-manage/m/app/docker"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jessevdk/go-flags"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

const (
	focusColor   = "#2EF8BB"
	breakColor   = "#FF5F87"
	listHeight   = 20
	defaultWidth = 100
)

type Options struct {
	Config string `short:"c" long:"config" env:"CONFIG" default:"config.yml" description:"config file"`
}

type item struct {
	title   string
	status  string
	compose compose.DockerCompose
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
	ch          chan string
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

	m := getModel(cnf)

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}

func getListItems(composes []compose.DockerCompose) ([]list.Item, []item) {
	items := []item{}

	for _, compose := range composes {
		title := fmt.Sprintf("%s", compose.Config.Name)
		items = append(items, item{title: title, status: compose.Status, compose: compose})
	}

	listItems := make([]list.Item, len(items))
	for i, itm := range items {
		listItems[i] = itm
	}

	return listItems, items
}

func getModel(cnf config.Config) model {
	composes, _ := compose.LoadComposes(cnf)

	listItems, items := getListItems(composes)

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
		ch:      make(chan string),
	}

	m.list.Title = "Items List"

	return m
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m *model) SetItems(items []list.Item) {
	listItems := make([]list.Item, len(items))
	for i, itm := range items {
		listItems[i] = itm
	}
	m.list.SetItems(listItems)
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
				if selectedItem.compose.Status == "stopped" {
					go selectedItem.compose.StartAsync(m.ch)
					selectedItem.compose.Status = "running"
					//status = "running"
				} else {
					go selectedItem.compose.StopAsync(m.ch)
					selectedItem.compose.Status = "stopped"
					//status = "stopped"
				}

				m.activeItem = selectedItem
				m.showSpinner = true
				m.spinner = spinner.New()
				m.choiceIndex = m.list.Index()
				return m, tea.Batch(m.spinner.Tick, processItem(m.ch))
			}

		}
	case processMsg:
		for i, itm := range m.items {
			if itm.title == m.activeItem.title {
				status := "stopped"
				if itm.compose.Status == "stopped" {
					status = "running"
				}
				m.items[i] = item{title: itm.title, status: status}
			}
		}

		listItems := make([]list.Item, len(m.items))
		for i, itm := range m.items {
			listItems[i] = itm
		}
		m.list.SetItems(listItems)

		m.SetItems(listItems)

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
		var status string
		for _, item := range m.items {
			if item.compose.Index != m.choiceIndex {
				continue
			}

			if item.compose.Status == "stopped" {
				status = "running"
			} else {
				status = "stopped"
			}
		}

		return fmt.Sprintf("Processing %s to status %s ... \n\n%s", m.activeItem.title, status, m.spinner.View())
	}
	return m.list.View()
}

func processItem(ch chan string) tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		//<-ch
		return processMsg{}
	})
}
