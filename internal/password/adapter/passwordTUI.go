package adapter

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/edlingao/psswrdMngr/internal/password/ports"
)

type viewState int

const (
	menuView viewState = iota
	initialPasswordInputView
	oldPasswordInputView
	newPasswordInputView
	confirmationView
)

type item struct {
	title, desc string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.title }

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

type PasswordTUI struct {
	PasswordService      ports.PasswordOperations
	list                 list.Model
	initialPasswordInput textinput.Model
	oldPasswordInput     textinput.Model
	newPasswordInput     textinput.Model
	state                viewState
	width                int
	height               int
	message              string
	err                  error
	oldPasswordValue     string
	passwordExists       bool
	parentMenu           tea.Model
}

func NewPasswordTUI(
	PasswordService ports.PasswordOperations,
) *PasswordTUI {
	items := []list.Item{
		item{title: "Update Password", desc: "Set a new encryption password"},
	}

	const defaultWidth = 80
	const listHeight = 14

	l := list.New(items, itemDelegate{width: defaultWidth}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	initialPassInput := textinput.New()
	initialPassInput.Placeholder = "Enter new password"
	initialPassInput.EchoMode = textinput.EchoPassword
	initialPassInput.EchoCharacter = '•'
	initialPassInput.CharLimit = 256

	oldPassInput := textinput.New()
	oldPassInput.Placeholder = "Enter current password"
	oldPassInput.EchoMode = textinput.EchoPassword
	oldPassInput.EchoCharacter = '•'
	oldPassInput.CharLimit = 256

	newPassInput := textinput.New()
	newPassInput.Placeholder = "Enter new password"
	newPassInput.EchoMode = textinput.EchoPassword
	newPassInput.EchoCharacter = '•'
	newPassInput.CharLimit = 256

	return &PasswordTUI{
		PasswordService:      PasswordService,
		list:                 l,
		initialPasswordInput: initialPassInput,
		oldPasswordInput:     oldPassInput,
		newPasswordInput:     newPassInput,
		state:                menuView,
		width:                defaultWidth,
		height:               listHeight,
	}
}

func (m *PasswordTUI) SetParentMenu(menu tea.Model) {
	m.parentMenu = menu
}

func (m *PasswordTUI) SetWindowSize(width, height int) {
	m.width = width
	m.height = height

	listWidth := width - 10
	listHeight := 10
	if height < 20 {
		listHeight = height - 10
	}

	m.list.SetSize(listWidth, listHeight)
	m.list.SetDelegate(itemDelegate{width: listWidth})
	m.initialPasswordInput.Width = listWidth - 4
	m.oldPasswordInput.Width = listWidth - 4
	m.newPasswordInput.Width = listWidth - 4
}

func (m PasswordTUI) returnToParent() tea.Model {
	if m.parentMenu == nil {
		return m
	}

	if sizer, ok := m.parentMenu.(interface {
		Update(tea.Msg) (tea.Model, tea.Cmd)
	}); ok {
		updated, _ := sizer.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
		return updated
	}

	return m.parentMenu
}

func (m PasswordTUI) Init() tea.Cmd {
	return nil
}

func (m PasswordTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := msg.Width - 10
		listHeight := 10
		if msg.Height < 20 {
			listHeight = msg.Height - 10
		}

		m.list.SetSize(listWidth, listHeight)
		m.list.SetDelegate(itemDelegate{width: listWidth})
		m.initialPasswordInput.Width = listWidth - 4
		m.oldPasswordInput.Width = listWidth - 4
		m.newPasswordInput.Width = listWidth - 4
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case menuView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				i, ok := m.list.SelectedItem().(item)
				if ok && i.Title() == "Update Password" {
					if !m.PasswordService.IsPasswordSet() {
						m.state = initialPasswordInputView
						m.initialPasswordInput.Focus()
						m.message = ""
						m.err = nil
						return m, textinput.Blink
					}

					m.state = oldPasswordInputView
					m.oldPasswordInput.Focus()
					m.message = ""
					m.err = nil
					m.oldPasswordValue = ""
					return m, textinput.Blink
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case initialPasswordInputView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = menuView
				m.initialPasswordInput.Blur()
				m.initialPasswordInput.SetValue("")
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				password := m.initialPasswordInput.Value()
				if password == "" {
					m.err = fmt.Errorf("password cannot be empty")
					return m, nil
				}

				_, err := m.PasswordService.NewPassword(password, "")
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Password set successfully!"
				m.state = confirmationView
				m.initialPasswordInput.SetValue("")
				m.initialPasswordInput.Blur()
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.initialPasswordInput, cmd = m.initialPasswordInput.Update(msg)
			return m, cmd

		case oldPasswordInputView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = menuView
				m.oldPasswordInput.Blur()
				m.oldPasswordInput.SetValue("")
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				password := m.oldPasswordInput.Value()
				if password == "" {
					m.err = fmt.Errorf("password cannot be empty")
					return m, nil
				}

				m.oldPasswordValue = password
				m.oldPasswordInput.SetValue("")
				m.oldPasswordInput.Blur()
				m.state = newPasswordInputView
				m.newPasswordInput.Focus()
				m.err = nil
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.oldPasswordInput, cmd = m.oldPasswordInput.Update(msg)
			return m, cmd

		case newPasswordInputView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = oldPasswordInputView
				m.newPasswordInput.Blur()
				m.newPasswordInput.SetValue("")
				m.oldPasswordInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				password := m.newPasswordInput.Value()
				if password == "" {
					m.err = fmt.Errorf("password cannot be empty")
					return m, nil
				}

				_, err := m.PasswordService.NewPassword(password, m.oldPasswordValue)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Password updated successfully!"
				m.state = confirmationView
				m.newPasswordInput.SetValue("")
				m.newPasswordInput.Blur()
				m.oldPasswordValue = ""
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.newPasswordInput, cmd = m.newPasswordInput.Update(msg)
			return m, cmd

		case confirmationView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc"))):
				m.state = menuView
				m.message = ""
				m.err = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m PasswordTUI) View() string {
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

	inputStyle := lipgloss.NewStyle().
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(2)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	switch m.state {
	case menuView:
		header := headerStyle.Render("Password Settings")
		content := header + "\n" + m.list.View()
		return boxStyle.Render(content)

	case initialPasswordInputView:
		header := headerStyle.Render("Password Not Set")

		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		content := header + "\n" + infoStyle.Render("Password is not set, add a new password")
		content += "\n" + inputStyle.Render(m.initialPasswordInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: cancel")
		return boxStyle.Render(content)

	case oldPasswordInputView:
		header := headerStyle.Render("Enter Current Password")
		content := header + "\n" + inputStyle.Render(m.oldPasswordInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: continue • esc: cancel")
		return boxStyle.Render(content)

	case newPasswordInputView:
		header := headerStyle.Render("Enter New Password")
		content := header + "\n" + inputStyle.Render(m.newPasswordInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: back")
		return boxStyle.Render(content)

	case confirmationView:
		header := headerStyle.Render("Success")

		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		content := header + "\n" + messageStyle.Render(m.message)
		content += "\n" + helpStyle.Render("enter: continue")
		return boxStyle.Render(content)

	default:
		return ""
	}
}
