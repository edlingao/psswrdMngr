package adapter

import (
	passwordPorts "github.com/edlingao/psswrdMngr/internal/password/ports"
	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/edlingao/psswrdMngr/internal/secured/ports"
)

type FieldsService struct {
	FieldDB  ports.FieldDBOperations
	Password passwordPorts.PasswordOperations
}

func NewFieldsService(
	fieldDB ports.FieldDBOperations,
) *FieldsService {
	return &FieldsService{
		FieldDB: fieldDB,
	}
}

func (fs *FieldsService) AddField(securedID, name, value string) (core.Field, error) {
	encryptedValue, err := fs.Password.Encrypt(value)
	if err != nil {
		return core.Field{}, err
	}

	field, err := fs.FieldDB.AddField(core.Field{
		SecuredID: securedID,
		Name:      name,
		Value:     encryptedValue,
	})

	if err != nil {
		return core.Field{}, err
	}

	return field, nil
}

func (fs *FieldsService) UpdateField(securedID, fieldID, newValue string) (core.Field, error) {
	field, err := fs.FieldDB.Get(fieldID)
	if err != nil {
		return core.Field{}, err
	}

	encryptedValue, err := fs.Password.Encrypt(newValue)
	if err != nil {
		return core.Field{}, err
	}

	field.Update(encryptedValue)
	field, err = fs.FieldDB.UpdateField(field)

	if err != nil {
		return core.Field{}, err
	}

	decryptedValue, err := fs.Password.Decrypt(field.Value)
	if err != nil {
		return core.Field{}, err
	}
	field.Update(decryptedValue)

	return field, nil
}

func (fs *FieldsService) DeleteField(securedID, fieldID string) error {
	return fs.FieldDB.DeleteField(securedID, fieldID)
}

func (fs *FieldsService) GetFields(securedID string) (core.Fields, error) {
	fields, err := fs.FieldDB.GetFields(securedID)
	if err != nil {
		return core.Fields{}, err
	}

	decryptedFields := make(core.Fields, 0, len(fields))
	for _, field := range fields {
		decryptedValue, err := fs.Password.Decrypt(field.Value)
		if err != nil {
			return core.Fields{}, err
		}
		field.Update(decryptedValue)
		decryptedFields = append(decryptedFields, field)
	}

	return decryptedFields, nil
}
