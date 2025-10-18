package adapter

import (
	"log"
	"strconv"

	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/jmoiron/sqlx"
)

type FieldsDBService struct {
	db *sqlx.DB
}

func NewFieldsDBService(db *sqlx.DB) *FieldsDBService {
	fieldService := &FieldsDBService{}
	if err := fieldService.connect(); err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	return fieldService
}

func (fDB *FieldsDBService) connect() error {
	db, err := sqlx.Connect("sqlite3", "./db/main.db")

	if err != nil {
		return err
	}
	fDB.db = db
	return nil
}

func (fDB *FieldsDBService) AddField(field core.Field) (core.Field, error) {
	result, err := fDB.db.NamedExec(`
		INSERT INTO Fields (secured_id, field_name, field_value)
		VALUES (:secured_id, :field_name, :field_value);
	`, field)

	lastID, err := result.LastInsertId()
	if err != nil {
		return core.Field{}, err
	}

	field, err = fDB.Get(strconv.Itoa(int(lastID)))
	if err != nil {
		return core.Field{}, err
	}

	return field, err
}

func (fDB *FieldsDBService) UpdateField(field core.Field) (core.Field, error) {
	_, err := fDB.db.NamedExec(`
		UPDATE Fields
		SET field_value = :field_value
		WHERE id = :id AND secured_id = :secured_id;
	`, field)
	if err != nil {
		return core.Field{}, err
	}

	field, err = fDB.Get(field.ID)

	return field, err
}

func (fDB *FieldsDBService) DeleteField(securedID, fieldID string) error {
	_, err := fDB.db.Exec(`
		DELETE FROM Fields
		WHERE id = ? AND secured_id = ?;
	`, fieldID, securedID)

	return err
}

func (fDB *FieldsDBService) GetFields(securedID string) (core.Fields, error) {
	var fields core.Fields
	err := fDB.db.Select(&fields, `
		SELECT id, secured_id, field_name, field_value
		FROM Fields
		WHERE secured_id = ?;
	`, securedID)

	return fields, err
}

func (fDB *FieldsDBService) Get(id string) (core.Field, error) {
	var field core.Field
	err := fDB.db.Get(&field, `
		SELECT id, secured_id, field_name, field_value
		FROM Fields
		WHERE id = ?;
	`, id)

	return field, err
}
