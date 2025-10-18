package core

type Group struct {
	ID   string `db:"id"`
	Name string `db:"name"`
}

func NewGroup(name string) *Group {
	return &Group{
		Name: name,
	}
}

func (g *Group) UpdateName(newName string) *Group {
	g.Name = newName
	return g
}
