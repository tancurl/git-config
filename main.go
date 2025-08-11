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

    "github.com/muesli/termenv"
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
    list          list.Model
    choice       *KeyEntry
    userChoice   *User
    quitting      bool
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
        Foreground(lipgloss.Color("#fafafa")).
        Bold(false)

    pointStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#afafaf")).
        Bold(false)

    titleStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#1e1e1e")).
        Background(lipgloss.Color("#fafafa")).
        PaddingLeft(1).
        PaddingRight(1).
        Bold(true)

    statusMessageStyle = lipgloss.NewStyle().
        Foreground(lipgloss.Color("#1e1e1e")).
        Background(lipgloss.Color("#fafafa")).
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

    lipgloss.SetColorProfile(termenv.ANSI256)

    const defaultHeight = 20
    const defaultWidth = 40
    l := list.New(items, list.NewDefaultDelegate(), defaultWidth, defaultHeight)
    l.Title = "Git Config Manager"
    l.Styles.Title = titleStyle
    l.SetShowHelp(false)
    l.SetShowStatusBar(false)
    l.SetFilteringEnabled(true)

    m := model{list: l}
    p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr))
    finalModel, err := p.Run()
    if err != nil {
        log.Fatal(err)
    }

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

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := listStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
    case tea.KeyMsg:
        switch msg.String() {
        case "enter":
            // user selection
            if !m.selectingUser {
                if i, ok := m.list.SelectedItem().(item); ok {
                    m.choice = &i.entry
                    if len(m.choice.Users) > 1 {
                        userItems := []list.Item{}
                        for _, u := range m.choice.Users {
                            userItems = append(userItems, item{
                                title: u.Name,
                                description: u.Mail,
                                entry: *m.choice,
                            })
                        }
                        m.list.SetItems(userItems)
                        m.selectingUser = true
                        return m, nil
                    }
                    // only one user
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
        case "ctrl+q", "ctrl+c", "q", "esc":
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
