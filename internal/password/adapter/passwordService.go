package adapter

import (
	"errors"

	"github.com/edlingao/psswrdMngr/internal/password/core"
	"github.com/edlingao/psswrdMngr/internal/password/ports"
)

type PasswordService struct {
	DBService ports.DBOperations
	password  *core.Password
}

func NewPasswordService(
	db ports.DBOperations,
) *PasswordService {
	return &PasswordService{
		DBService: db,
	}
}

func (ps *PasswordService) Verify(verificationPassword string) (core.Password, error) {
	password, err := ps.DBService.Get(verificationPassword)
	if err != nil {
		return core.Password{}, err
	}

	if !password.Verify(verificationPassword) {
		return core.Password{}, errors.New("invalid password")
	}

	ps.password = &password

	return password, nil
}

func (ps *PasswordService) IsPasswordSet() bool {
	password, err := ps.DBService.Get("")
	if err != nil {
		return false
	}

	return password.Password != ""
}

func (ps *PasswordService) NewPassword(newPassword, oldPassword string) (core.Password, error) {
	password, err := ps.DBService.UpdatePassword(newPassword, oldPassword)
	if err != nil {
		return core.Password{}, err
	}

	ps.password = &password

	return password, nil
}

func (ps *PasswordService) Encrypt(text string) (string, error) {
	encryptedText, err := ps.password.Encrypt(text)
	if err != nil {
		return "", err
	}

	return encryptedText, nil
}

func (ps *PasswordService) Decrypt(encryptedText string) (string, error) {
	decryptedText, err := ps.password.Decrypt(encryptedText)
	if err != nil {
		return "", err
	}

	return decryptedText, nil
}
