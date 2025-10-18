package adapter

import (
	"log"

	"github.com/edlingao/psswrdMngr/internal/password/core"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type PasswordDBService struct {
	db *sqlx.DB
}

func NewPasswordDBService() *PasswordDBService {
	db, err := sqlx.Connect("sqlite3", "./db/main.db")
	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	return &PasswordDBService{
		db: db,
	}
}

func (pDB *PasswordDBService) Connect() error {
	db, err := sqlx.Connect("sqlite3", "./db/main.db")

	if err != nil {
		log.Fatal("Error connecting to DB: ", err)
		return err
	}

	pDB.db = db

	return nil
}

func (pDB *PasswordDBService) Disconnect() error {
	return pDB.db.Close()
}

func (pDB *PasswordDBService) Get(verificationPassword string) (core.Password, error) {
	var password core.Password
	err := pDB.db.Get(&password, "SELECT password FROM Password WHERE id = 1")
	if err != nil {
		log.Fatal(err)
	}

	return password, nil
}

func (pDB *PasswordDBService) UpdatePassword(newPassword, oldPassword string) (core.Password, error) {
	password, err := pDB.Get(oldPassword)
	if err != nil {
		return core.Password{}, err
	}

	pswrdErr := password.Set(oldPassword, newPassword)
	if pswrdErr != nil {
		return core.Password{}, pswrdErr
	}

	_, erro := pDB.db.NamedExec(`UPDATE Password SET password = :password WHERE id = 1`, password)
	if erro != nil {
		return core.Password{}, erro
	}

	return password, nil
}
