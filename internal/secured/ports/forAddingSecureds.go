package ports

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/internal/secured/core"
)

type SecuredDBOperations interface {
	AddSecured(secured core.Secured) (core.Secured, error)
	UpdateSecuredTitle(secured core.Secured) (core.Secured, error)
	DeleteSecured(securedID string) error
	GetAllSecureds() ([]core.Secured, error)
	GetSecuredsByGroup(groupID string) ([]core.Secured, error)
	GetUngroupedSecureds() ([]core.Secured, error)
	Get(id string) (core.Secured, error)
}

type SecuredTUI interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type SecuredServiceOperations interface {
	AddSecured(title string, groupID *string) (core.Secured, error)
	UpdateSecuredTitle(securedID, newTitle string) (core.Secured, error)
	DeleteSecured(securedID string) error
	GetAllSecureds() ([]core.Secured, error)
	GetSecuredsByGroup(groupID string) ([]core.Secured, error)
	GetUngroupedSecureds() ([]core.Secured, error)
	AddFieldToSecured(securedID, fieldName, fieldValue string) (core.Field, error)
}
