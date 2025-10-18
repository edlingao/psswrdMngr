package ports

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/internal/password/core"
)

type PasswordOperations interface {
	Verify(password string) (core.Password, error)
	NewPassword(newPassword, oldPassword string) (core.Password, error)
	Encrypt(text string) (string, error)
	Decrypt(encryptedText string) (string, error)
	IsPasswordSet() bool
}

type DBOperations interface {
	Get(password string) (core.Password, error)
	UpdatePassword(newPassword, oldPassword string) (core.Password, error)
	Connect() error
	Disconnect() error
}

type PasswordTUI interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}
