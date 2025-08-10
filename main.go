package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

type Item struct {
	TitleText       string
	DescriptionText string
	Details         string
}

func (i Item) Title() string       { return i.TitleText }
func (i Item) Description() string { return i.DescriptionText }
func (i Item) FilterValue() string { return i.TitleText }

type Model struct {
	list    list.Model
	items   []Item
	preview string
}

var (
	outerStyle   = lipgloss.NewStyle().Margin(1)                  // outer gap
	listStyle    = lipgloss.NewStyle().Width(40).MarginRight(2)   // left list
	previewStyle = lipgloss.NewStyle().Width(50).PaddingTop(2)    // preview with top padding
)

var sampleYAML = `
- title: Server Alpha
  description: Primary API server
  details: |
    OS: Ubuntu 24.04
    IP: 10.0.0.1
    Status: Online

- title: Server Beta
  description: Backup server
  details: |
    OS: Rocky Linux 9
    IP: 10.0.0.2
    Status: Maintenance
`

func parseYAML(data string) ([]Item, error) {
	var raw []map[string]string
	if err := yaml.Unmarshal([]byte(data), &raw); err != nil {
		return nil, err
	}
	var items []Item
	for _, entry := range raw {
		items = append(items, Item{
			TitleText:       entry["title"],
			DescriptionText: entry["description"],
			Details:         entry["details"],
		})
	}
	return items, nil
}

func initialModel() Model {
	items, err := parseYAML(sampleYAML)
	if err != nil {
		log.Fatal(err)
	}

	var listItems []list.Item
	for _, it := range items {
		listItems = append(listItems, it)
	}

	l := list.New(listItems, list.NewDefaultDelegate(), 40, 20)
	l.Title = "Servers"

	return Model{
		list:    l,
		items:   items,
		preview: items[0].Details,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}

	m.list, cmd = m.list.Update(msg)

	if sel, ok := m.list.SelectedItem().(Item); ok {
		m.preview = sel.Details
	}

	return m, cmd
}

func (m Model) View() string {
	left := listStyle.Render(m.list.View())
	right := previewStyle.Render(m.preview)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	return outerStyle.Render(mainContent) // add outer gap
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

