package adapter

import (
	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/edlingao/psswrdMngr/internal/secured/ports"
)

type SecureService struct {
	SecureDB      ports.SecuredDBOperations
	FieldsService ports.FieldServiceOperations
}

func NewSecureService(
	securedDB ports.SecuredDBOperations,
	fieldService ports.FieldServiceOperations,
) *SecureService {
	return &SecureService{
		SecureDB:      securedDB,
		FieldsService: fieldService,
	}
}

func (svc *SecureService) AddSecured(title string, groupID *string) (core.Secured, error) {
	secured := core.NewSecured(title, groupID)
	return svc.SecureDB.AddSecured(*secured)
}

func (svc *SecureService) UpdateSecuredTitle(securedID, newTitle string) (core.Secured, error) {
	secured, err := svc.SecureDB.Get(securedID)
	if err != nil {
		return core.Secured{}, err
	}

	secured.UpdateTitle(newTitle)
	return svc.SecureDB.UpdateSecuredTitle(secured)
}

func (svc *SecureService) DeleteSecured(securedID string) error {
	return svc.SecureDB.DeleteSecured(securedID)
}

func (svc *SecureService) GetAllSecureds() ([]core.Secured, error) {
	return svc.SecureDB.GetAllSecureds()
}

func (svc *SecureService) GetSecuredsByGroup(groupID string) ([]core.Secured, error) {
	return svc.SecureDB.GetSecuredsByGroup(groupID)
}

func (svc *SecureService) GetUngroupedSecureds() ([]core.Secured, error) {
	return svc.SecureDB.GetUngroupedSecureds()
}

func (svc *SecureService) AddFieldToSecured(securedID, fieldName, fieldValue string) (core.Field, error) {
	return svc.FieldsService.AddField(securedID, fieldName, fieldValue)
}
