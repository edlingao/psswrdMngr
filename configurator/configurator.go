package configurator

import (
	tea "github.com/charmbracelet/bubbletea"
	passwordAdapters "github.com/edlingao/psswrdMngr/internal/password/adapter"
	passswordPorts "github.com/edlingao/psswrdMngr/internal/password/ports"
	securedAdapters "github.com/edlingao/psswrdMngr/internal/secured/adapter"
	securedPorts "github.com/edlingao/psswrdMngr/internal/secured/ports"
)

type Configurator struct {
	PasswordService passswordPorts.PasswordOperations
	PasswordHex     passswordPorts.PasswordTUI
	SecuredService  securedPorts.SecuredServiceOperations
	SecuredTUI      *securedAdapters.SecuredTUI
	Menu            Menu
}

func NewConfig() *Configurator {
	return &Configurator{}
}

func (c *Configurator) Start() *Configurator {
	c.Menu = NewMenu(c.PasswordHex, c.SecuredTUI)
	return c
}

func (c *Configurator) AddPassword() *Configurator {
	passwordDB := passwordAdapters.NewPasswordDBService()
	passwordService := passwordAdapters.NewPasswordService(
		passwordDB,
	)
	c.PasswordService = passwordService
	c.PasswordHex = passwordAdapters.NewPasswordTUI(
		passwordService,
	)
	return c
}

func (c *Configurator) AddSecured() *Configurator {
	securedDB := securedAdapters.NewSecuredDB(nil)
	fieldsDB := securedAdapters.NewFieldsDBService(nil)
	fieldsService := securedAdapters.NewFieldsService(fieldsDB)
	fieldsService.Password = c.PasswordService
	securedService := securedAdapters.NewSecureService(securedDB, fieldsService)
	c.SecuredService = securedService
	c.SecuredTUI = securedAdapters.NewSecuredTUI(securedService, fieldsService)
	return c
}

func (c *Configurator) GetInitialModel() tea.Model {
	if c.PasswordService.IsPasswordSet() {
		loginTUI := passwordAdapters.NewLoginTUI(c.PasswordService)
		loginTUI.SetMainMenu(c.Menu)
		return loginTUI
	}
	return c.Menu
}
