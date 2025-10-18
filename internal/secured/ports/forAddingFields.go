package ports

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/internal/secured/core"
)

type FieldDBOperations interface {
	AddField(field core.Field) (core.Field, error)
	UpdateField(field core.Field) (core.Field, error)
	DeleteField(securedID, fieldID string) error
	GetFields(securedID string) (core.Fields, error)
	Get(id string) (core.Field, error)
}

type FieldTUI interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type FieldServiceOperations interface {
	AddField(securedID, name, value string) (core.Field, error)
	UpdateField(securedID, fieldID, newValue string) (core.Field, error)
	DeleteField(securedID, fieldID string) error
	GetFields(securedID string) (core.Fields, error)
}
