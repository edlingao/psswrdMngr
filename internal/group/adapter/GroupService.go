package adapter

import (
	"github.com/edlingao/psswrdMngr/internal/group/core"
	"github.com/edlingao/psswrdMngr/internal/group/ports"
)

type GroupService struct {
	GroupDB ports.GroupDBOperations
}

func NewGroupService(
	groupDB ports.GroupDBOperations,
) *GroupService {
	return &GroupService{
		GroupDB: groupDB,
	}
}

func (svc *GroupService) AddGroup(name string) (core.Group, error) {
	group := core.NewGroup(name)
	return svc.GroupDB.AddGroup(*group)
}

func (svc *GroupService) UpdateGroupName(groupID, newName string) (core.Group, error) {
	group, err := svc.GroupDB.Get(groupID)
	if err != nil {
		return core.Group{}, err
	}

	group.UpdateName(newName)
	return svc.GroupDB.UpdateGroupName(group)
}

func (svc *GroupService) DeleteGroup(groupID string) error {
	return svc.GroupDB.DeleteGroup(groupID)
}

func (svc *GroupService) GetAllGroups() ([]core.Group, error) {
	return svc.GroupDB.GetAllGroups()
}
