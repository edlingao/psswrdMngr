package adapter

import (
	"log"
	"strconv"

	"github.com/edlingao/psswrdMngr/internal/group/core"
	"github.com/jmoiron/sqlx"
)

type GroupDB struct {
	db *sqlx.DB
}

func NewGroupDB(db *sqlx.DB) *GroupDB {
	groupDB := &GroupDB{
		db: db,
	}

	if err := groupDB.connect(); err != nil {
		log.Fatal("Error connecting to DB: ", err)
	}

	return groupDB
}

func (gDB *GroupDB) connect() error {
	db, err := sqlx.Connect("sqlite3", "./db/main.db")

	if err != nil {
		return err
	}
	gDB.db = db
	return nil
}

func (gDB *GroupDB) Get(id string) (core.Group, error) {
	var group core.Group
	err := gDB.db.Get(&group, `
		SELECT id, name FROM Groups WHERE id = ?
	`, id)
	if err != nil {
		return group, err
	}

	return group, nil
}

func (gDB *GroupDB) AddGroup(group core.Group) (core.Group, error) {
	result, err := gDB.db.NamedExec(`
		INSERT INTO Groups (name)
		VALUES (:name);
	`, &group)
	if err != nil {
		return core.Group{}, err
	}

	lastInsertedId, err := result.LastInsertId()
	if err != nil {
		return core.Group{}, err
	}

	group, err = gDB.Get(strconv.Itoa(int(lastInsertedId)))
	if err != nil {
		return core.Group{}, err
	}

	return group, nil
}

func (gDB *GroupDB) UpdateGroupName(group core.Group) (core.Group, error) {
	_, err := gDB.db.NamedExec(`
		UPDATE Groups
		SET name = :name
		WHERE id = :id
	`, group)
	if err != nil {
		return core.Group{}, err
	}

	group, err = gDB.Get(group.ID)
	if err != nil {
		return core.Group{}, err
	}

	return group, nil
}

func (gDB *GroupDB) DeleteGroup(id string) error {
	_, err := gDB.db.Exec(`
		DELETE FROM Groups WHERE id = ?
	`, id)
	if err != nil {
		return err
	}

	return nil
}

func (gDB *GroupDB) GetAllGroups() ([]core.Group, error) {
	var groups []core.Group
	err := gDB.db.Select(&groups, `
		SELECT id, name FROM Groups
	`)
	if err != nil {
		return nil, err
	}

	return groups, nil
}
