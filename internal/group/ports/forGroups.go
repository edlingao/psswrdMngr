package ports

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/edlingao/psswrdMngr/internal/group/core"
)

type GroupDBOperations interface {
	AddGroup(group core.Group) (core.Group, error)
	UpdateGroupName(group core.Group) (core.Group, error)
	DeleteGroup(groupID string) error
	GetAllGroups() ([]core.Group, error)
	Get(id string) (core.Group, error)
}

type GroupTUI interface {
	Init() tea.Cmd
	Update(tea.Msg) (tea.Model, tea.Cmd)
	View() string
}

type GroupServiceOperations interface {
	AddGroup(name string) (core.Group, error)
	UpdateGroupName(groupID, newName string) (core.Group, error)
	DeleteGroup(groupID string) error
	GetAllGroups() ([]core.Group, error)
}
