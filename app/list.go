package main

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type item struct {
	title  string
	status string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.status }
func (i item) FilterValue() string { return i.title }

type model struct {
	list        list.Model
	spinner     spinner.Model
	showSpinner bool
	activeItem  item
	items       []item
}

type processMsg struct{}

func main() {
	items := []item{
		{title: "Task 1", status: "Pending"},
		{title: "Task 2", status: "Pending"},
		{title: "Task 3", status: "Pending"},
	}

	listItems := make([]list.Item, len(items))
	for i, itm := range items {
		listItems[i] = itm
	}

	m := model{
		list:    list.New(listItems, list.NewDefaultDelegate(), 0, 0),
		spinner: spinner.New(),
		items:   items,
	}

	m.list.Title = "Items List"

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
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
		return fmt.Sprintf("Processing %s...\n\n%s", m.activeItem.title, m.spinner.View())
	}
	return m.list.View()
}

func processItem() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return processMsg{}
	})
}
