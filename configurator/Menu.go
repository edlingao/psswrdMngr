package configurator

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	passswordPorts "github.com/edlingao/psswrdMngr/internal/password/ports"
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("39"))
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

type Menu struct {
	list        list.Model
	help        help.Model
	keys        keyMap
	passwordTUI passswordPorts.PasswordTUI
	groupTUI    tea.Model
	width       int
	height      int
}

type keyMap struct {
	Up    key.Binding
	Down  key.Binding
	Enter key.Binding
	Exit  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Exit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down},    // first column
		{k.Enter, k.Exit}, // second column
	}
}

var DefaultKeymap = keyMap{
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("↑/k", "move up"),
	),

	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("↓/j", "move down"),
	),

	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),

	Exit: key.NewBinding(
		key.WithKeys("esc", "q", "ctrl+c"),
		key.WithHelp("esc/q/ctrl+c", "exit"),
	),
}

func NewMenu(
	passwordTUI passswordPorts.PasswordTUI,
	groupTUI tea.Model,
) Menu {
	items := []list.Item{
		item{title: "Groups", desc: "View and manage groups of secured items"},
		item{title: "Settings", desc: "Setup master password to encrypt your data"},
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{width: defaultWidth}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	menu := Menu{
		list:        l,
		help:        help.New(),
		keys:        DefaultKeymap,
		passwordTUI: passwordTUI,
		groupTUI:    groupTUI,
		width:       defaultWidth,
		height:      listHeight,
	}

	if setter, ok := passwordTUI.(interface{ SetParentMenu(tea.Model) }); ok {
		setter.SetParentMenu(menu)
	}

	return menu
}

type itemDelegate struct {
	width int
}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 1 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Width(d.width).
			Align(lipgloss.Center)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(d.width).
			Align(lipgloss.Center)

		title = titleStyle.Render(title)
		desc = descStyle.Render(desc)
	} else {
		titleStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(d.width).
			Align(lipgloss.Center)

		title = titleStyle.Render(title)
		desc = descStyle.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func (m Menu) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m Menu) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := msg.Width - 10
		overhead := 8
		listHeight := msg.Height - overhead
		if listHeight < 5 {
			listHeight = 5
		}

		m.list.SetSize(listWidth, listHeight)
		m.list.SetDelegate(itemDelegate{width: listWidth})
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, DefaultKeymap.Exit):
			return m, tea.Quit
		case key.Matches(msg, DefaultKeymap.Enter):
			i, ok := m.list.SelectedItem().(item)
			if ok {
				switch i.Title() {
				case "Settings":
					if setter, ok := m.passwordTUI.(interface{ SetWindowSize(int, int) }); ok {
						setter.SetWindowSize(m.width, m.height)
					}
					return m.passwordTUI, nil
				case "Groups":
					if m.groupTUI != nil {
						if setter, ok := m.groupTUI.(interface{ SetWindowSize(int, int) }); ok {
							setter.SetWindowSize(m.width, m.height)
						}
						if setter, ok := m.groupTUI.(interface{ SetParentMenu(tea.Model) }); ok {
							setter.SetParentMenu(m)
						}
						if initer, ok := m.groupTUI.(interface{ Init() tea.Cmd }); ok {
							return m.groupTUI, initer.Init()
						}
						return m.groupTUI, nil
					}
					return m, tea.Printf("Groups not available")
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Menu) View() string {
	listWidth := m.width - 10

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(listWidth).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	header := headerStyle.Render("Welcome to your password manager")
	content := header + "\n" + m.list.View()

	return boxStyle.Render(content)
}
