package core

type Field struct {
	ID        string `db:"id"`
	SecuredID string `db:"secured_id"`
	Name      string `db:"field_name"`
	Value     string `db:"field_value"`
}

type Fields []Field

func NewField(securedID, name, value string) *Field {
	return &Field{
		Name:      name,
		Value:     value,
		SecuredID: securedID,
	}
}

func (f *Field) Update(newValue string) *Field {
	f.Value = newValue
	return f
}
