package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"gopkg.in/yaml.v3"
)

type User struct {
	Name string `yaml:"name"`
	Mail string `yaml:"mail"`
	Sig  string `yaml:"sig"`
}

type KeyEntry struct {
	Key    string `yaml:"key"`
	Create string `yaml:"create"`
	Expire string `yaml:"expire,omitempty"`
	Algo   string `yaml:"algo"`
	Users  []User `yaml:"users"`
}

type GPGData struct {
	GPG []KeyEntry `yaml:"gpg"`
}

type item struct {
	title       string
	description string
	entry       KeyEntry
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title + " " + i.description }

type model struct {
	list     list.Model
	choice   *KeyEntry
	quitting bool
}

var (
	listStyle    = lipgloss.NewStyle().Width(40).MarginRight(2).Padding(1)
	// previewStyle = lipgloss.NewStyle().Width(60).Padding(1).MarginTop(1)

	// listStyle = lipgloss.NewStyle().
	// 	Width(60).
	// 	Padding(1).
	// 	MarginTop(1).
	// 	Foreground(lipgloss.Color("#00ff00")). // bright green text
	// 	Background(lipgloss.Color("#000000")). // black background
	// 	Bold(true)

	previewStyle = lipgloss.NewStyle().
		Width(60).
		Padding(4).
		MarginLeft(1).
		Foreground(lipgloss.Color("#FAFAFA")).   // light gray text
		// Background(lipgloss.Color("#1E1E1E")).   // dark background
		Bold(false)

	pointStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#aaaaaa")).   // light gray text
		// Background(lipgloss.Color("#1E1E1E")).   // dark background
		Bold(false)

	statusMessageStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).Render
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <yamlfile>")
		os.Exit(1)
	}

	data, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var gpgData GPGData
	if err := yaml.Unmarshal(data, &gpgData); err != nil {
		log.Fatal(err)
	}

	items := []list.Item{}
	for _, k := range gpgData.GPG {
		if len(k.Users) > 0 {
			title := k.Users[0].Name
			desc := k.Users[0].Mail
			items = append(items, item{title, desc, k})
		}
	}

	const defaultHeight = 20
	const defaultWidth = 40
	l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
	l.Title = "Git Config Manager"
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	m := model{list: l}
	if err := tea.NewProgram(m).Start(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = &i.entry
				fmt.Println(pointStyle.Render("Selected Key:"), m.choice.Key)
				fmt.Println("Algorithm:", m.choice.Algo)
				fmt.Println("Created:", m.choice.Create)
				fmt.Println("Expire:", m.choice.Expire)
				for _, u := range m.choice.Users {
					fmt.Printf("User: %s <%s> sig: %s\n", u.Name, u.Mail, u.Sig)
				}
				return m, tea.Quit
			}
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.choice != nil {
		return ""
	}

	left := listStyle.Render(m.list.View())
	right := previewStyle.Render(m.previewView())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) previewView() string {
	if i, ok := m.list.SelectedItem().(item); ok {
		k := i.entry
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("key"), k.Key))
		sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("algorithm"), k.Algo))
		sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("created"), k.Create))
		if k.Expire != "" {
			sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("expires"), k.Expire))
		} else {
			sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("expires"), "Never"))
		}
		sb.WriteString(pointStyle.Render("users")+":\n")
		for _, u := range k.Users {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", pointStyle.Render("name"), u.Name))
			sb.WriteString(fmt.Sprintf("    %s: %s\n", pointStyle.Render("mail"), u.Mail))
			sb.WriteString(fmt.Sprintf("    %s:  %s\n", pointStyle.Render("sig"), u.Sig))
		}
		return sb.String()
	}
	return "No item selected."
}
