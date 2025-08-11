package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
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
	userChoice *User
	quitting bool
	selectingUser bool
}

var (
	listStyle = lipgloss.NewStyle().
		Width(40).
		MarginRight(2).
		Padding(1)

	previewStyle = lipgloss.NewStyle().
		Width(60).
		Padding(1).
		MarginLeft(1).
		Foreground(lipgloss.Color("#FAFAFA")).   // light gray text
		Bold(false)

	pointStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#aaaaaa")).   // light gray text
		Bold(false)

	titleStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.CompleteColor{TrueColor: "#005577", ANSI256: "4", ANSI: "4"}).
		PaddingLeft(1).
		PaddingRight(1).
		Bold(true)

	statusMessageStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.CompleteColor{TrueColor: "#770000", ANSI256: "1", ANSI: "1"}).
		PaddingLeft(1).
		PaddingRight(1).
		MarginBottom(1).
		Bold(true)
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
	l.Styles.Title = titleStyle
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	m := model{list: l}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		log.Fatal(err)
	}

	// After TUI exits, print the selected item
	if fm, ok := finalModel.(model); ok && fm.choice != nil && fm.userChoice != nil {
		fmt.Printf("key:%s,", fm.choice.Key)
		fmt.Printf("algo:%s,", fm.choice.Algo)
		fmt.Printf("created:%s,", fm.choice.Create)
		if fm.choice.Expire == "" {
			fmt.Printf("expire:%s,", "never")
		} else {
			fmt.Printf("expire:%s,", fm.choice.Expire)
		}
		fmt.Printf("user:%s,email:%s,sig:%s",
		fm.userChoice.Name, fm.userChoice.Mail, fm.userChoice.Sig)
	}


}

func (m model) Init() tea.Cmd {
	return nil
}

type listKeyMap struct {
	toggleSpinner    key.Binding
	toggleTitleBar   key.Binding
	toggleStatusBar  key.Binding
	togglePagination key.Binding
	toggleHelpMenu   key.Binding
	insertItem       key.Binding
}
func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		insertItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add item"),
		),
		toggleSpinner: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "toggle spinner"),
		),
		toggleTitleBar: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("T", "toggle title"),
		),
		toggleStatusBar: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("S", "toggle status"),
		),
		togglePagination: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle pagination"),
		),
		toggleHelpMenu: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "toggle help"),
		),
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if !m.selectingUser {
				if i, ok := m.list.SelectedItem().(item); ok {
					m.choice = &i.entry
					if len(m.choice.Users) > 1 {
						// Switch to user selection mode
						userItems := []list.Item{}
						for _, u := range m.choice.Users {
							userItems = append(userItems, item{
								title: u.Name,
								description: u.Mail,
								entry: *m.choice, // keep full entry for final output
							})
						}
						m.list.SetItems(userItems)
						// m.list.Title = "Select User"
						m.selectingUser = true
						return m, nil
					}
					// Only one user
					m.userChoice = &m.choice.Users[0]
					return m, tea.Quit
				}
			} else {
				if i, ok := m.list.SelectedItem().(item); ok {
					for _, u := range m.choice.Users {
						if u.Name == i.title && u.Mail == i.description {
							m.userChoice = &u
							break
						}
					}
					return m, tea.Quit
				}
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
	if m.choice != nil && (!m.selectingUser || m.userChoice != nil) {
		return ""
	}

	left := listStyle.Render(m.list.View())
	right := previewStyle.Render(m.previewView())

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m model) previewView() string {
	if m.selectingUser {
		if i, ok := m.list.SelectedItem().(item); ok {
			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("%s\n", statusMessageStyle.Render("Select User")))
			sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("key"), i.entry.Key))
			sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("name"), i.title))
			sb.WriteString(fmt.Sprintf("%s: %s\n", pointStyle.Render("mail"), i.description))
			return sb.String()
		}
		return "No user selected."
	}
	if i, ok := m.list.SelectedItem().(item); ok {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("%s\n", statusMessageStyle.Render("Preview")))
		k := i.entry
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
