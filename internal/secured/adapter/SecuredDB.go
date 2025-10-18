package adapter

import (
	"log"
	"strconv"

	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/jmoiron/sqlx"
)

type SecuredDB struct {
	db *sqlx.DB
}

func NewSecuredDB(db *sqlx.DB) *SecuredDB {
	securedDB := &SecuredDB{
		db: db,
	}

	if err := securedDB.connect(); err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	return securedDB
}

func (sDB *SecuredDB) connect() error {
	db, err := sqlx.Connect("sqlite3", "./db/main.db")

	if err != nil {
		return err
	}
	sDB.db = db
	return nil
}

func (sDB *SecuredDB) Get(id string) (core.Secured, error) {
	var secured core.Secured
	err := sDB.db.Get(&secured, `
		SELECT id, title FROM secured WHERE id == ?
	`, id)
	if err != nil {
		return secured, err
	}

	return secured, nil
}

func (sDB *SecuredDB) AddSecured(secured core.Secured) (core.Secured, error) {
	result, err := sDB.db.NamedExec(`
		INSERT INTO Secured (title)
		VALUES (:title);
	`, &secured)
	if err != nil {
		return core.Secured{}, err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return core.Secured{}, err
	}

	secured, err = sDB.Get(strconv.Itoa(int(lastInsertedId)))
	if err != nil {
		return core.Secured{}, err
	}

	return secured, nil
}

func (sDB *SecuredDB) UpdateSecuredTitle(secured core.Secured) (core.Secured, error) {
	_, err := sDB.db.NamedExec(`
		UPDATE Secured
		SET
			title = :title
		WHERE 
			id = :id
	`, secured)
	if err != nil {
		return core.Secured{}, err
	}

	secured, err = sDB.Get(secured.ID)
	if err != nil {
		return core.Secured{}, err
	}

	return secured, nil
}

func (sDB *SecuredDB) DeleteSecured(id string) error {
	_, err := sDB.db.Exec(`
		DELETE FROM Secured WHERE id == ?
	`, id)
	if err != nil {
		return err
	}

	return nil
}

func (sDB *SecuredDB) GetAllSecureds() ([]core.Secured, error) {
	var secureds []core.Secured
	err := sDB.db.Select(&secureds, `
		SELECT id, title FROM secured
	`)
	if err != nil {
		return nil, err
	}

	return secureds, nil
}
