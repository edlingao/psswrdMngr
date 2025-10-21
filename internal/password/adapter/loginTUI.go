package adapter

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/edlingao/psswrdMngr/internal/password/ports"
)

type LoginTUI struct {
	PasswordService ports.PasswordOperations
	passwordInput   textinput.Model
	mainMenu        tea.Model
	width           int
	height          int
	err             error
}

func NewLoginTUI(
	passwordService ports.PasswordOperations,
) *LoginTUI {
	passInput := textinput.New()
	passInput.Placeholder = "Enter your password"
	passInput.EchoMode = textinput.EchoPassword
	passInput.EchoCharacter = '•'
	passInput.CharLimit = 256
	passInput.Focus()

	return &LoginTUI{
		PasswordService: passwordService,
		passwordInput:   passInput,
		width:           80,
		height:          14,
	}
}

func (m *LoginTUI) SetMainMenu(menu tea.Model) {
	m.mainMenu = menu
}

func (m *LoginTUI) SetWindowSize(width, height int) {
	m.width = width
	m.height = height
	m.passwordInput.Width = width - 14
}

func (m LoginTUI) Init() tea.Cmd {
	return tea.Batch(tea.ClearScreen, textinput.Blink)
}

func (m LoginTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.passwordInput.Width = msg.Width - 14
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			password := m.passwordInput.Value()
			if password == "" {
				m.err = fmt.Errorf("password cannot be empty")
				return m, nil
			}

			_, err := m.PasswordService.Verify(password)
			if err != nil {
				m.err = fmt.Errorf("incorrect password")
				m.passwordInput.SetValue("")
				return m, nil
			}

			if m.mainMenu != nil {
				if sizer, ok := m.mainMenu.(interface {
					Update(tea.Msg) (tea.Model, tea.Cmd)
				}); ok {
					updated, _ := sizer.Update(tea.WindowSizeMsg{
						Width:  m.width,
						Height: m.height,
					})
					return updated, nil
				}
				return m.mainMenu, nil
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.passwordInput, cmd = m.passwordInput.Update(msg)
	return m, cmd
}

func (m LoginTUI) View() string {
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
		Align(lipgloss.Left).
		MarginTop(2)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	header := headerStyle.Render("Welcome Back")
	content := header + "\n" + inputStyle.Render(m.passwordInput.View())

	if m.err != nil {
		content += "\n" + errorStyle.Render(m.err.Error())
	}

	content += "\n" + helpStyle.Render("enter: login • ctrl+c: exit")

	return boxStyle.Render(content)
}
